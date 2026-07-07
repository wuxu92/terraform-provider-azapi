# terraform-provider-azapi

Glossary for the AzAPI Terraform provider, with emphasis on **azapin** — the static, typed, per-resource generation effort layered beside the existing dynamic `azapi_resource`.

## Language

### Resource surfaces

**azapi_resource**:
The dynamic, day-zero resource: one freeform `body` object accepting raw ARM JSON for any type/API version. The escape hatch — unchanged by azapin, always available.
_Avoid_: generic resource, dynamic resource (as a proper noun)

**Azapin**:
The subsystem (and its generated output) that turns a single ARM resource type into a typed, per-resource Terraform resource (`azapi_<service>_<resource>`) generated from Azure's bicep types plus the azwise overlay. Ships in the same provider binary and coexists with `azapi_resource`.
_Avoid_: native provider, static provider

**Static resource** / **native resource**:
An individual azapin-generated resource (e.g. `azapi_storage_account`). "Static" = typed schema fixed at build time; "native" = a first-class provider resource rather than a dynamic body. The two terms are used interchangeably; prefer **static resource** in user-facing text and **native** for the internal package (`internal/native`).
_Avoid_: typed resource, schema resource

### Generation

**Azwise**:
The curated, AzureRM-derived operational knowledge (ForceNew, real defaults, validation, sensitive/computed classification, timeouts) that sharpens the schema and lifecycle. Applied at **two times**: generation time (`ApplyAzwise`, baked into the compiled schema flags/validators) and runtime (`StripComputedFields`, `TimeoutDefault`, `CheckForceNew`). For `azapi_resource` it is toggleable via the `disable_resource_knowledge` feature; for azapin it is compiled in and **not** toggleable (see ADR-0005).
_Avoid_: knowledge base, overrides

**Overlay**:
The generation-time application of azwise knowledge onto the raw bicep-derived type graph (`ApplyAzwise`). An overlay _augments_ the mechanical schema; it never replaces the type graph.

**Type graph**:
The bicep-derived in-memory model of an ARM resource body — `Type`/`Property`/`Kind*` plus `ParseTypesJSON`/`PostProcess`/`ApplyAzwise` and the model helpers. Lives in its own module (`internal/native/typegraph`), separate from the code-emission engine (`internal/native/generator`), and is the only azapin package the runtime CRUD path imports (see ADR-0007). Both the runtime and the generator read it; only the generator emits from it.
_Avoid_: schema model, AST, generator types

**Generation pipeline**:
The three named entry points that own the ordered generation stages (see ADR-0007): `typegraph.BuildRuntimeGraph` (parse + post-process, no customizers — the graph the runtime reads), `generator.BuildForGeneration` (adds `customizers.Apply` — the graph baked into `_gen.go`), and `generator.Generate` (adds invariants → emit → verify). Replaces the ordering that was formerly hand-spelled at five call sites.
_Avoid_: codegen flow, build steps

**Schema verification**:
Checking that a generated schema faithfully covers the bicep type graph and vice versa. Two adapters over one shared `Mismatch` core (see ADR-0008): the emitted-**source** adapter (`ValidateEmittedSchema`, presence of attribute names, gates codegen pre-write) and the compiled-**schema** adapter (`SchemaAgainstBicep`, type-mapping check post-init). The middle step of `Graduation`.
_Avoid_: schema validation, schema diff

**Operational envelope**:
The synthesized top-level attributes every static resource needs beyond its ARM body: `name`, the parent reference (`parent_id` / typed parent attr), and `id`. Added during post-processing, not present in the bicep body type.
_Avoid_: wrapper, metadata fields

### Binding

**API-version pin**:
The single ARM API version (latest stable) baked into a static resource at generation time. The typed schema, its validators, and runtime ARM-name resolution all correspond to exactly this version; there is no user-settable `api_version`. Advancing the pin is a governed regeneration event, not a runtime choice.
_Avoid_: version selector

**Safe default**:
The generator's mechanical flag for a body property that is neither `Required` nor `ReadOnly` and has no curated default: `Optional + Computed` with `UseStateForUnknown`. Prevents perpetual diffs on server-populated fields at the cost of unset-by-omission.
_Avoid_: default flags, optional-computed (as a noun)

**Unset-by-omission**:
Clearing a resource attribute by deleting it from HCL and expecting the value to be removed on the next apply. Static resources do **not** support this for `Safe default` (Optional+Computed) attributes — the prior value persists via state reuse; azwise downgrades known user-owned fields to plain `Optional` to restore it.
_Avoid_: unset, clear-by-delete

### Extension model

**Hook**:
A per-resource-name callback registered via `RegisterHooks` that runs inside the generic `Base` lifecycle: `BeforeCreate`/`AfterCreate`/`BeforeUpdate`/`AfterUpdate`/`BeforeRead`/`AfterRead`/`BeforeDelete`, plus `ValidateConfig` and `ModifyPlan`. The data-driven seam for per-resource logic the mechanical schema can't express (conditional ForceNew, cross-field validation, extra API calls). `BeforeRead` fires before the read GET (`State` live, `Response` null) and `AfterRead` after it — both wired in `Base.Read`. The authoritative per-hook-point contract (which `CrudCtx` fields are live vs null, whether a `Body` mutation reaches ARM, lifecycle position, `Singleton` suppression) is the doc comment on `Hooks`/`CrudCtx` in `internal/native/resource/hooks.go`, guarded against drift from `Base` by `base_hook_contract_test.go`.
_Avoid_: callback, plugin, middleware

**Singleton default**:
A static resource that ARM neither creates nor deletes on its own — it exists as a fixed-named default child of its parent (e.g. `storageAccounts/blobServices/default`). Create is really an in-place PUT; Delete resets it to a baseline body instead of issuing an ARM DELETE (and does not run `BeforeDelete`). Marked by a non-nil `Hooks.Singleton`.
_Avoid_: default resource, implicit resource

**Bundling**:
Folding several ARM sub-APIs into one static resource as nested attributes via hooks (AzureRM's `blob_properties` pattern). A deliberate per-resource opt-in, never the default — the default is one ARM type per resource.
_Avoid_: aggregation, composite resource

### Typing gaps

**Customizer**:
A per-resource, generation-time Go rule (`generator/customizers/<resource>.go`, keyed by ARM type) that gets the final say over the schema after the mechanical pipeline and azwise overlay. The hand-written escape hatch for schema shape the mechanical rules get wrong.
_Avoid_: override, patch, tweak

**Discriminated union**:
A polymorphic ARM body (bicep `DiscriminatedObjectType`) whose shape depends on a discriminator field. Emitted mechanically as a `Dynamic fallback`; a customizer may opt a small, stable, high-value variant set into partial typing (typically typing the discriminator enum while keeping the variant body dynamic).
_Avoid_: polymorphic type, oneOf

**Dynamic fallback**:
Emitting a body subtree as a Terraform `Dynamic`/`DynamicAttribute` (from bicep `KindAny`) when it can't be statically typed — discriminated unions and `KindAny`-valued open maps (e.g. user-assigned identities). The typed-schema benefits (autocomplete, per-attribute validation) do not apply inside a dynamic subtree. Note: string-valued open maps like `tags` are fully typed as `MapAttribute` and are **not** a dynamic fallback.
_Avoid_: any-type, untyped blob

### Rollout

**Allowlist**:
The explicit, curated set of ARM types the generator emits static resources for (the generator's `targets`). Adding a resource is a deliberate opt-in, never a bulk sweep; the untyped long tail is served by `azapi_resource`.
_Avoid_: coverage list, catalog

**Graduation**:
The gate a candidate resource must clear before it ships as a static resource: framework flag-invariants + a live-Azure acceptance test + an authored azwise overlay. The live-API schema verification harness produces the evidence for the middle step.
_Avoid_: promotion, release gate
