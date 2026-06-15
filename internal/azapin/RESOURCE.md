# Azapin Static Resource — Design

This document designs the Terraform `resource.Resource` implementation for azapin's
generated static schemas. It precedes any implementation. It builds on:

- **DESIGN.md** — generator architecture and resource naming
- **GENERATOR.md** — schema emission rules (flags, defaults, validators, naming)

The generated schema (`generated.AzapiStorageAccountSchema()` etc.) describes the
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
| Op knowledge | `internal/azure/azwise` | `TimeoutDefault`, `CheckForceNew`, `Validate`, `StripComputedFields` |
| ARM type graph | `internal/azure` / `internal/azapin/generator` | `azure.GetResourceDefinition` / `generator.ParseTypesJSON` |
| Registry | `internal/azapin/generated` | `Registry map[string]SchemaFunc` |

Provider registration appends to the existing list in `provider.go`:

```go
func (p Provider) Resources(ctx context.Context) []func() resource.Resource {
    list := []func() resource.Resource{
        func() resource.Resource { return &services.AzapiResource{} },
        // … existing …
    }
    for name := range generated.Registry {           // azapin static resources
        name := name
        list = append(list, func() resource.Resource { return azapin.New(name) })
    }
    return list
}
```

## Composed Schema

The generated schema covers only the ARM **body** properties (`location`, `tags`,
`sku`, `kind`, `identity`, `properties`, `extendedLocation`, …). The bicep
`SystemManaged` fields (`id`, `name`, `type`, `apiVersion`) are excluded by the
generator (GENERATOR.md Rule 5).

The base **wraps** the generated body schema with the operational envelope every
resource needs, then hands the union to Terraform:

```
final schema = generated body attributes
             + base envelope attributes:
                 "name"       (Optional+Computed, RequiresReplace)
                 "parent_id"  (Optional+Computed, RequiresReplace, resource-id validator)
                 "id"         (Computed — the ARM resource ID)
                 "timeouts"   (create/read/update/delete block)
```

`location`, `tags`, `identity` are NOT re-added — they are already typed body
attributes from the bicep graph, so the user sets them natively:

```hcl
resource "azapi_storage_account" "example" {
  name      = "mystorage"
  parent_id = azurerm_resource_group.example.id

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
// In package generated (emitted)
type Descriptor struct {
    Name       string                  // "azapi_storage_account"
    ARMType    string                  // "Microsoft.Storage/storageAccounts"
    APIVersion string                  // "2025-01-01"
    Schema     func() schema.Schema    // generated body schema
    // Hooks is nil for fully-generated resources; hand-written overlay files
    // populate it for resources that need custom behavior.
    Hooks      *Hooks
}

func Register(d Descriptor) { registry[d.Name] = d }
```

The ARM type + version replace the brittle `[azapin:…]` description-tag parsing
the validator currently uses; the tag stays for the validator's convenience but
the descriptor is the source of truth at runtime.

## Base Struct

```go
// package azapin
type Base struct {
    desc     generated.Descriptor
    provider *clients.Client      // set in Configure
    typeGraph *generator.Type     // lazily resolved ARM body type graph (cached)
}

func New(name string) resource.Resource { return &Base{desc: generated.Lookup(name)} }
```

`Base` implements the full framework method set; every generated resource uses it
directly (`azapin.New(name)`), so there is exactly one implementation to maintain:

| Interface | Method | Behavior |
|---|---|---|
| `resource.Resource` | `Metadata` | `TypeName = desc.Name` |
| | `Schema` | `composeSchema(desc.Schema())` (body + envelope) |
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
1. id := parse.NewResourceID(plan.name, plan.parent_id, ARMType@APIVersion)
2. timeout := azwise.TimeoutDefault(ARMType, APIVersion, "create"|"update", default)
3. hook.BeforeCreate / BeforeUpdate (optional)           ← customization
4. armBody := mapper.Expand(plan.bodyObject, typeGraph)   ← unified composition
5. resp := client.CreateOrUpdate(ctx, id.AzureResourceId, id.ApiVersion, armBody, opts)
6. getResp := client.Get(...)                             ← read-back
7. newState := mapper.Flatten(getResp, schema, typeGraph) ← unified composition
8. newState.id = id.ID(); newState.name/parent_id carried from plan
9. hook.AfterCreate / AfterUpdate (optional)              ← customization
10. resp.State.Set(newState)
```

`Create` and `Update` are thin wrappers selecting `isNew`. Pre-existence check on
create (`Get` → "already exists") mirrors `azapi_resource`.

### read(ctx, state)

```
1. id := parse.ResourceIDWithResourceType(state.id, ARMType@APIVersion)
2. timeout := azwise.TimeoutDefault(..., "read", 5m)
3. hook.BeforeRead (optional)
4. getResp := client.Get(...)   // 404 → RemoveResource
5. newState := mapper.Flatten(getResp, schema, typeGraph)
6. hook.AfterRead (optional)    ← e.g. normalize location casing, drop write-only echoes
7. resp.State.Set(newState)
```

### delete(ctx, state)

```
1. id := parse.ResourceIDWithResourceType(state.id, ARMType@APIVersion)
2. timeout := azwise.TimeoutDefault(..., "delete", 30m)
3. hook.BeforeDelete (optional)
4. client.Delete(...)           // 404 tolerated
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

A hand-written overlay file sets `Descriptor.Hooks`. Hooks receive a `*CrudCtx`
carrying everything they might touch and return diagnostics:

```go
type CrudCtx struct {
    Ctx       context.Context
    Client    *clients.Client
    ID        parse.ResourceId
    Plan      *types.Object   // typed plan (create/update)
    State     *types.Object   // typed prior state
    Body      map[string]any  // the ARM body being composed (mutate before PUT)
    Response  map[string]any  // the GET response (read/after-create)
    Diags     *diag.Diagnostics
}

type Hooks struct {
    BeforeCreate, AfterCreate func(*CrudCtx)
    BeforeUpdate, AfterUpdate func(*CrudCtx)
    BeforeRead,   AfterRead   func(*CrudCtx)
    BeforeDelete              func(*CrudCtx)
    ValidateConfig            func(context.Context, resource.ValidateConfigRequest, *resource.ValidateConfigResponse)
    ModifyPlan                func(context.Context, resource.ModifyPlanRequest, *resource.ModifyPlanResponse)
    ImportState               func(context.Context, resource.ImportStateRequest, *resource.ImportStateResponse)
    Expand                    func(*CrudCtx) (map[string]any, bool) // full override of Expand; ok=false → default
    Flatten                   func(*CrudCtx) (*types.Object, bool)  // full override of Flatten
}
```

This satisfies goals #2 and #3: a resource can tweak the body (`BeforeCreate`),
post-process state (`AfterRead`), or fully replace expand/flatten — without
touching the base.

### 2. Method override (code-driven, escape hatch)

For wholesale-different behavior, a resource defines its own type embedding the base
and shadows a method:

```go
type StorageAccountResource struct{ *azapin.Base }
func (r *StorageAccountResource) Create(ctx, req, resp) { /* bespoke */ }
func newStorageAccount() resource.Resource {
    return &StorageAccountResource{Base: azapin.New("azapi_storage_account")}
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

- `name` + `parent_id` build the ARM ID via `parse.NewResourceID`. Both
  `RequiresReplace`.
- `id` (computed) = `id.ID()` (the ARM resource path).
- `location`/`tags`/`identity` flow through the typed body — `Flatten` normalizes
  `location` casing in `AfterRead` (shared hook in the base, not per-resource).
- `ImportState`: parse the ARM ID (+type), GET, Flatten into state, set `id`.

## Phasing

1. **Mapper** (`expand`/`flatten`) + unit tests round-tripping storage_account
   against recorded ARM JSON. Reuse cycle-safe `generator` type graph.
2. **Base** implementing Resource/Configure/Schema/CRUD with no hooks; wire one
   resource (`azapi_storage_account`) end-to-end behind a feature flag.
3. **ModifyPlan/ValidateConfig/ImportState** + azwise ForceNew integration.
4. **Hooks** + one overlay (storage_account SKU zone-migration ForceNew) to prove
   the seam.
5. **Provider registration** of the full `generated.Registry`; acceptance test a
   handful of representative resources.

### v1 status (implemented)

- **Mapper** — `internal/azapin/mapper`: `Expand`, `Flatten`, `FlattenInto`, driven
  by the bicep type graph; round-trip tested on storage_account (`mapper_test.go`).
- **Base** — `internal/azapin/resource/base.go`: implements `Resource`,
  `ResourceWithConfigure/ModifyPlan/ValidateConfig/ImportState`; composed schema
  (envelope + body) validated by `Schema.ValidateImplementation` in tests; unified
  `put`/`Read`/`Delete`; runtime body loader reads the authoritative types.json from
  `azure.StaticFiles` (`loader.go`).
- **Hooks** — `hooks.go`: `Before/After` per op + `ValidateConfig`/`ModifyPlan`
  overrides; storage overlay (`overlay_storage_account.go`) implements the SKU
  zone-migration ForceNew via `azwise.CheckForceNew`.
- **Generated descriptor** — `generated.Descriptor{Name, ARMType, APIVersion, Schema}`
  registered via each generated file's `init()`.
- **Provider** — `Resources()` appends `azapinresource.New(name)` for every
  `generated.Registry` entry.
- Not yet exercised against live ARM (acceptance tests need credentials); unit and
  schema-validation coverage in place.

## Open Questions

1. **Generic state I/O.** Implement expand/flatten over `tftypes.Value`
   (`req.Plan.Raw`) or over `types.Object`? `tftypes.Value` is lower-level but
   avoids per-attribute path juggling; lean that way and confirm in phase 1.
2. **Operational knobs.** Do v1 static resources need `retry` / `ignore_casing` /
   `response_export_values`? Proposed: no — keep the surface minimal; revisit once
   real configs demand it.
3. **Update semantics.** ARM uses PUT for create and update alike; confirm whether
   any generated resource needs PATCH (rare) and, if so, expose it via a hook.
