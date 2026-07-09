# azapin (azapi-native/next) — Development Specification

Status: Draft (PoC proven end-to-end against live Azure; productionization in progress)
Owners: azapi provider team
Related docs: `DESIGN.md`, `GENERATOR.md`, `RESOURCE.md` (under `internal/native/`),
`azapin-vs-azapi.md`, `static-schema-generation-analysis.md`,
`approach-comparison-report.md` (under `tools/azwise/`).

---

## 1. Project Overview & Goals

### 1.1 Purpose

**azapin** generates **static, typed Terraform resources** for Azure (e.g.
`azapi_storage_account`) from the bicep type definitions Azure already publishes,
and layers **azwise** — curated AzureRM-derived operational knowledge — on top so
the resources behave like hand-written AzureRM resources rather than a raw spec
dump. The resources ship inside the existing **terraform-provider-azapi** binary,
alongside the dynamic `azapi_resource`.

Two subsystems:

- **azwise** (`internal/azure/azwise/`) — a registry of per-resource operational
  knowledge (ForceNew rules, validation, real default values, sensitive/computed
  classification, timeouts). Consumed at runtime by `azapi_resource` and at
  build-time by the azapin generator.
- **azapin** (`internal/native/`) — the static-schema generator, the generic
  runtime resource base, and the supporting tooling (validators, acceptance
  framework, CLI).

### 1.2 Business problem

AzureRM remains important, but feature delivery is constrained by finite review
bandwidth and by a provider stack that depends on Pandora and the HashiCorp SDK,
where upstream swagger or SDK churn can stall new-resource work or maintenance.
azapin is one response to that constraint — improve the azapi user experience,
reuse Azure's own published type data directly, and make migration from AzureRM to
azapi materially easier for customers who want typed resources without waiting for
full AzureRM coverage.

`azapi_resource` is the universal, day-zero way to manage any Azure resource. It
already validates the dynamic `body` against the embedded ARM/bicep types at plan
time (`schemaValidate`, on by default via `schema_validation_enabled`). In VS Code,
the Azure Terraform extension/LSP adds body completion, hover, and diagnostics for
azapi resources. But that help lives in an external editor integration, not in the
provider schema itself: core Terraform still sees one dynamic `body`, errors land
on JSON paths rather than typed attributes, computed/read-only values are not part
of the resource (you reach them via `response_export_values` → `.output`), defaults
are not injected, and there is no per-attribute lifecycle modeling (ForceNew,
computed handling, real defaults).

Teams that want an **AzureRM-like authoring experience** — a native typed schema
with per-attribute validation, editor support that comes from the provider itself,
first-class computed attributes, and correct lifecycle — must otherwise wait for
AzureRM to add the resource or hand-write provider code. azapin closes that gap by
generating those typed resources mechanically from Azure's own schema, with azwise
supplying the human-curated nuances a raw spec can't express.

### 1.3 Target audience

- **Terraform practitioners** managing stable Azure resources who want
  autocomplete, validation, and clean diffs without leaving the azapi provider.
- **Provider maintainers** who want to expand typed coverage at codegen speed
  instead of per-resource hand-authoring.
- **Teams migrating from / comparing against AzureRM** who want parity behavior on
  the azapi backend.

### 1.4 Goals & non-goals

**Goals**
1. Generate framework-valid typed schemas for the latest stable API version of each
   stable ARM resource type (~2,631 types).
2. One generic runtime base implementing the full Terraform resource lifecycle —
   no per-resource Go struct.
3. Overlay azwise knowledge to deliver ForceNew, verified defaults, validation, and
   sensitive/computed correctness.
4. Coexist with `azapi_resource`; identical auth, client, and backend.
5. Per-resource customization seams: generation-time schema **customizers**
   (validators/defaults/ForceNew baked into the generated schema), runtime hooks,
   and method override.

**Non-goals**
1. Replacing `azapi_resource` (it remains the day-zero / preview / escape hatch).
2. Generating per-resource Go model structs.
3. Statically typing polymorphic/discriminated bodies (fall back to dynamic).
4. AzureRM property-name parity (azapin uses mechanical snake_case from the spec).

---

## 2. User Requirements

User stories with acceptance criteria. **MUST/SHOULD/MAY** per RFC 2119.

### 2.1 Practitioner — author a typed resource

> As a Terraform user, I want to declare an Azure storage account with named,
> typed attributes so I get autocomplete and plan-time validation.

**Acceptance**
- `azapi_storage_account` exposes ARM body properties as typed attributes in
  snake_case (`sku`, `properties.access_tier`, …). **MUST**
- `terraform validate` rejects unknown attributes and wrong types before apply. **MUST**
- Enum attributes reject out-of-set values via schema validators. **MUST**
- The resource MUST create the identical ARM resource that an `azapi_resource` with
  the equivalent `body` would create.

### 2.2 Practitioner — clean plans

> As a Terraform user, I want `apply` then `plan` to be empty (no perpetual diff).

**Acceptance**
- After a successful `apply`, a follow-up `plan` with unchanged config MUST be
  empty (idempotent).
- Server-computed values (e.g. `properties.primary_endpoints`, `provisioning_state`)
  MUST round-trip into state and not re-plan as `(known after apply)`.
- Apply results MUST be fully known (no unknown values returned post-apply).

### 2.3 Practitioner — correct lifecycle

> As a Terraform user, I want only the right changes to force replacement.

**Acceptance**
- Changing a declared ForceNew property (e.g. `properties.is_hns_enabled`) MUST plan
  a replace; changing a normal property MUST plan an in-place update.
- Conditional rules that a static modifier can't express (e.g. storage **SKU
  zone-migration** `Standard_LRS → Standard_ZRS`) MUST force replace, while an
  in-tier SKU change MUST NOT. **MUST**
- Read-only attributes MUST be `Computed` and MUST NOT be sent in the request body.
- Optional attributes with a known AzureRM default SHOULD surface that default at
  plan time rather than `(known after apply)`.

### 2.4 Practitioner — import & outputs

**Acceptance**
- `terraform import azapi_storage_account.x <armId>` MUST hydrate typed state from a
  GET.
- Computed values MUST be referenceable directly
  (`azapi_storage_account.x.properties.primary_endpoints.blob`) without
  `response_export_values`.

### 2.5 Maintainer — add a resource

> As a maintainer, I want to add a new typed resource by generating it.

**Acceptance**
- Running the generator for an ARM type produces a compiling `*.go` schema file that
  passes `Schema.ValidateImplementation`.
- The generated schema MUST cover exactly the bicep body properties (validated by
  the schema↔bicep cross-validator: no extra, no missing, no type-mismatched).
- The resource auto-registers (via `init()` → `services.Registry`) and the provider
  exposes it with no further wiring.

### 2.6 Maintainer — customize a resource

> As a maintainer, I want to add resource-specific logic without forking the base.

**Acceptance**
- A maintainer MAY register a generation-time **customizer** (keyed by ARM type in
  `internal/native/generator/customizers`) that mutates the parsed type graph so
  validators/defaults/ForceNew are baked into the generated schema and never
  mutated at runtime. **SHOULD** be preferred for schema-shape rules.
- A hand-written generated-service hook file MAY register `Hooks` (Before/After per
  op, ModifyPlan, ValidateConfig) keyed by resource name.
- A resource MAY override a base method via Go embedding when hooks are
  insufficient.

### 2.7 Maintainer — curate knowledge (azwise)

**Acceptance**
- azwise knowledge MUST be expressible declaratively (ForceNew paths, string/int
  rules, sensitive/computed fields, default values, timeouts).
- The generator MUST apply azwise as an overlay that wins over heuristics.
- azwise MUST be toggleable for `azapi_resource` via the
  `disable_resource_knowledge` provider feature.

### 2.8 Operator — safe defaults

**Acceptance**
- With `TF_ACC` unset, acceptance suites MUST skip (no Azure calls).
- azwise validation MUST NOT false-positive (only flag sensitive/computed fields
  actually present in the body).

---

## 3. System Architecture & Tech Stack

### 3.1 Tech stack

| Concern | Choice |
|---|---|
| Language | Go (module `github.com/Azure/terraform-provider-azapi`) |
| Provider framework | `terraform-plugin-framework` v1.19.0 (Protocol v6) |
| Plugin protocol | tfprotov6 via `providerserver.NewProtocol6WithError` |
| Acceptance testing | `terraform-plugin-testing` v1.16.0 + **Ginkgo** v2.21.0 / **Gomega** v1.34.2 |
| Unit testing | Go `testing` + `testify` |
| "Database" | None. Two embedded, read-only data sources: the **bicep type graph** (`internal/azure/generated`, `go:embed`) and the **azwise knowledge registry** (compiled Go) |
| External API | Azure Resource Manager REST (via `internal/clients.ResourceClient`) |
| Code generation | Custom Go generator (`internal/native/generator`), run via `go run` / `go generate` |

There is **no frontend, no server, no database**. The "frontend" is the Terraform
CLI; the "backend" is the provider plugin process plus ARM.

### 3.2 Component map

```
terraform CLI ──tfprotov6──► terraform-provider-azapi (plugin process)
                                  │
        ┌─────────────────────────┼──────────────────────────────┐
        ▼                         ▼                              ▼
  azapi_resource          azapin runtime resource +        azapi data sources
  (dynamic body)          data source (native/resource)    (dynamic; unchanged)
        │                         │
        │                  ┌──────┴───────┐
        │                  ▼              ▼
        │            mapper (state↔   loader (bicep body
        │            ARM JSON)        type graph, cached)
        ▼                  │              │
   internal/clients.ResourceClient ──REST──► Azure Resource Manager
        ▲                  ▲
        │                  │
   internal/azure/azwise (knowledge): Validate / CheckForceNew /
   StripComputedFields / TimeoutDefault — used by BOTH paths
```

Build-time (not in the plugin process):

```
internal/azure/generated/*/types.json  (embedded source of truth)
        │
        ▼  internal/native/generator
  ParseTypesJSON ─► PostProcess(+ApplyAzwise from azwise, +envelope) ─► customizers.Apply ─► EmitSchema
        │                                                       │
        ▼ ValidateEmittedSchema (source-string)                 ▼
  internal/native/services/<service>/<resource>_gen.go  ──compiled──►  runtime Registry
        ▲
  internal/native/validate (compiled-schema cross-check, run as test/CLI)
```

### 3.3 Package responsibilities

| Package | Responsibility |
|---|---|
| `internal/azure/azwise` | Knowledge registry + interface; `Validate`, `CheckForceNew`, `StripComputedFields`, `TimeoutDefault`, `SchemaKnowledge` accessors |
| `internal/native/naming` | ARM type → TF resource name; camelCase ↔ snake_case |
| `internal/native/generator` | `types.json` walker, post-processing, azwise overlay, schema emitter, source-string validator |
| `internal/native/generator/customizers` | Generation-time per-ARM-type schema customizers (`Register`/`Apply`) — bake validators/defaults/ForceNew into the schema |
| `internal/native/armtypes` | ARM resource type string constants |
| `internal/native/schema` | Runtime static-default impls (`Static*`), shared generic validators (`UUID`, `AzureResourceID`), plan-modifier anchors |
| `internal/native/mapper` | Generic state ↔ ARM JSON (`Expand`/`Flatten`/`FlattenInto`/`ResolveUnknowns`) |
| `internal/native/resource` | Generic `Base` resource + read-only `DataSource` (schema converted from the resource schema, GET-only read path), hook types/registry, body loader |
| `internal/native/services` | Registry (`Descriptor`/`Register`/`Registry`); per-service sub-packages `services/<service>` (schemas) + `services/<service>/validators/` + hand-written `<resource>_hooks.go`; `services/all` blank-imports every service to populate the registry and hooks |
| `internal/native/validate` | Compiled-schema ↔ bicep cross-validator |
| `internal/native/acceptance` | Ginkgo BDD acceptance framework |
| `internal/native/cmd/azapin-validate` | Standalone validation CLI |

---

## 4. Technical Design

### 4.1 Data models

#### Bicep type graph (`generator`)

```go
type Type struct {
    Kind        TypeKind            // String, StringLiteral, Int, Bool, Object, Array, Union, Any
    Name        string              // ARM struct/enum name (or literal value)
    Properties  map[string]*Property // ObjectType, keyed by ARM camelCase name
    ElementType *Type               // ArrayType element
    Elements    []*Type             // UnionType members
}

type Property struct {
    Name         string       // ARM JSON name (camelCase)
    Type         *Type
    Flags        PropertyFlag // Required | ReadOnly | WriteOnly | SystemManaged (bicep bits)
    Description  string
    DefaultValue string                 // description-mined or azwise-verified
    Validators   []DescriptionValidator // ARM-ID / regex / range / OneOf / length / custom / shared
    ForceNew      bool        // azwise overlay → RequiresReplace
    Sensitive     bool        // azwise SensitiveFields leaf overlay → Sensitive: true
    ForceComputed bool        // azwise overlay → Computed-only
}
```

The walker resolves `$ref` into a **DAG** (shared nodes, cycle-broken by an
in-progress set — recursive back-edges become a `KindAny` sentinel).

#### azwise knowledge

```go
type ResourceKnowledge interface {
    GetResourceType() string
    GetApiVersions() []string
    Validate(name string, body map[string]any, hasSensitiveBody bool) diag.Diagnostics
    CheckForceNew(oldBody, newBody map[string]any) bool
    TimeoutDefault(operation string, fallback time.Duration) time.Duration
    IsSoftDelete() bool
    GetComputedFields() []string // read-only (server-computed) paths
    GetDefaultFields() []string  // paths with server-side defaults
    GetDefaultValues() []DefaultValue
    GetRequiredFields() []string
}
type SchemaKnowledge interface { // consumed by the generator
    GetForceNewPaths() []string
    GetSensitiveFields() []string
    GetStringRules() []StringRule
    GetIntRules() []IntRule
    GetFloatRules() []FloatRule
}
// BaseKnowledge implements both; concrete resources embed it (storage_account.go, …).
```

#### Generated descriptor & runtime base

```go
// internal/native/services
type Descriptor struct {
    Name           string // "azapi_storage_account"
    ARMType        string // "Microsoft.Storage/storageAccounts"
    APIVersion     string // "2025-01-01"
    Schema         func() schema.Schema // generated body + operational envelope
    WritableScopes int                  // bicep scope bitmask (metadata)
    ParentAttr     string               // generated parent-reference attr, e.g. "resource_group_id"
    Relational     []RelationalConstraint // azwise cross-property constraints → framework ConfigValidators
    Timeouts       Timeouts               // per-op timeout defaults baked in from azwise; zero field → runtime fallback
}
var Registry = map[string]Descriptor{} // each generated init() calls Register; services/all blank-imports every service package to populate it

// internal/native/resource
type Base struct { desc Descriptor; hooks *Hooks; provider *clients.Client }
type Hooks struct {
    BeforeCreate, AfterCreate, BeforeUpdate, AfterUpdate,
    BeforeRead, AfterRead, BeforeDelete func(*CrudCtx)
    Singleton      *SingletonDefault // fixed-named default child: skip create existence check, reset-via-PUT on destroy (no BeforeDelete)
    ValidateConfig func(ctx, req, resp)
    ModifyPlan     func(ctx, req, resp)
}
// Per-hook-point field liveness (which of Plan/State/Body/Response are populated,
// whether a Body mutation reaches ARM, lifecycle position, Singleton suppression)
// is documented on Hooks/CrudCtx in resource/hooks.go and guarded against drift by
// resource/base_hook_contract_test.go. That seam doc is authoritative.
type CrudCtx struct {
    Ctx context.Context; Client *clients.Client; ID parse.ResourceId
    Plan, State types.Object   // Plan live on create/update; State live on read/delete
    Body, Response map[string]any // Body live on create/update (mutate in Before*); Response live in After*
    Diags *diag.Diagnostics
}
```

### 4.2 Composed schema

The operational envelope (`name`, the parent reference, `id`) is **generated into
the schema** at generation time — the generator synthesizes it from the ARM type
and writable scope, so the shipped schema is final and never mutated at runtime.
The runtime base adds only the `timeouts` block (it needs a `context.Context` the
static schema function can't hold):

```
final schema = generated body attributes (location, tags, sku, kind, identity, properties, …)
             + name          (Required, RequiresReplace, optional name validator)
             + <parent>_id   (Required, RequiresReplace; resource-specific, e.g. resource_group_id; in Descriptor.ParentAttr)
             + id            (Computed,  UseStateForUnknown)
             + timeouts      (runtime-added block: create/read/update/delete)
```

### 4.3 Generation pipeline (build-time sequence)

```mermaid
sequenceDiagram
    participant CLI as go run generator
    participant W as Walker
    participant P as PostProcess
    participant A as azwise
    participant C as customizers
    participant E as Emitter
    participant V as ValidateEmitted
    participant FS as services/*.go

    CLI->>W: ParseTypesJSON(types.json)
    W-->>CLI: DAG (cycle/discriminated-safe)
    CLI->>P: PostProcess(defs)
    P->>P: extract defaults + validators from descriptions
    P->>P: promote single-optional child → Required
    P->>A: ApplyAzwise(def) → Get(armType,version)
    A-->>P: ForceNew/Computed/Sensitive/Defaults/Rules
    P->>P: seed operational envelope (name + parent reference) from scope
    P-->>CLI: enriched DAG
    CLI->>C: customizers.Apply(defs)
    C-->>CLI: per-ARM-type validators/defaults baked into graph + envelope
    CLI->>E: EmitSchema(def)
    E-->>CLI: Go source (package <service>, flags, validators, plan modifiers, init())
    CLI->>V: ValidateEmittedSchema(source, body)
    V-->>CLI: 0 mismatches (else fail)
    CLI->>FS: write services/storage/storage_account_gen.go
```

### 4.4 Create / Update (runtime sequence)

```mermaid
sequenceDiagram
    participant TF as terraform
    participant B as Base
    participant M as mapper
    participant AW as azwise
    participant H as Hooks
    participant C as ResourceClient
    participant ARM as Azure RM

    TF->>B: ModifyPlan (RequiresReplace from schema + hook SKU-zone via azwise.CheckForceNew)
    TF->>B: Create(plan)
    B->>B: parse.NewResourceID(name, <parent>_id (Descriptor.ParentAttr), type@version)
    B->>B: b.timeout("create") (from Descriptor.Timeouts)
    B->>M: Expand(plan, bodyGraph) → ARM JSON
    B->>AW: StripComputedFields(body)
    B->>H: BeforeCreate(CrudCtx{Body})
    B->>C: CreateOrUpdate(id, apiVersion, body)
    C->>ARM: PUT (+ poll)
    B->>C: Get(id, apiVersion)
    C->>ARM: GET
    ARM-->>B: response JSON
    B->>H: AfterCreate(CrudCtx{Response})
    B->>M: FlattenInto(response, plan) → state object
    B->>B: set id (computed)
    B->>M: ResolveUnknowns(state)
    B-->>TF: fully-known state
```

### 4.5 Read / refresh

```mermaid
sequenceDiagram
    participant TF as terraform
    participant B as Base
    participant C as ResourceClient
    participant M as mapper
    participant H as Hooks
    TF->>B: Read(state)
    B->>H: BeforeRead(CrudCtx{State})
    B->>C: Get(id, apiVersion)
    alt 404
        C-->>B: not found
        B-->>TF: RemoveResource
    else found
        C-->>B: response JSON
        B->>H: AfterRead(CrudCtx{Response})
        B->>M: FlattenInto(response, priorState)
        B-->>TF: refreshed state
    end
```

### 4.6 Mapper contract (state ↔ ARM JSON)

- **Expand**: walk the bicep body graph; for each non-SystemManaged property read
  the plan attribute by `CamelToSnake(armName)`; skip null/unknown; emit keyed by
  the **ARM camelCase name from the graph** (never by reversing snake_case).
- **Flatten / FlattenInto**: walk the graph/schema; only overwrite body attributes
  the response actually contains (preserves write-only and envelope values).
- **ResolveUnknowns**: recursively null any value still unknown after flatten
  (apply results must be fully known).

### 4.7 User flows

- **Author** (practitioner): write `resource "azapi_storage_account"` with typed
  attributes → `plan`/`apply` → outputs via direct attribute refs.
- **Add a resource** (maintainer): run generator → file auto-registers → ship.
- **Customize** (maintainer): register a generation-time customizer (bakes
  validators/defaults/ForceNew into the schema), and/or runtime hooks in
  `services/<service>/<name>_hooks.go`, or embed `*Base` to override a method.
- **Curate knowledge** (maintainer): edit/add an azwise `BaseKnowledge` entry →
  regenerate → overlay flows into the schema.

---

## 5. Constraints & Edge Cases

### 5.1 Non-functional requirements

**Security**
- Sensitive properties MUST be `Sensitive: true` (azwise `SensitiveFields`) and,
  for genuinely settable secrets, SHOULD be moved to `sensitive_body` semantics; the
  validator MUST only flag fields actually present in the body (no false positives).
- No credentials are stored; auth reuses the provider's `ARM_*` env-based clients.
- Generated files carry a `// DO NOT EDIT` header; no runtime `eval` of spec data.

**Performance**
- The bicep body graph is parsed once per `ARMType@APIVersion` and **cached**
  (`loader.go`) for the plugin process lifetime.
- The provider binary already embeds the bicep types (~334 MB); generated Go adds
  marginal size. Generated resources are sub-packaged by service
  (`services/<service>/`) to bound compile time as coverage grows.

**Reliability / correctness**
- The generated schema MUST pass `Schema.ValidateImplementation`.
- The schema MUST round-trip against the bicep graph (compiled-schema validator: 0
  errors). Storage account: 236 properties, 0 mismatches.
- Apply MUST produce fully-known state (`ResolveUnknowns`).

**Accessibility / DX** (no UI; "accessibility" = maintainer/practitioner ergonomics)
- Each generated attribute carries the ARM description.
- Docs (`DESIGN.md`/`GENERATOR.md`/`RESOURCE.md`) MUST stay current with the
  generator rules and the resource lifecycle.

### 5.2 Edge cases & required handling

| Edge case | Handling |
|---|---|
| Recursive ARM types (`ErrorDetail.details[]`) | walker breaks cycles → `KindAny` sentinel; no stack overflow |
| Discriminated/polymorphic bodies | parsed without failing the walk; emit `DynamicAttribute`; root-level → resource skipped (use `azapi_resource`) |
| Dynamic type inside a list | framework forbids it → degrade to `StringAttribute` inside `ListNested` |
| `Default` requires `Computed` | emit `Optional + Computed + Default` |
| Optional+Computed perpetual diff | `UseStateForUnknown` on every Computed (non-default) attribute |
| Read-only field marked ForceNew | dropped (a Computed field can't `RequiresReplace`) |
| Shared type node + per-path overlay | bicep dedups structurally identical types; a customizer can un-share an **array** element via `IsolateArrayElement` (e.g. `ipRules` vs `ipv6Rules`). Object-node sharing (e.g. `encryption.services.*.key_type`) remains a **known limitation** for azwise overlays; fix = path-aware emission |
| additionalProperties maps (`tags` content, user-assigned identities) | not modeled as typed sub-attributes (empty nested object) — **known limitation** |
| ARM omits an Optional+Computed field on GET | `FlattenInto` keeps prior value; `ResolveUnknowns` nulls leftover unknowns |
| ARM accepts a writable property on PUT but normalizes/overrides it on GET (write-accepted, not honored — e.g. `Microsoft.Web/sites` `scmSiteAlsoStopped`, `siteConfig.websiteTimeZone`) | provider reflects the server value (never fabricate the read); hold the property at its round-tripping value in acceptance configs so the update chain stays drift-free |
| Name collisions across ARM types | build-time collision detection (9/3,246); resolved via namespace prefix |
| Resource not found on read | `RemoveResource` |
| azwise has no knowledge for a type | overlay is a no-op; schema from spec + descriptions only |

### 5.3 Error handling

- Build-time: the generator MUST fail (non-zero) on `ValidateEmittedSchema` errors
  (extra/type-mismatch), refusing to write the file. Resources whose root body is
  non-object MUST be skipped (logged), not crash the batch.
- Runtime: ARM errors surface as Terraform diagnostics with the resource ID; 404 on
  read/delete is tolerated; provider-not-configured is an explicit diagnostic.

---

## 6. Verification

### 6.1 Testing strategy (layers)

1. **Unit tests** (offline, `go test`):
   - `naming` — conversion + collision scan over all 3,246 ARM types.
   - `generator` — walker (incl. `TestParseSelfReferentialType`,
     `TestParseDiscriminatedObjectType`), post-processing, azwise overlay
     (`TestApplyAzwiseStorageAccount`), emitter.
   - `mapper` — round-trip Expand/Flatten on storage; null/unknown handling.
   - `validate` — compiled schema (`services.Registry[...].Schema()`) ↔ bicep (0 mismatches);
     synthetic mismatch detection.
   - `resource` — `Schema.ValidateImplementation`; interface assertions; hook
     registration; cached body loader.
   - `azwise` — Validate/ForceNew/timeouts/sensitive semantics.
2. **Generation smoke** (offline): generate schemas for every latest-stable type;
   assert **0 parse failures, 0 panics** (current: 1,564 emit OK, 28 polymorphic skips).
3. **Schema validity gate**: every shipped generated resource MUST pass
   `Schema.ValidateImplementation` in a resource-layer test.
4. **Acceptance** (Ginkgo, gated on `TF_ACC` + `ARM_SUBSCRIPTION_ID`): one BDD suite
   per resource; scenarios = create+import, in-place update, full-mutation updatability
   (`*_Complete` → `*_Complete_update`), conditional ForceNew,
   verified default; CheckDestroy verifies deletion via ARM GET.
5. **Manual workspaces** (`pkg/azapin/`, `pkg/azapi/`, gitignored): dev-override
   Terraform configs for hands-on debugging and the azapi-vs-azapin demo.

### 6.2 Deployment criteria (per shipped resource)

A generated resource is shippable when:
- It compiles and passes `Schema.ValidateImplementation`. **MUST**
- Compiled-schema ↔ bicep validation reports 0 errors. **MUST**
- Acceptance suite passes: create, **idempotent empty re-plan**, update, import,
  destroy. **MUST**
- ForceNew, computed round-trip, and verified defaults behave per §2.3. **MUST**
- Any azwise over-application (shared-node) is reviewed/accepted or fixed. **SHOULD**

### 6.3 Step-by-step implementation plan

**Phase 0 — done (PoC)**
- [x] Generator: walker (cycle/discriminated-safe), post-processing, azwise overlay,
      emitter, source-string + compiled-schema validators, CLI.
- [x] Runtime: generic `Base` (CRUD/ModifyPlan/ValidateConfig/Import), mapper
      (Expand/Flatten/FlattenInto/ResolveUnknowns), body loader, hooks.
- [x] `azapi_storage_account` generated, registered, **applied against live Azure**.
- [x] Ginkgo acceptance framework + storage scenarios.
- [x] azwise overlay (ForceNew/computed/sensitive/defaults/validation).

**Phase 1 — storage account to production-clean**
- [ ] Confirm idempotent empty re-plan end-to-end (acceptance green).
- [ ] Fix shared-node ForceNew over-application (path-aware emission).
- [ ] Decide handling for additionalProperties maps (`tags`, identities): typed
      `MapAttribute` vs dynamic; implement.
- [ ] Generated docs (registry docs) from attribute descriptions.

**Phase 2 — scale generation**
- [ ] `go generate` target producing the top-N resources (per-service sub-package layout already in place).
- [ ] Per-resource schema-validity + bicep-cross-check in CI.
- [ ] Naming override table for the 9 collisions; finalize naming policy.
- [ ] Compile-time/binary-size budget; sub-packaging strategy.

**Phase 3 — coverage & hardening**
- [ ] Expand azwise knowledge for high-priority services (compute, network,
      containerservice, etc.).
- [ ] Acceptance suites per shipped resource; CI gating on `TF_ACC` lanes.
- [ ] Discriminated-union strategy (flatten simple unions vs dynamic fallback).
- [ ] Plan-time default consumption / drift hardening.

**Phase 4 — release**
- [ ] Provider feature flag to enable/disable static resources.
- [ ] Migration guidance `azapi_resource` ↔ static resource.
- [ ] Versioning policy for API-version pinning per resource.

### 6.4 Tooling commands

```sh
# regenerate the storage account schema (PoC generator)
go run ./internal/native/generator/cmd/generate_poc.go

# cross-validate a compiled schema against bicep
go run ./internal/native/cmd/azapin-validate/ -r azapi_storage_account

# offline unit tests (skips acceptance)
TF_ACC= go test ./internal/native/... ./internal/azure/azwise/...

# acceptance (creates real Azure resources)
TF_ACC=1 go test ./internal/native/services/storage/ -run TestStorageAcceptance \
  -ginkgo.focus "creates a basic account and imports it" -timeout 900s

# build the dev provider for manual workspaces
go build -o "$(go env GOPATH)/bin/terraform-provider-azapi" .
```

---

## Appendix A — Generator rule reference

See `internal/native/GENERATOR.md` for the authoritative generator rules (flag derivation,
block-level computed inference, defaults, single-optional promotion, azwise overlay,
validators, naming, runtime mapping, scope selection) plus the recursion/
discriminated-type and dynamic-in-collection handling.

## Appendix B — azapi vs azapin

See `tools/azwise/azapin-vs-azapi.md` for the stakeholder-facing feature comparison.
