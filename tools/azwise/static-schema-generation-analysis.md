# Static Schema Generation Analysis

## Overview

This document analyzes the approach of generating static Terraform resource schemas from the bicep type definitions already embedded in `internal/azure/generated/`. Instead of the current `azapi_resource` with a single `body` dynamic attribute, each Azure resource type would get its own Terraform resource (e.g., `azapi_storage_account`) with a fully typed schema derived from the ARM API spec.

## Source Material

AzAPI already embeds **31,175** resource type+version definitions covering **3,314** unique ARM resource types in `internal/azure/generated/index.json`. Each definition is a `types.json` file containing a full type graph: ObjectType, StringLiteralType, UnionType, ArrayType, IntegerType, BooleanType — with property flags (Required, ReadOnly, WriteOnly, etc.) and descriptions. This is the same data currently used for `schema_validation_enabled` runtime checks on the dynamic `body`.

## How It Would Work

### Schema Generation Pipeline

1. **Read** `types.json` for a resource type + API version
2. **Walk** the type graph starting from the ResourceType's body reference
3. **Map** each property to a Terraform `schema.Attribute`:
   - ObjectType → `schema.SingleNestedAttribute`
   - ArrayType of ObjectType → `schema.ListNestedAttribute`
   - StringType / StringLiteralType / UnionType of strings → `schema.StringAttribute` with validators
   - IntegerType → `schema.Int64Attribute`
   - BooleanType → `schema.BoolAttribute`
4. **Apply** property flags:
   - Flag `Required` → `Required: true`
   - Flag `ReadOnly` → `Computed: true`
   - Flag `WriteOnly` → `Sensitive: true` (or a write-only marker)
   - Default → `Optional: true, Computed: true`
5. **Emit** Go source with resource struct, Schema(), CRUD methods, and plan modifiers

### Resource Registration

Each generated resource registers itself via `init()`:

```go
// Generated: Microsoft.Storage/storageAccounts@2025-01-01
func init() {
    RegisterResource("azapi_microsoft_storage_storage_account", NewStorageAccountResource)
}
```

### Customization Layer

Generated resources would accept overlay functions for custom logic:

```go
// Hand-written overlay in storage_account_overlay.go
func init() {
    OverlayResource("azapi_microsoft_storage_storage_account", StorageAccountOverlay{})
}

type StorageAccountOverlay struct{}

func (o StorageAccountOverlay) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
    // Custom ForceNew for SKU zone migration
}

func (o StorageAccountOverlay) Validate(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
    // Custom cross-field validation
}
```

## Pros

### 1. Full IDE Autocomplete and Type Safety

Every property is a named Terraform attribute. Users get autocomplete, type checking at plan time, and clear error messages pointing to the exact attribute — not a JSON path inside a dynamic blob.

```hcl
# Static schema — autocomplete works, typos caught at plan time
resource "azapi_storage_account" "example" {
  name     = "mystorage"
  location = "westus2"

  sku {
    name = "Standard_LRS"  # validated enum
  }

  properties {
    minimum_tls_version = "TLS1_2"  # autocomplete, validated
    access_tier         = "Hot"      # validated enum
  }
}
```

### 2. Property-Level Plan Diffs

Terraform can show which specific property changed, not "body changed":

```diff
# azapi_storage_account.example will be updated in-place
  ~ resource "azapi_storage_account" "example" {
      ~ properties {
          ~ access_tier = "Hot" -> "Cool"
        }
    }
```

### 3. ForceNew / Computed / Sensitive at the Schema Level

No runtime azwise lookup needed — the schema itself declares `ForceNew`, `Computed`, `Sensitive` per property. Terraform handles plan modification natively.

### 4. Cross-Field References Work Naturally

```hcl
resource "azapi_storage_account" "example" {
  properties {
    minimum_tls_version = "TLS1_2"
  }
}

output "tls_version" {
  value = azapi_storage_account.example.properties.minimum_tls_version
}
```

No `response_export_values` or `output` dynamic extraction needed.

### 5. Documentation Generated Automatically

Terraform registry docs can be generated from the schema descriptions embedded in `types.json`. Each property gets its own doc line with ARM description.

### 6. ARM API Coverage Is Complete by Construction

The bicep types cover every ARM resource type and API version. No manual extraction or knowledge transfer needed — the schema IS the API.

### 7. State Migration and Import Are Simpler

Each attribute has a known type and path. `terraform import` can populate the state from a GET response by walking the schema, not guessing dynamic types.

## Cons

### 1. Massive Code Generation Volume

3,314 unique resource types × multiple API versions = potentially tens of thousands of generated Go files. Binary size, compilation time, and test surface area explode.

**Mitigation**: Generate only "popular" resources, or generate on-demand per API version. But this reintroduces the coverage gap problem.

### 2. Property Name Mapping: ARM camelCase → Terraform snake_case

ARM uses `minimumTlsVersion`; Terraform convention is `minimum_tls_version`. This mapping must be:
- Consistent (same algorithm everywhere)
- Reversible (to reconstruct the ARM JSON)
- Collision-free (e.g., `iPRules` → `ip_rules` vs `ipRules` → `ip_rules`)

Edge cases: acronyms (DNS, TLS, HTTP, SKU), single-letter segments, consecutive uppercase. AzureRM's manual naming fixes over 10 years demonstrate this is non-trivial.

### 3. Schema Drift Across API Versions

Each API version has a different schema. Users must upgrade their `.tf` config when changing `api_version`. A property added in `2025-01-01` doesn't exist in `2023-01-01` — Terraform would reject it at plan time.

The current dynamic body handles this transparently: any JSON is accepted, validation is optional.

### 4. Breaking Changes Between API Versions

ARM API versions sometimes rename properties, change nesting, or alter types. A static schema makes these breaking changes visible to Terraform users as plan-time errors — which is both a pro (explicit) and a con (painful upgrades).

### 5. Read-Only vs Settable Is Not Always Clear in Bicep Types

The bicep type flags mark properties as ReadOnly, but this flag is sometimes inaccurate:
- Some "ReadOnly" properties are actually settable in PUT (the spec is wrong)
- Some settable properties are marked ReadOnly because they're only meaningful on GET
- The flag granularity is per-API-version and may differ from what the service actually accepts

This leads to false rejections: "you can't set X" when the ARM API happily accepts it.

### 6. Discriminated Unions Are Hard to Map

ARM uses discriminated unions (polymorphic types based on a `type` field). Mapping these to Terraform's type system requires either:
- One attribute per variant (with mutual exclusivity validators) — verbose
- A single dynamic attribute for the union body — defeats the purpose
- Terraform's `object` type with all variants flattened — confusing

### 7. Nested Block Depth Explosion

ARM resources like Virtual Machines have deeply nested properties (5-7 levels). In Terraform schema, each level becomes a `SingleNestedAttribute` or `ListNestedAttribute`. The resulting HCL is deeply indented and verbose.

### 8. Write-Only and Sensitive Properties

ARM has "write-only" properties (accepted in PUT, never returned in GET). Terraform's state stores all attributes, so write-only properties cause perpetual diffs unless special plan modifiers suppress them. This requires per-property handling.

### 9. Existing azapi_resource Users Must Migrate

Every user of `azapi_resource` for a generated resource type would need to migrate their config. The migration path from dynamic body to static schema is non-trivial: JSON structure → HCL attributes with different naming.

### 10. ARM Envelope Fields Are Duplicated

ARM resources have envelope fields (`name`, `location`, `tags`, `identity`, `sku`) that azapi_resource already handles as top-level attributes. Generated resources must avoid duplicating these while still including them in the ARM body.

### 11. Maintenance Burden of Overlays

Custom overlays (ForceNew, validation, defaults) must be maintained per resource type, similar to what azwise does today. The overlay system adds framework complexity, and overlays must be updated when the generated schema changes.

## Scale Estimate

| Metric | Value |
|---|---|
| Unique ARM resource types | 3,314 |
| Total type+version definitions | 31,175 |
| Storage account properties (2025-01-01) | 688 type entries in types.json |
| Estimated generated Go LOC per resource | 500-2,000 |
| Total generated Go LOC (all resources) | ~2-6M lines |
| Binary size impact | +50-200 MB (estimated) |

## Key Technical Risks

1. **camelCase → snake_case collisions**: ARM naming is inconsistent; automated conversion will produce collisions or unintuitive names
2. **Discriminated union mapping**: No clean Terraform representation for polymorphic ARM types
3. **State compatibility**: Generated schema must round-trip through Terraform state without data loss, including properties the schema doesn't know about (forward compatibility)
4. **API version selection**: Must decide whether each resource is version-pinned or version-flexible; either choice has drawbacks
5. **Compilation time**: Tens of thousands of Go files will significantly slow `go build` and CI

## Conclusion

Static schema generation from bicep types is technically feasible — the type information is rich enough to produce valid Terraform schemas. The main value is a dramatically better user experience: autocomplete, typed diffs, native ForceNew/Computed. The main risk is the scale (3,314 resource types), the naming problem (camelCase → snake_case), and the loss of flexibility that makes azapi_resource uniquely useful as a "day-zero" provider.
