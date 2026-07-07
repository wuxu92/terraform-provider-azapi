# PRD: Shared envelope renderer for the acceptance config builders

Status: ready-for-agent

_Source: architecture review of `internal/native` (2026-07-06), candidate 6 — "Extract a shared envelope renderer for the `_config.go` builders"._

## Problem Statement

Every static resource ships a config builder in its generated service package
(`*Cfg` types in the five `_config.go` files) that renders the HCL an acceptance
scenario applies. Each builder hand-spells the same **operational envelope** shell
around its per-resource property body:

```
resource <tfType> <label> {
  name              = <name expr>
  <parent attr>     = <parent ref>
  location          = "{{.Location}}"   # most, not all
  kind              = "<kind>"          # some
  ...per-resource body...
}
```

The property bodies are irreducibly per-resource (they encode real Azure
semantics and round-trip constraints — roughly 80% of each file and correctly so).
But the envelope shell — the `resource %q %q { name = … / <parent>_id = … /
location = … }` scaffolding — is copied verbatim across all five builders, with
only small, mechanical differences:

- the **parent-reference attribute** name varies (`subscription_id`,
  `resource_group_id`, `storage_account_id`);
- the **parent-reference value** varies (a literal subscription path for the
  resource group vs. a dependency's `IDRef()` for the rest);
- `location` is present for every resource **except** the blob service (a
  singleton default child that has none);
- `kind` is present for some (`StorageV2`, `app`) and absent for others.

Because the shell is duplicated, a change to the envelope shape (a new common
attribute, a formatting fix, a naming convention) means editing five files in
lockstep, and each builder is a place the shell can silently diverge. The
structural boilerplate is shallow-repetitive even though the bodies are not.

## Solution

Add one shared **envelope renderer** beside `ResourceConfigBase` in the `services`
package. It takes the envelope-varying pieces (the resource's type/label — already
on `ResourceConfigBase` — plus the name expression, the parent attribute and its
reference, optional `location`, optional `kind`, and the per-resource body
fragment) and returns the `resource … { … }` shell. Each builder keeps its own
scenario types (`_Basic`, `_Complete`, `_Complete_update`, `_Named`, …) and its
property-body helper methods **verbatim**; it stops hand-spelling the shell and
instead hands its body fragment to the renderer.

The rendered output stays byte-for-byte identical to today's, because these
strings are live acceptance configs — the goal is one home for the envelope, not a
new config. One place renders the shell; the per-resource bodies remain exactly as
authored.

## User Stories

1. As a maintainer, I want the `resource … { name / parent / location }` envelope
   rendered in one place, so that a change to its shape is one edit, not five.
2. As a maintainer, I want each config builder to supply only what actually varies
   per resource (name, parent attribute + reference, optional location/kind, body),
   so that the shared shell can't silently diverge between builders.
3. As a maintainer, I want the renderer to omit `location` when a resource has none
   (the singleton blob service), so that the one legitimate exception is expressed
   as data, not a separately hand-written shell.
4. As a maintainer, I want the renderer to omit `kind` when a resource doesn't set
   one, so that resource-group-style envelopes and kind-bearing envelopes share the
   same renderer.
5. As a maintainer, I want the parent-reference attribute name and value to be
   inputs, so that `subscription_id` (literal path), `resource_group_id` and
   `storage_account_id` (dependency `IDRef`) all route through the same shell.
6. As a maintainer, I want each builder's `Config()` output to stay byte-identical
   after the refactor, so that the live acceptance scenarios apply exactly the same
   HCL and no Azure behavior changes.
7. As a maintainer, I want the per-resource property bodies left verbatim, so that
   the irreducible, semantics-bearing part of each builder is untouched and easy to
   diff.
8. As a maintainer, I want the renderer unit-tested once, so that envelope
   correctness is proven in a fast in-process test instead of only via a live-Azure
   apply.
9. As a resource author adding a new static resource, I want to render its config
   envelope by supplying a small spec, so that I don't copy another resource's
   shell and inherit its quirks.
10. As a resource author, I want the renderer to preserve the acceptance
    framework's template tokens (`{{.RandomString}}`, `{{.Location}}`,
    `{{.SubscriptionID}}`) untouched in the output, so that the framework's
    renderer still fills them.
11. As a maintainer reviewing a config change, I want to see at a glance whether a
    diff touches the shared shell or a per-resource body, so that envelope changes
    and body changes are reviewed distinctly.
12. As a maintainer, I want the renderer to live in `services` beside
    `ResourceConfigBase` (not in `nativeacc`), so that no import cycle is
    introduced (the generated service packages already embed `ResourceConfigBase`).
13. As a maintainer, I want the `sku` and other structural sub-blocks to remain
    part of each resource's body fragment, so that the renderer stays minimal and
    only owns the genuinely shared envelope lines.

## Implementation Decisions

- **Module touched.** `internal/native/services` — a new renderer beside
  `ResourceConfigBase` in the shared `services` package, plus edits to the five
  builders (`resources/resource_group_config.go`,
  `storage/storage_account_config.go`,
  `storage/storage_account_blob_service_config.go`,
  `web/web_server_farm_config.go`, `web/web_site_config.go`) to call it. No change
  to `nativeacc` (the acceptance framework) or to any generated `_gen.go`.
- **Renderer placement and cycle-safety.** The renderer lives in `services`
  alongside `ResourceConfigBase`, for the same reason `ResourceConfigBase` does:
  the generated service packages import `services`, and `nativeacc` transitively
  imports them, so putting shared config helpers in `services` (not `nativeacc`)
  avoids an import cycle. It may be a method on `ResourceConfigBase` (it already
  carries `ResourceType()`/`ResourceLabel()`) or a package function taking the
  base — the former reads most naturally since type+label come for free.
- **The envelope spec is the set of varying inputs.** Derived from the five
  existing builders, the shell varies along exactly these axes, which become the
  renderer's inputs (type/label supplied by the embedded base):

  ```
  name        string // name expression, e.g. `"acctestsa{{.RandomString}}"` or an explicit name
  parentAttr  string // "subscription_id" | "resource_group_id" | "storage_account_id"
  parentRef   string // literal subscription path, or a dependency's IDRef()
  location    bool   // emit the location line (false only for the singleton blob service)
  kind        string // "" = omit; else the kind value ("StorageV2", "app", …)
  body        string // per-resource property fragment, appended verbatim
  ```

  (Shape derived from analysis of the five current builders, not a prototype.)
- **`sku` and other sub-blocks fold into `body`.** The storage-account and
  web-server-farm `sku { … }` blocks are per-resource structural content, not part
  of the universal envelope, so they stay in each builder's body fragment. The
  renderer owns only the always-or-conditionally-present envelope lines (`name`,
  parent attr, optional `location`, optional `kind`). This keeps the shared surface
  minimal and honest.
- **Output is byte-identical.** The refactor is behavior-preserving: for every
  existing scenario, the renderer must produce the exact same HCL string the
  builder produces today (same field order, same indentation, same template
  tokens). The builders' public scenario types and `Config()` signatures are
  unchanged; only the private `config(...)` helper's body is replaced by a call to
  the shared renderer.
- **Template tokens pass through untouched.** The renderer does no template
  expansion — `{{.RandomString}}`, `{{.Location}}`, `{{.SubscriptionID}}` and any
  `%s`-injected dependency `IDRef()` are emitted verbatim for the acceptance
  framework's renderer to fill, exactly as now.

## Testing Decisions

- **What makes a good test here.** It asserts the renderer's **external output** —
  the HCL shell string for a given envelope spec — not how it's assembled
  internally. The load-bearing property is byte-identity with the current
  builders, so the highest-value assertions are: (a) the renderer emits the exact
  shell for representative specs covering each axis (with/without `location`,
  with/without `kind`, each parent-attribute form, body appended vs. empty); and
  (b) each existing builder's canonical scenario (`_Basic`) `Config()` output is
  unchanged by the refactor.
- **Seam (approved).** The renderer is a single pure string function that all five
  builders route through — the highest, narrowest seam available (one function, N
  callers). It is directly unit-testable in-process with no Terraform and no Azure.
  This is a new seam, introduced at the highest point per the guidance.
- **Prior art to follow.** String-output assertions already exist in the tree:
  `internal/native/generator/customizers/customizers_emit_test.go` asserts emitted
  HCL/Go substrings, and `internal/native/acceptance/scope_test.go`
  (`TestStringConfigure`) asserts an exact config string. The renderer test is
  table-driven in that style, in the `services` package (white-box or same-package
  test) since the renderer is package-scoped.
- **Modules tested.** `internal/native/services` (the renderer) and, via the
  builder-output identity check, each generated service package's `_Basic`
  scenario.
- **Coverage target.** One table case per envelope axis in the renderer test, plus
  an identity assertion for each of the five builders' `_Basic` `Config()`. Test
  authoring is delegated to the Tester agent per repo convention. The existing
  live-Azure acceptance suites remain the end-to-end guarantee that the configs
  still apply.

## Out of Scope

- Any change to the per-resource **property bodies** or scenario semantics
  (`_Complete`, `_Complete_update`, `_Named`, `_SKU`, `_ChangeFeed`, …). Those are
  irreducible and stay verbatim.
- Any change to the **rendered HCL** for existing scenarios. This is a pure
  extraction; a different output would change live acceptance behavior and is
  explicitly disallowed.
- The acceptance framework's template rendering (`{{.RandomString}}` etc.) in
  `nativeacc` — untouched; the renderer only emits tokens, never expands them.
- Generating config builders from the type graph, or otherwise mechanizing the
  property bodies. That is a much larger effort; this PRD only collapses the shared
  envelope shell.
- The `DataSourceConfigBase` path — no data-source config builders exhibit the
  duplication; leave it alone unless a future data source needs the same shell.

## Further Notes

- This is honest deduplication: it collapses only the ~20% of each builder that is
  genuinely shared boilerplate and deliberately leaves the ~80% that is real,
  per-resource Azure configuration. The renderer is worth extracting precisely
  because the envelope shape is uniform while the bodies are not.
- The byte-identity constraint is what keeps this safe to ship without a live run:
  if the unit test proves the renderer reproduces each `_Basic` config exactly and
  the build is green, the acceptance configs are unchanged by construction.
- No ADR is reopened. The change is confined to test-support config builders and
  touches neither the generation pipeline (ADR-0007) nor the schema/runtime split
  (ADR-0002).
