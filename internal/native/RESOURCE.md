# Azapin Static Resource — Design

This document designs the Terraform `resource.Resource` implementation for azapin's
generated static schemas. It precedes any implementation. It builds on:

- **DEVELOPER_SPEC.md** — architecture, resource/property naming, design decisions, project scope
- **GENERATOR.md** — schema emission rules (flags, defaults, validators, naming)

The generated schema (e.g. `services/storage`'s `AzapiStorageAccountSchema()`,
reachable layout-agnostically as `services.Registry["azapi_storage_account"].Schema()`)
describes the
**ARM request body** as typed Terraform attributes. This document describes the
runtime that turns those schemas into working resources: a shared base that
implements CRUD against ARM, plus per-resource customization seams.

## Goals

1. **One shared base** implements the framework interface methods (Create/Read/
   Update/Delete) and a single unified payload-composition path that converts the
   typed Terraform config into the ARM HTTP request body — and the ARM response
   back into typed state.
2. **Each resource customizes the unified CRUD** with its own logic (tweak the
   body before PUT, post-process state after GET, special-case a field) without
   reimplementing the whole method.
3. **Customizable validate / modify-plan / import** per resource.

## Non-Goals

- Replacing `azapi_resource`. The dynamic-body resource stays as the day-zero /
  preview / escape-hatch path. Static resources coexist.
- Generating ~2,631 Go model structs. The base maps state ↔ ARM JSON **generically**,
  driven by the schema and the embedded bicep type graph — no per-resource struct.
- Polymorphic (discriminated) bodies. Those fall back to `azapi_resource`.

## Where It Fits

Static resources reuse the existing provider plumbing — no new client, auth, or
polling:

| Concern | Reused from | Symbol |
|---|---|---|
| Provider data | `internal/clients` | `*clients.Client` (`.ResourceClient`, `.Features`, `.Account`) |
| PUT + poll | `internal/clients` | `ResourceClient.CreateOrUpdate(ctx, resourceID, apiVersion, body, opts)` |
| GET | `internal/clients` | `ResourceClient.Get(ctx, resourceID, apiVersion, opts)` |
| DELETE + poll | `internal/clients` | `ResourceClient.Delete(ctx, resourceID, apiVersion, opts)` |
| Resource ID | `internal/services/parse` | `NewResourceID(name, parentID, "type@version")`, `ResourceIDWithResourceType(id, type)` |
| Op knowledge | `internal/azure/azwise` | `CheckForceNew`, `Validate`, `StripComputedFields` (timeouts are baked into `services.Descriptor.Timeouts` at gen time) |
| ARM type graph | `internal/azure` / `internal/native/generator` | `azure.GetResourceDefinition` / `generator.ParseTypesJSON` |
| Registry | `internal/native/services` | `Registry map[string]SchemaFunc` |

Provider registration appends to the existing list in `provider.go`:

```go
func (p Provider) Resources(ctx context.Context) []func() resource.Resource {
    list := []func() resource.Resource{
        func() resource.Resource { return &services.AzapiResource{} },
        // … existing …
    }
    for name := range services.Registry {           // native static resources
        name := name
        list = append(list, func() resource.Resource { return nativeresource.New(name) })
    }
    return list
}
```

## Composed Schema

The generated schema covers only the ARM **body** properties (`location`, `tags`,
`sku`, `kind`, `identity`, `properties`, `extendedLocation`, …). The bicep
`SystemManaged` fields (`id`, `name`, `type`, `apiVersion`) are excluded by the
generator (GENERATOR.md Rule 5).

The **generated** schema already includes the operational envelope: the generator
synthesizes `name`, the parent reference, and `id` directly into the generated
file (GENERATOR.md Rule 9c). The runtime base adds only the `timeouts` block — it
needs a `context.Context` the static generated function cannot hold — and serves
the rest verbatim. The schema is therefore never mutated at runtime:

```
final schema = generated body attributes
             + generated envelope attributes:
                 "name"               (Required, RequiresReplace, optional name validator)
                 "<parent>_id"        (Required, RequiresReplace, resource-id validator)
                 "id"                 (Computed — the ARM resource ID)
             + runtime-added block:
                 "timeouts"           (create/read/update/delete block)
```

The parent reference takes a resource-specific name and validator derived from the
ARM scope (e.g. `resource_group_id`, `subscription_id`, `storage_account_id`,
falling back to `parent_id`); the chosen name is recorded in `Descriptor.ParentAttr`
so the runtime can compose/parse the ARM ID. `location`, `tags`, `identity` are
already typed body attributes from the bicep graph, so the user sets them natively:

```hcl
resource "azapi_storage_account" "example" {
  name              = "mystorage"
  resource_group_id = azurerm_resource_group.example.id

  location = "westus2"
  kind     = "StorageV2"
  sku { name = "Standard_LRS" }
  properties {
    access_tier         = "Hot"
    minimum_tls_version = "TLS1_2"
  }
}
```

A later phase may add the `azapi_resource` operational knobs (`retry`,
`ignore_casing`, `*_headers`, `response_export_values`) behind the same envelope;
they are out of scope for v1.

## Resource Descriptor

Each generated resource is described by a small data value. The generator emits one
per resource (alongside the schema func) and registers it:

```go
// In package services (emitted; each generated file calls Register in its init())
type Descriptor struct {
    Name           string               // "azapi_storage_account"
    ARMType        string               // "Microsoft.Storage/storageAccounts"
    APIVersion     string               // "2025-06-01" (latest stable, from index.json)
    Schema         func() schema.Schema // generated schema: body + operational envelope
    WritableScopes int                  // bicep scope bitmask (Tenant=1…ResourceGroup=8); metadata
    ParentAttr     string               // generated parent-reference attribute, e.g. "resource_group_id"
}

var Registry = map[string]Descriptor{}
func Register(d Descriptor) { Registry[d.Name] = d }
```

The descriptor is deliberately dependency-light (schema only — no framework/client
imports), so generated files stay pure. The ARM type + version replace the brittle
`[azapin:…]` description-tag parsing the validator uses; the tag stays for the
validator's convenience but the descriptor is the source of truth at runtime.

Runtime **behavior** hooks are **not** on the descriptor. `internal/native/resource`
owns the hook types and `hookRegistry`; hand-written `<resource>_hooks.go` files live
beside the generated resource schema in `internal/native/services/<service>/` and
register via `resource.RegisterHooks`. `Base` looks hooks up once at construction
(`hookRegistry[name]`); a resource with no hooks simply uses the base behavior.

## Base Struct

```go
// package resource
type Base struct {
    desc      services.Descriptor
    hooks     *Hooks               // hookRegistry[name]; nil = base behavior only
    provider  *clients.Client      // set in Configure
    typeGraph *generator.Type      // lazily resolved ARM body type graph (cached)
}

func New(name string) resource.Resource { return &Base{desc: services.Lookup(name)} }
```

`Base` implements the full framework method set; every generated resource uses it
directly (`nativeresource.New(name)`), so there is exactly one implementation to maintain:

| Interface | Method | Behavior |
|---|---|---|
| `resource.Resource` | `Metadata` | `TypeName = desc.Name` |
| | `Schema` | `composeSchema` serves `desc.Schema()` (body + generated envelope) + adds the `timeouts` block |
| | `Create` | unified `createUpdate(isNew=true)` |
| | `Read` | unified `read` |
| | `Update` | unified `createUpdate(isNew=false)` |
| | `Delete` | unified `delete` |
| `…WithConfigure` | `Configure` | store `*clients.Client` |
| `…WithModifyPlan` | `ModifyPlan` | ForceNew (azwise + schema), default/computed handling, hook |
| `…WithValidateConfig` | `ValidateConfig` | schema validators run automatically; hook for cross-field rules |
| `…WithImportState` | `ImportState` | parse ARM ID → GET → flatten into typed state, hook |

## Unified CRUD Flow

All four operations funnel through two private methods on `Base`. The mapper
(below) is the single payload-composition path required by goal #1.

### createUpdate(ctx, plan, state, isNew)

```
1. id := parse.NewResourceID(plan.name, plan.<parent>_id, ARMType@APIVersion)
2. timeout := b.timeout("create"|"update", default)   ← baked into services.Descriptor at gen time
3. if isNew && !Singleton: client.Get(id) → "already exists" error   ← pre-create existence check
4. armBody := mapper.Expand(plan.bodyObject, typeGraph)   ← unified composition
5. azwise.StripComputedFields(armBody)                    ← drop server-computed read-only fields
6. hook.BeforeCreate / BeforeUpdate (optional)           ← Body live; mutations reach ARM
7. client.CreateOrUpdate(ctx, id.AzureResourceId, id.ApiVersion, armBody, opts)  ← PUT
8. getResp := client.Get(...)                            ← read-back
9. hook.AfterCreate / AfterUpdate (optional)             ← Response live; runs before flatten
10. newState := mapper.FlattenApplyInto(getResp, plan, typeGraph); newState.id = id.ID(); ResolveUnknowns(newState)
11. resp.State.Set(newState)
```

`Create` and `Update` are thin wrappers selecting `isNew`. Pre-existence check on
create (`Get` → "already exists") mirrors `azapi_resource`.

### read(ctx, state)

```
1. id := parse.ResourceIDWithResourceType(state.id, ARMType@APIVersion)
2. timeout := b.timeout("read", 5m)   ← from services.Descriptor.Timeouts
3. hook.BeforeRead (optional)   ← State live; append a diag to short-circuit before the GET
4. getResp := client.Get(...)   // 404 → RemoveResource
5. hook.AfterRead (optional)    ← Response live; massage before flatten (e.g. location casing)
6. newState := mapper.FlattenInto(getResp, priorState, typeGraph)
7. resp.State.Set(newState)
```

### delete(ctx, state)

```
1. id := parse.ResourceIDWithResourceType(state.id, ARMType@APIVersion)
2. timeout := b.timeout("delete", 30m)   ← from services.Descriptor.Timeouts
3. if hooks.Singleton: client.CreateOrUpdate(..., Singleton.DefaultBody) → return  // fixed-name default, no ARM delete: reset to baseline; BeforeDelete is NOT run
4. hook.BeforeDelete (optional)  ← State live; not reached on the Singleton reset path
5. if hooks.Delete: hook.Delete(...) else client.Delete(...)  // custom delete replaces ARM DELETE; 404 tolerated on default path
6. hook.AfterDelete (optional)   ← State live; not reached on the Singleton reset path
```

## The Mapper (state ↔ ARM JSON)

This is the heart of the design and the single composition path. It is **generic**
— driven by the schema and the bicep type graph — so no per-resource Go struct is
needed.

### Expand: typed plan → ARM JSON (`map[string]interface{}`)

Walk the bicep body type graph. For each property:

- Resolve the Terraform attribute by `naming.CamelToSnake(armName)`.
- Read its value from the plan's `types.Object` tree.
- Skip `null` / `unknown`. Skip server-computed-only fields the user did not set.
- Convert by kind: string/enum→string, bool→bool, int→float64(JSON), object→nested
  map (recurse), array-of-object→`[]interface{}` (recurse), array-of-scalar→slice.
- **Key the output with the ARM camelCase name taken from the type graph node**,
  never by reversing snake_case (avoids `ip_rules`→`ipRules`/`iPRules` ambiguity).

### Flatten: ARM JSON response → typed state (`types.Object`)

Walk the bicep body type graph (or the schema). For each ARM key present in the
response:

- Resolve the Terraform attribute name (`CamelToSnake`).
- Convert the JSON value to the `attr.Value` whose type matches the schema attribute.
- Assemble the nested `types.Object` / `types.List` to set as state.

Properties absent from the response keep their planned value (Optional+Computed) or
become null/known-after-apply per the schema flags. Computed-only fields are
populated from the response.

### Naming authority (open decision)

The mapper needs the exact ARM name for every snake attribute. Two options:

- **A — runtime type graph (honors GENERATOR.md Rule 15).** The base resolves the
  bicep type graph for `ARMType@APIVersion` once (cached) via the embedded
  `internal/azure/generated` data and reads ARM names from it. Zero extra codegen.
  *Recommended.*
- **B — emitted name map.** The generator emits a `snake→ARM` map per resource.
  Simple lookups, but reintroduces the ~215-entries-per-resource duplication Rule 15
  removed.

Recommendation: **A**, reusing `generator.ParseTypesJSON` + `PostProcess` (already
cycle-safe and discriminated-safe after the recent fixes), cached per
`ARMType@APIVersion`. The azapin validator already proves the schema and type graph
agree, so the mapper's two walks stay in lockstep.

## Customization Seams

Two layered mechanisms, in order of preference:

### 1. Hooks (data-driven, primary)

Hooks customize runtime **behavior** only. Schema customization — attribute
validators, defaults, the parent-reference name — is baked into the generated
schema at generation time via generator customizers (GENERATOR.md Rule 9d), so the
runtime never mutates the schema.

A hand-written `<resource>_hooks.go` file in the generated service package registers
`Hooks` via `resource.RegisterHooks`. Hooks receive a `*CrudCtx` carrying everything they might touch:

```go
type CrudCtx struct {
    Ctx      context.Context
    Client   *clients.Client
    ID       parse.ResourceId
    Plan     types.Object           // typed plan (create/update); null on read/delete
    State    types.Object           // typed prior state (read/delete); null on create/update
    Body     map[string]interface{} // ARM body being composed — mutate in Before* create/update
    Response map[string]interface{} // ARM GET response — read in After*
    Diags    *diag.Diagnostics
}

type Hooks struct {
    BeforeCreate, AfterCreate func(*CrudCtx)
    BeforeUpdate, AfterUpdate func(*CrudCtx)
    BeforeRead,   AfterRead   func(*CrudCtx)
    BeforeDelete              func(*CrudCtx)
    Delete                    func(*CrudCtx) // optional replacement for default ARM DELETE
    AfterDelete               func(*CrudCtx)
    Singleton                 *SingletonDefault // fixed-named default child (reset-on-destroy)
    Relational                []services.RelationalConstraint // declarative cross-property constraints -> framework ConfigValidators
    ConfigValidators          []resource.ConfigValidator      // prebuilt/reusable validator objects, appended after Relational
    ValidateConfig            func(context.Context, resource.ValidateConfigRequest, *resource.ValidateConfigResponse)
    ModifyPlan                func(context.Context, resource.ModifyPlanRequest, *resource.ModifyPlanResponse)
}
```

The **authoritative per-hook-point contract** — which of `Plan`/`State`/`Body`/
`Response` are live vs null at each hook, whether a `Body` mutation still reaches
ARM, the lifecycle position, and the `Singleton` suppression rules — lives in the
`Hooks`/`CrudCtx` doc comments in `internal/native/resource/hooks.go`, guarded
against drift from `Base` by `base_hook_contract_test.go`. Consult the seam; the
flows above are a summary of it.

### 2. Method override (code-driven, escape hatch)

For wholesale-different behavior, a resource defines its own type embedding the base
and shadows a method:

```go
type StorageAccountResource struct{ *nativeresource.Base }
func (r *StorageAccountResource) Create(ctx, req, resp) { /* bespoke */ }
func newStorageAccount() resource.Resource {
    return &StorageAccountResource{Base: nativeresource.New("azapi_storage_account")}
}
```

Go promotes the base methods; the override wins via the outer type's method set.
The provider registers the constructor for resources that need this. Hooks cover
the common cases; this is reserved for the rare resource that needs to bypass the
unified flow entirely.

## ForceNew, Computed, Defaults

- **ForceNew**: primarily schema-level (`RequiresReplace` plan modifiers on the
  attribute, emitted by the generator from a future override list). `ModifyPlan`
  additionally consults `azwise.CheckForceNew(ARMType, APIVersion, oldBody, newBody)`
  for conditional rules the schema can't express (e.g. storage SKU zone migration),
  expanding old/new state via the mapper.
- **Computed / defaults**: already encoded in the generated schema (GENERATOR.md
  Rules 3, 6, 8). `Optional+Computed` fields read their server value back via
  Flatten; `Default(...)` fields get plan-time defaults natively.
- **Computed-field diff suppression**: not needed — the typed schema marks
  read-only fields `Computed`, so Terraform never diffs user config against them.
  (`azwise.StripComputedFields` is an `azapi_resource`-only concern.)

## Envelope & ID Details

- `name` + the parent reference build the ARM ID via `parse.NewResourceID` (the
  runtime reads the parent attribute name from `Descriptor.ParentAttr`). Both
  `RequiresReplace`.
- `id` (computed) = `id.ID()` (the ARM resource path).
- `location`/`tags`/`identity` flow through the typed body — `Flatten` normalizes
  `location` casing in `AfterRead` (shared hook in the base, not per-resource).
- `ImportState`: parse the ARM ID (+type), GET, Flatten into state, set `id`.

## Data Source

Every native static resource also gets a **data source** of the same name (e.g.
`data.azapi_storage_account`), for reading an existing resource into state. There
is no second generated schema and no per-resource data-source code: one generic
runtime (`internal/native/resource/datasource.go`) serves them all, mirroring the
resource `Base`.

- **Schema (reused, converted at runtime).** The framework keeps resource and
  data-source schemas in distinct packages (whose shared `Attribute` interface is
  in an un-importable `internal` package), so `schema_convert.go` type-switches on
  the public concrete `resource/schema.*Attribute` types and rebuilds the
  `datasource/schema` equivalents. `name` and the parent reference
  (`Descriptor.ParentAttr`) become **Required** inputs the practitioner supplies to
  locate the resource; every other attribute — the whole ARM body plus `id` —
  becomes a **Computed** output. Plan modifiers and defaults are dropped; element
  types and descriptions are preserved, and string validators on the identity
  inputs are carried over. The runtime adds only the read-only `timeouts` block.
- **Read (reuses the resource read path).** `name` + `<parent>_id` compose the ARM
  ID via `parse.NewResourceID`; a GET (404 → error, since a data source must find
  its target) feeds `mapper.FlattenInto` over the config object, so the identity
  inputs and the `timeouts` block are preserved while the computed body is populated
  from the response; `id`/`name`/`<parent>_id` are set from the parsed ID.
- **Registration.** `Provider.DataSources()` appends `nativeresource.NewDataSource(name)`
  for every `services.Registry` entry, exactly as `Resources()` does for
  `New(name)`.

A data source is read-only, so it runs no hooks; the lookup is by `name` +
`<parent>_id` only (no `resource_id` alternate path in v1).

## Phasing

1. **Mapper** (`expand`/`flatten`) + unit tests round-tripping storage_account
   against recorded ARM JSON. Reuse cycle-safe `generator` type graph.
2. **Base** implementing Resource/Configure/Schema/CRUD with no hooks; wire one
   resource (`azapi_storage_account`) end-to-end behind a feature flag.
3. **ModifyPlan/ValidateConfig/ImportState** + azwise ForceNew integration.
4. **Hooks** + one generated-service hook file (storage_account SKU zone-migration
   ForceNew) to prove the seam.
5. **Provider registration** of the full `services.Registry`; acceptance test a
   handful of representative resources.

### v1 status (implemented)

- **Mapper** — `internal/native/mapper`: `Expand`, `Flatten`, `FlattenInto`, driven
  by the bicep type graph; round-trip tested on storage_account (`mapper_test.go`).
- **Base** — `internal/native/resource/base.go`: implements `Resource`,
  `ResourceWithConfigure/ModifyPlan/ValidateConfig/ImportState`; serves the
  generated schema (body + envelope) plus the runtime `timeouts` block, validated by
  `Schema.ValidateImplementation` in tests; unified
  `put`/`Read`/`Delete`; runtime body loader reads the authoritative types.json from
  `azure.StaticFiles` (`loader.go`).
- **Hooks** — `hooks.go`: runtime behavior only — `Before/After` per op +
  `ValidateConfig`/`ModifyPlan` overrides; generated-service hook files such as
  `services/storage/storage_account_hooks.go` implement resource-specific runtime
  behavior like the SKU zone-migration ForceNew via `azwise.CheckForceNew`.
- **Schema customizers** — `internal/native/generator/customizers`: generation-time, per-ARM-type
  Go hooks that bake validators/defaults/parent-name into the generated schema
  (see GENERATOR.md Rule 9d). Not part of the provider runtime.
- **Generated descriptor** — `services.Descriptor{Name, ARMType, APIVersion, Schema, WritableScopes, ParentAttr, Timeouts}`
  registered via each generated file's `init()`. `Timeouts` carries the per-op
  timeout defaults baked in from azwise at gen time (the runtime reads them, not the
  azwise registry). Resource-level cross-property constraints are **not** in the
  descriptor — they are hand-declared in `Hooks.Relational` (see below), keeping the
  generated `_gen.go` minimal and letting them be tuned without regeneration.
- **Provider** — `Resources()` appends `nativeresource.New(name)` and
  `DataSources()` appends `nativeresource.NewDataSource(name)` for every
  `services.Registry` entry.
- Exercised against live ARM via the Ginkgo acceptance suites
  (`services/<service>/<resource>_test.go`); unit and schema-validation coverage in
  place.

## Open Questions

1. **Generic state I/O.** Implement expand/flatten over `tftypes.Value`
   (`req.Plan.Raw`) or over `types.Object`? `tftypes.Value` is lower-level but
   avoids per-attribute path juggling; lean that way and confirm in phase 1.
2. **Operational knobs.** Do v1 static resources need `retry` / `ignore_casing` /
   `response_export_values`? Proposed: no — keep the surface minimal; revisit once
   real configs demand it.
3. **Update semantics.** ARM uses PUT for create and update alike; confirm whether
   any generated resource needs PATCH (rare) and, if so, expose it via a hook.
