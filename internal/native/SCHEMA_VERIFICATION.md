# Live-API Schema Verification — Design & Feasibility

Status: Draft (design/analysis — no implementation yet)
Owners: azapi provider team
Related: `DEVELOPER_SPEC.md` (§4 lifecycle, §5.2 edge cases, §6 verification),
`DEVELOPER_GUIDE.md` (hooks, acceptance scenarios), `GENERATOR.md` (customizers/overlay),
`skill://azwise` (knowledge categories), `TODO.md` #6/#10/#1.

---

## 1. Problem

The generated native schema (`internal/native/services/<svc>/<resource>_gen.go`) encodes
per-property *knowledge*: types, `Required`/`Optional`/`Computed`, `Sensitive`,
`ForceNew` (→ `RequiresReplace`), validators (enum/`OneOf`, int/float range, string
length, regex, ARM-ID/UUID, array `MaxItems`), and defaults. It is assembled from two
lossy sources:

- **bicep `types.json`** (swagger-derived) — carries shapes and read-only/required bits
  but *drops* most operational nuance: numeric ranges, string lengths, enum
  completeness, `MaxItems`, immutability, real defaults, conditional constraints.
- **azwise overlay** (AzureRM-derived) — recovers some of that, but only for curated
  resources, and reflects *AzureRM's* modeling choices, which are themselves not a
  faithful mirror of ARM API behavior (AzureRM adds client-side validation, models some
  mutable properties as `ForceNew` for UX/safety, and omits properties ARM accepts).

Neither source is the ground truth. The ground truth is **what the ARM API actually
does** when you send a value. This doc designs a mechanism to *verify* (and where cheap,
*discover*) that per-property knowledge against the live API.

---

## 2. What the API can and cannot tell us

Feasibility is per knowledge category. "Observable" means a single PUT/GET (± a prior
create) yields a signal that confirms or refutes the declared knowledge.

| Knowledge (schema / azwise field) | Live-API observable? | Probe | Positive signal | Caveat |
|---|---|---|---|---|
| Enum / `OneOf` (`StringRule.AllowedValues`) | Yes | PUT an out-of-set value | 4xx (`InvalidValue`/service code) | Some enums silently accept case variants or undocumented values |
| Int / float range (`IntRule`/`FloatRule` min/max) | Mostly | PUT `min-1`, `max+1` | 4xx, **or** clamp-on-GET | May clamp silently → must read-back and diff |
| String length (min/max) | Mostly | PUT under-/over-length | 4xx | Some services truncate silently |
| Regex / format (ARM-ID, UUID) | Partial | PUT malformed value | 4xx | Some formats are only validated client-side by AzureRM, never by ARM |
| Array `MaxItems`/`MinItems` (`ArrayRule`) | Yes | PUT oversized / empty array | 4xx | |
| `Required` (`RequiredFields` / `Required` flag) | Yes | PUT omitting it | 4xx (`MissingRequired`) | Conditionally-required needs the right baseline context |
| `ReadOnly` / `Computed` (`ForceComputed`) | Yes | PUT a value, then GET | value not reflected, or 4xx | This *is* the write-not-honored detector |
| `ForceNew` / immutable (`ForceNew`) | **Partial** | create, then PUT/PATCH a changed value, GET | 4xx (`CannotChange`) / changed / ignored | API reveals **immutability**, not "requires replace". AzureRM `ForceNew` on a *mutable* property (a UX choice) is **not** API-confirmable |
| Default value (`DefaultValues`) | Yes | PUT omitting, then GET | server-assigned value | Defaults can be region-/SKU-/tier-dependent |
| Write-accepted-but-not-honored (hook hold-list) | **Yes — only way** | PUT value, then GET, diff | normalized/overridden value on read | Undetectable without PUT→GET diff (e.g. `scmSiteAlsoStopped`, `siteConfig.websiteTimeZone`; see `web_site_hooks.go`) |

**Two load-bearing consequences:**

1. **"No error" ≠ "valid".** ARM frequently accepts an out-of-spec value and *coerces*
   it (clamp, normalize, ignore) rather than rejecting. Every probe **MUST read the
   resource back and diff**, never just check for a PUT error. The write-not-honored
   class (§5.2 of the spec) is *only* discoverable this way.
2. **`ForceNew` is a Terraform concept, not an API fact.** The API answers a narrower
   question — *is property X mutable in place?* — which is a **lower bound** on
   `ForceNew`. Immutable-per-API ⟹ must be `ForceNew`. Mutable-per-API ⟹ `ForceNew`
   is a policy choice azwise/AzureRM may still make; the API cannot confirm or refute it.

---

## 3. The probe surface — the central design decision

Three candidate surfaces to drive a probe at ARM:

| # | Surface | How | Verdict |
|---|---|---|---|
| A | **Typed native resource** (`azapi_web_site`) in a Terraform workspace + provider in `-debuggable` mode (the proposed path) | write `.tf`, `terraform apply`, mutate, re-apply, read ARM error | Right for **interactive debugging** and **regression**; wrong for **discovery** |
| B | **Dynamic `azapi_resource`** with `schema_validation_enabled = false` | raw `body` JSON straight to ARM through Terraform | Good middle ground when TF lifecycle semantics matter |
| C | **Direct ARM REST** via `internal/clients.ResourceClient` (`CreateOrUpdate`/`Get`/`Delete`) | Go harness, raw body, poll, GET, diff | **Best automated discovery/verification engine** |

**Why the typed resource (A) is the wrong probe for discovery.** The typed schema *is the
artifact under test*. Its own validators reject out-of-set enums, out-of-range ints, and
oversized arrays **at plan time, before ARM ever sees the value** (`ApplyExpectError` in
the acceptance framework targets exactly this provider-side gate). Using the typed
resource to probe the API therefore *filters out the very signal we want to measure* — we
would only learn that the schema rejects what the schema declares, a tautology. It also
cannot reach properties the generator didn't emit, and pays a full plan/apply/refresh/
state cycle per probe.

The correct discovery surface bypasses the typed schema and hits ARM with an arbitrary
body. Both **B** (`azapi_resource`, `schema_validation_enabled = false`,
confirmed at `internal/services/azapi_resource.go:520`) and **C** (raw `ResourceClient`)
do this. **C is preferred** for the engine: no HCL round-trip, no TF plan/state overhead,
exact control of the request body and mutation sequence, and it returns the raw ARM error
envelope (`azcore.ResponseError` via `runtime.NewResponseError`,
`internal/clients/resource_client.go:90,149`). **B** is the fallback when a probe must
observe Terraform's own lifecycle (e.g. does `RequiresReplace` fire), and **A** stays as
the human-in-the-loop debugger and the final regression gate.

**Reframe the goal accordingly.** Brute-force *discovery* of unknown constraints
(binary-searching every numeric range, fuzzing every string, N-way property interaction)
is combinatorially and financially intractable against live ARM. The tractable, valuable
goal is **verify/falsify the *declared* knowledge with a tight boundary probe**
(value-at-limit + value-just-past-limit), and surface discrepancies as suggested azwise/
customizer edits. Discovery of genuinely new constraints is an opportunistic byproduct
(e.g. a PUT that unexpectedly 4xx's on a property we declared unconstrained).

---

## 4. Architecture

```
                       ┌──────────────────────────────────────────────┐
 declared knowledge    │  Probe Planner                               │
 ─ services.Registry[] │  per property → probe cases derived from the │
   .Schema() validators│  DECLARED rule: control (in-spec) + boundary │
 ─ azwise SchemaKnow-  │  (just-past-spec) + expectation              │
   ledge/ResourceKnow- └───────────────┬──────────────────────────────┘
   ledge accessors                     │ []ProbeCase{ baseline, mutation, expect }
                                        ▼
 baseline body   ─────►  ┌──────────────────────────┐
 (acceptance *_Complete  │  Executor                │   REST (primary):
  config, one-variable-  │  apply baseline once,     │   clients.ResourceClient
  at-a-time perturbation)│  then each mutation on the│   TF (fallback): nativeacc.
                         │  live resource, PUT→poll→ │   Workspace + azapi_resource
                         │  GET, capture (err, body) │   (schema_validation_enabled=false)
                         └───────────┬──────────────┘
                                     ▼
                         ┌──────────────────────────┐
                         │  Classifier              │   Confirmed / Refuted /
                         │  parse azcore.Response-   │   Coerced(write-not-honored) /
                         │  Error {code,target}; ALW-│   Inconclusive(transient/quota)
                         │  AYS read-back-and-diff   │
                         └───────────┬──────────────┘
                                     ▼
                         ┌──────────────────────────┐
                         │  Reporter                │   per-resource table:
                         │  property × declared ×    │   property | declared | observed |
                         │  observed × verdict +     │   verdict + suggested azwise edit
                         │  suggested edits          │
                         └──────────────────────────┘

              Safety Governor wraps the Executor: isolated random-named RG per batch,
              concurrency cap, destructive-probe gate, 429/quota/region retry+classify,
              guaranteed teardown.
```

### Reuse map (what already exists — build the thin parts only)

| Component | Reuse |
|---|---|
| REST executor | `internal/clients.ResourceClient` (`CreateOrUpdate`/`Get`/`Delete`, LRO poll built in) |
| TF executor + isolation/teardown | `internal/native/acceptance` `Workspace` (in-process reattach provider, scoped RG lifecycle, `GINKGO_DUMP_TF_CONFIG*` config dump) |
| Interactive debugging | `main.go -debuggable` + reattach; the `pkg/<name>/` dev workspaces already in the tree |
| Declared knowledge (planner input) | azwise `SchemaKnowledge`/`ResourceKnowledge` accessors (`GetStringRules`/`GetIntRules`/`GetFloatRules`/`GetForceNewPaths`/`GetRequiredFields`/`GetDefaultValues`/`GetComputedFields`); and `services.Registry[name].Schema()` for the emitted validators/flags |
| Baseline bodies | acceptance `*_Basic`/`*_Complete` config builders (`services/<svc>/<resource>_config.go`) |
| Body ↔ state (only if using TF path) | `internal/native/mapper` `Expand`/`Flatten` |
| Error envelope | `azcore.ResponseError` (`.StatusCode`, `.ErrorCode`, raw `error.{code,message,target}`) |
| Write-not-honored precedent | `web_site_hooks.go` `normalizeWebSiteResponse`/`isDefaultRestriction` — the canonical PUT→GET-diff finding this mechanism generalizes |

Genuinely new code = Probe Planner, Classifier, Reporter, Safety Governor, and the
thin REST executor wrapper. Estimated as a CLI under `internal/native/cmd/azapin-verify/`.

---

## 5. Probe taxonomy (one-variable-at-a-time)

Every probe starts from a **known-valid baseline body** (the `*_Complete` acceptance
config) and perturbs **exactly one property**, so a failure attributes cleanly. Cases:

- **Enum** — control: an in-set value (expect 2xx). Boundary: an out-of-set token (expect
  4xx). Refuted if 2xx *and* the value round-trips → the declared enum is incomplete.
- **Int/float range** — control: the declared max (expect 2xx). Boundary: `max+1` /
  `min-1` (expect 4xx). If 2xx, GET and diff: reflected `max+1` → declared max wrong;
  clamped to max → silent-clamp (note, don't treat as valid).
- **String length** — control at limit, boundary at `limit+1`; same diff logic.
- **Array MaxItems** — control `N`, boundary `N+1` elements.
- **Required** — omit the property from an otherwise-valid create (expect 4xx
  `MissingRequired`). 2xx ⟹ property is actually optional → refute `RequiredFields`.
- **Computed / ReadOnly** — PUT a distinct value into a declared-computed path, GET,
  diff. Reflected ⟹ it is actually settable (false `ComputedFields` → silent data loss
  risk, per skill checklist item 4). Ignored/normalized ⟹ confirmed read-only (or
  write-not-honored).
- **Immutable / ForceNew** — create with value A, then PUT/PATCH value B on the *same*
  resource, GET. 4xx `CannotChange` ⟹ immutable ⟹ `ForceNew` justified. Changed to B
  ⟹ mutable ⟹ `ForceNew`, if declared, is a *policy choice* (flag, don't refute).
  Ignored (stays A, no error) ⟹ write-not-honored on update.
- **Default** — omit the property, GET, record the server-assigned value; compare to the
  declared `DefaultValue`.

---

## 6. Classifier — outcome → verdict

Inputs per case: `(putErr, putBody, getBody)`. Verdicts:

- **Confirmed** — observed behavior matches the declared rule.
- **Refuted** — schema declares a constraint the API does not enforce (accepts + reflects
  an out-of-spec value; accepts omission of a "required" field; accepts a change to a
  "ForceNew"/immutable field). Highest-value output — a concrete schema/azwise bug.
- **Coerced / write-not-honored** — PUT accepted, GET differs from what was sent
  (clamp/normalize/ignore). Feeds the hooks hold-list convention (spec §5.2).
- **Inconclusive** — transient/quota/regional/throttle (429, `SubscriptionNotRegistered`,
  capacity). Retried with backoff; still failing → skipped and flagged, never scored as
  a constraint.

Error attribution parses `azcore.ResponseError`: `StatusCode`, `ErrorCode`, and
`error.target` (ARM often names the offending JSON path), matched back to the perturbed
property so a probe that fails for an unrelated reason isn't miscredited.

---

## 7. Feasibility caveats / risks (must be designed around)

1. **Cost / time / quota.** Every probe is a real create+mutate+delete — minutes each,
   real spend, subscription core/count quotas. → Batch: one baseline create, many
   independent in-place mutations against the *same* live resource; parallelize across
   isolated RGs under the governor's concurrency cap; scope to the declared boundary, not
   an exhaustive sweep.
2. **Coupled / conditional constraints.** Validity often depends on another property
   (SKU tier gates features; `websiteTimeZone` only honored while the app is stopped).
   One-variable-at-a-time on a fixed valid baseline controls for this but yields
   *conditional* truth; the report must record the baseline context.
3. **ARM inconsistency.** Accept-and-coerce vs. reject varies by RP. Mandatory
   read-back-and-diff (§2) is the mitigation.
4. **Async / transient noise.** LRO polling, provisioning states, 429s, regional
   capacity. The classifier's Inconclusive bucket + retry isolates these from real
   constraints.
5. **`ForceNew` ≠ immutable (§2).** Report immutability as a lower bound; never
   auto-refute a `ForceNew` on an API-mutable property.
6. **Destructive probes.** Some mutations trigger data loss or long re-provisioning.
   ForceNew/immutability probes are gated (opt-in) and always run on throwaway resources.
7. **Coverage honesty.** This verifies declared boundaries; it does not prove the absence
   of undeclared constraints. Report states exactly which properties were probed and how.

---

## 8. Recommended phased pathway

- **Phase 1 — REST PoC (one resource).** Thin `ResourceClient` wrapper; hand-written probe
  set for `azapi_web_site` (reuse its `*_Complete` body as baseline); manual classifier;
  printed report. Proves the loop end-to-end and the read-back-diff discipline. Validate
  against the two *known* web-site findings (`scmSiteAlsoStopped`, `websiteTimeZone`) — the
  mechanism must independently rediscover them.
- **Phase 2 — Probe Planner from declared knowledge.** Generate probe cases mechanically
  from `services.Registry[name].Schema()` + azwise accessors, so any registered resource
  is probeable without hand-authoring.
- **Phase 3 — Classifier + Reporter + suggestions.** Structured verdicts; per-resource
  report emitting suggested azwise/customizer edits (add `IntRule` max, drop a false
  `RequiredField`, add to the write-not-honored hold-list, justify/relax a `ForceNew`).
- **Phase 4 — TF executor + regression gate.** Add the `azapi_resource`
  (`schema_validation_enabled=false`) and typed-resource paths (reuse `nativeacc.Workspace`)
  so the *generated schema* can be regression-checked against the REST-discovered truth;
  keep `-debuggable` for interactive drill-down.
- **Phase 5 — Governor + CI cadence.** Batching, concurrency cap, quota/region handling,
  guaranteed teardown; run out-of-band (nightly/opt-in, `TF_ACC`-style gate), never in the
  fast unit lane. Ties into TODO #6 (breaking-change detector) and #10 (constraint engine).

---

## 9. Open decisions (for the owner)

1. **Primary engine surface** — confirm REST (`ResourceClient`) as the discovery engine,
   with the typed-resource TF path as regression only (this doc's recommendation), vs.
   the originally-proposed typed-resource-first workspace path.
2. **Scope of run** — verify declared knowledge only (recommended, tractable) vs. attempt
   bounded discovery of unknown numeric/length boundaries (expensive).
3. **Output** — report-only vs. auto-drafting azwise/customizer patches for review.
4. **Cadence & budget** — per-resource opt-in vs. nightly sweep; per-run Azure spend/quota
   ceiling.
