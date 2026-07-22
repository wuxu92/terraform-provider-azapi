# Azapin developer authoring pain points — analysis & effort report

Scope: the **developer-facing friction of adding or updating a native resource**, not
end-user capability gaps (those live in `TODO.md`). Grounded in the four-layer
generation model (`DEVELOPER_GUIDE.md`), the three workflows (add / regenerate /
version-upgrade), and a census of the actual hand-written surfaces in the tree.
Solve-effort figures are `[INFERENCE]`; the measured surface sizes are not.

## Method & baseline (measured 2026-07-17, 16 shipped resources)

| Surface | Total LOC | Resources touched | ~LOC / resource | Tooling status |
|---|---:|---:|---:|---|
| `*_gen.go` (generated) | 18,864 | 16/16 | ~1,180 | **free** (regenerated) |
| azwise overlay (`github.com/wuxu92/azwise/*.go`) | 6,016 | 16/16 | ~376 | **semi-tooled** — `skill://azwise` drafts, `azwise_validate` checks |
| acceptance `*_test.go` (Ginkgo wiring) | 2,390 | 16/16 | ~149 | **untooled** |
| acceptance `*_config.go` (builders) | 2,189 | 16/16 | ~137 | **tooled** — `skill://azapi-acceptance-author` (new) |
| customizers | 1,615 | 15/16 | ~108 | **semi-tooled** — fluent helpers on `ResourceDefinition` (new) |
| runtime hooks (`*_hooks.go`) | 747 | 8/16 | ~93 | **untooled** |
| service-specific validators | 262 | 2/16 | — | fluent-helper adjacent |
| generic validators | 161 | shared | — | reuse-first, low churn |

**Headline:** ~770 LOC of hand-written surface per resource (plus ~93 when hooks are
needed). Codegen is free; the cost is the four curated layers and their wiring. Two
already got tooling this cycle (config builder, customizer helpers); azwise is
semi-tooled. The **untooled, still-expensive** surfaces are hooks, test wiring,
multi-site registration, and the version-upgrade reconciliation loop — plus two
cross-cutting *failure-mode* costs (silent azwise drops, schema-fidelity guesswork)
that consume debugging time rather than authoring time.

---

## Pain points, ranked by leverage (impact × frequency ÷ solve-effort)

Each: what it costs today, why, solve complexity/effort `[INFERENCE]`, and dependencies.
"Solve effort" is the cost to *build the tooling*, not to do the task once.

### P1 — Multi-site manual registration & wiring (mechanical, forget-one hazard)
**Cost today:** Adding one resource edits **3–4 redundant sites**, all keyed by the same
ARM type / Terraform name:
`armtype.go` const → azwise `services/<service>/<resource>.go` `init(){ Register(New…()) }` → `main.go` target
→ `customizers/register.go` (when a customizer exists) → `services/all/all.go` blank
import (only for a *new* service). No single source of truth; the same identifier is
re-spelled 3–5×.
**Why it hurts:** Pure boilerplate, but **failure is silent-ish**: add the azwise file
but forget the generator target → no regeneration, schema ships uncurated; add the
target but forget the azwise `Register` → curation silently absent; forget the
`services/all` import for a new service → resource compiles but never registers into the
provider. Each omission is a debugging session, not a compile error.
**Impact:** Low LOC, but hits **every** add (16/16) and every reviewer.
**Solve:** *Complexity Low–Med.* A generator-adjacent scaffold command
(`azapin new <ARM-type>`) that: appends the `armtype` const, stubs the azwise file +
`Register` line, adds the target, and (new service) the blank import — from **one** ARM
type argument. Or invert it: make the generator discover targets from the azwise
registry / a manifest so the target list stops being a second source of truth.
**Effort:** ~3–5 days for the scaffold; ~1–2 days for the manifest-driven target list.
**Leverage: HIGH** (cheap, every-resource, removes a silent-failure class).

### P2 — azwise silent-drop on rename / version-upgrade (correctness landmine)
**Cost today:** `ApplyAzwise` **skips an unresolved path silently** (`DEVELOPER_GUIDE.md`
"azwise paths fail silently"). When a body property is renamed/moved — routine on an
API-version upgrade — every azwise rule on that path (ForceNew, validator, default) is
**dropped with no error**, and the schema regenerates green. Only the
`TestApplyAzwise*` overlay tests (which parse live `types.json`) catch it, and only if
the developer runs and reads them.
**Why it hurts:** The failure is invisible at generation and at `azapin-validate` (which
checks *presence/parity*, not that a curated rule survived). A dropped ForceNew ships a
resource that silently does in-place updates ARM rejects at apply — discovered in
production, not review.
**Impact:** Medium frequency (every version-upgrade, every azwise-path edit), **high
severity**.
**Solve:** *Complexity Low–Med.* `azwise_validate` already resolves paths against SDK
structs — extend it (or the generator) to **fail loud** when a registered azwise path
doesn't resolve in the *generated* graph, promoting the silent skip to a gate error
behind an allowlist for intentionally-catch-all rules. Ties into P6.
**Effort:** ~3–5 days to wire the resolve-or-fail gate into the pipeline + escape hatch.
**Leverage: HIGH** (converts a production-class landmine into a build error).

### P3 — Runtime hook authoring (bespoke Go, schema-vs-behavior classification)
**Cost today:** 8/16 resources hand-write `*_hooks.go` (~93 LOC each): `ModifyPlan`
conditional ForceNew, `AfterRead` response massaging, `AfterDelete` orchestration,
`Singleton` handling. The `CrudCtx` field-liveness matrix (which fields are live in
which hook) and the **"schema concern vs behavior concern"** rule are genuine cognitive
load — putting a validator in a hook, or reading ARM state in a customizer, is a
classic mistake the guide warns against explicitly.
**Why it hurts:** Not templatable in general (the logic *is* the resource), but the
**recurring shapes** (azwise-`CheckForceNew` ModifyPlan, empty-list normalization,
reserved-value read filter + schema cap pairing) are copy-adapted by hand each time,
and the field-liveness rules are re-derived from the doc on every write.
**Impact:** Medium frequency (50% of resources), medium per-instance cost.
**Solve:** *Complexity Med.* A small library of **parameterized hook builders** for the
proven patterns (e.g. `hooks.SkuForceNew(descriptor, "sku.name")`,
`hooks.NormalizeEmptyList(path)`, `hooks.FilterReadValue(path, sentinel)`), so a
resource wires a one-liner instead of re-implementing the CrudCtx dance. The truly
bespoke hooks stay hand-written.
**Effort:** ~1–1.5 weeks to extract the 3–4 dominant patterns + tests; ongoing as new
patterns recur.
**Leverage: MED-HIGH** (halves the hook surface where it applies; residual stays manual).

### P4 — Acceptance test wiring beyond the config builder
**Cost today:** The config builder is now drafted by an agent, but the **`_test.go`
Ginkgo container** — `Workspace`/`Scope`/`ResourceFor`, `Ordered` `BeforeAll` chains,
dependency-scope nesting, `Check`/`ImportVerify` calls — is hand-written (~149
LOC/resource, 16/16). And the **native-dependency wall** (`DEVELOPER_GUIDE.md` hard
stop) blocks a resource's tests entirely until every prerequisite ARM type is *also* a
native resource, so authoring order is constrained and a missing dep stalls the whole
resource.
**Why it hurts:** The wiring is semi-mechanical (given the scenarios), but the scope/
dependency graph and the `Check` assertions for computed-default facts need judgment.
The native-dep wall turns one resource's test into a dependency-tree project.
**Impact:** Every resource (16/16), but partially addressed — the config half is tooled.
**Solve:** *Complexity Med.* Extend `skill://azapi-acceptance-author` (or a sibling) to
also draft the `_test.go` Ordered skeleton from the config scenarios + the descriptor's
parent chain, leaving `Check` assertions and dependency provisioning for the human.
Separately, a **native-dependency preflight** that, given a target, reports the
transitive ARM types missing from `services.Registry` — so the wall is known at intake,
not mid-authoring.
**Effort:** ~1 week for the test-skeleton drafter; ~2–3 days for the dependency
preflight. Overlaps `TODO.md` #1 (framework polish).
**Leverage: MED** (config half already won; test-wiring half + preflight remain).

### P5 — API-version-upgrade reconciliation (manual checklist, no diff tooling)
**Cost today:** Workflow 3 is a **7-step manual checklist**: vendor bicep types,
confirm latest-stable advanced, regenerate, reconcile azwise (silent drops — see P2),
reconcile customizers (loud panics), `azapin-validate`, update composition + acceptance
tests. No tool diffs the old vs new schema, lists moved/removed/renamed properties, or
tells you which azwise paths need updating — the developer greps and eyeballs.
**Why it hurts:** Combines P2's silent-drop risk with pure manual diffing. The loud
customizer panic is actually the *friendly* half; the azwise + parity reconciliation is
where time goes.
**Impact:** Low frequency per resource (only when upstream publishes), but high
per-event cost and rising as coverage grows.
**Solve:** *Complexity Med-High.* A **schema-diff tool** (old model vs regenerated
model: added/removed/renamed/flag-changed/type-changed properties) driving an upgrade
report that flags exactly which azwise/customizer paths moved. This is `TODO.md` #6
(breaking-change detector) + #8 (upgrade tools); the authoring-friction framing is the
same tool.
**Effort:** ~2–3 weeks for the detector; upgrade-report layer ~1 week on top. Shared
with the deprecation/state-migration work (`TODO.md` #11).
**Leverage: MED** (high per-event, lower frequency; big shared-infrastructure payoff).

### P6 — Schema-fidelity guesswork (lossy sources, no pre-acceptance verification)
**Cost today:** bicep flags + description-mining + azwise are **lossy vs real ARM**
(ranges, enum completeness, `MaxItems`, immutability, real defaults,
write-accepted-but-not-honored fields — `DEVELOPER_GUIDE.md` calls these out
explicitly). The developer *guesses* during azwise/customizer authoring and only learns
they were wrong at **live acceptance** (minutes/apply, human-gated). Description-mined
defaults/validators can be silently wrong.
**Why it hurts:** It pushes the correctness feedback to the slowest, most expensive loop
(live apply), and the same class of unknowns recurs on every resource. It's the root
cause under several other pains (P2's dropped rules, P4's acceptance iterations).
**Impact:** Every resource, diffuse — hard to attribute but large in aggregate.
**Solve:** *Complexity High.* The **live-API schema verification harness** (`TODO.md`
#16, `SCHEMA_VERIFICATION.md`): probe boundary values, diff PUT→GET, classify
Confirmed/Refuted/Coerced/Inconclusive into suggested azwise/customizer edits. Turns
guesswork into a mechanical verify-and-suggest loop.
**Effort:** ~1 week REST PoC on one resource; 2–4 weeks planner+classifier+reporter;
CI cadence ongoing. Largest build here.
**Leverage: MED** (highest ceiling, highest cost; foundational for trustworthy overlays).

### P7 — azwise overlay authoring depth (largest surface, already semi-tooled)
**Cost today:** ~376 LOC/resource, the single biggest hand-written surface, requiring
deep AzureRM cross-referencing: field-type discipline (`StringRules` vs `IntRules`
against SDK structs), sub-service separation (`blob_properties` → the *sub-service* file,
not the parent), path mapping, and "transfer every validator, never silently drop one."
**Why it hurts less than its size suggests:** `skill://azwise` already drafts it and
`azwise_validate` catches type/path/enum mistakes — so the raw surface is large but the
*marginal* authoring cost is already cut. Residual pain is the judgment (which validators
are semantic → customizer vs declarative → azwise) and sub-service boundary calls.
**Impact:** Every resource, but mitigated.
**Solve:** *Complexity Low (incremental).* Sharpen the existing agent: auto-classify
declarative-vs-semantic validators, flag sub-service-owned fields, and lean harder on
`azwise_validate` in the loop. Not a new tool — polish.
**Effort:** ~2–4 days of skill/agent refinement, ongoing.
**Leverage: MED** (big surface, but diminishing returns — the agent already exists).

### P8 — Discriminated-variant / Dynamic-fallback authoring
**Cost today:** When a resource root is a `DiscriminatedObjectType`, the config +
overlay authoring is materially harder (per-variant blocks, `ExactlyOneOf`, mostly-
ForceNew variant properties — as seen drafting `deployment_script`). 632 resources fall
back to `azapi_resource` entirely today.
**Why it hurts:** Overlaps a generator *capability* gap (`TODO.md` #11-equivalent /
tech-summary feature 12), but the *authoring* slice is real: variant resources need more
hand-holding across every layer.
**Impact:** Currently rare in shipped set (few discriminated resources), growing as the
variant-block work lands.
**Solve:** *Complexity High*, but it's primarily the discriminator generator project,
not a standalone authoring tool. Track it there.
**Effort:** part of the discriminator epic (weeks). **Leverage: LOW now** (few
resources), re-rank when variant coverage expands.

---

## Recommended order (by leverage)

1. **P1 — scaffold/manifest wiring** (Low effort, every-resource, kills a silent-failure class).
2. **P2 — azwise resolve-or-fail gate** (Low-Med effort, removes a production landmine; reuses `azwise_validate`).
3. **P3 — parameterized hook builders** (Med effort, halves the 50%-of-resources hook surface).
4. **P4 — acceptance test-skeleton drafter + native-dep preflight** (Med; completes the acceptance-authoring story begun this cycle).
5. **P5 — schema-diff / upgrade report** (Med-High; high per-event value, shared with breaking-change + migration work).
6. **P6 — live-API verification harness** (High; foundational, highest ceiling — fund when P1–P4 have cleared the cheap wins).
7. **P7 — azwise agent polish** (incremental, ongoing).
8. **P8 — variant authoring** (defer into the discriminator epic).

**Cheapest high-value cluster:** P1 + P2 (~1.5–2 weeks combined) remove the two
silent-failure classes and the every-resource boilerplate — the friction developers hit
first and most often — before the larger P5/P6 infrastructure investments.

## Relationship to existing backlogs
- `TODO.md` is **capability-centric** (docs gen, migration, list resources, identity); this report is **authoring-friction-centric**. Overlaps: P5↔#6/#8, P6↔#16, P4↔#1.
- Tech-summary Section A features 2 (azwise), 5 (customizers+hooks seams), 10 (self-verification gate), and 11 (acceptance-author) are the tooled/partially-tooled layers; P1–P4 target the **gaps between** them.
