# Static Schema Generation Analysis

## Overview

This document analyzes the approach of generating static Terraform resource schemas from the bicep type definitions already embedded in `internal/azure/generated/`. Each Azure resource type would get its own Terraform resource (e.g., `azapi_storage_account`) with a fully typed schema derived from the ARM API spec. These static resources coexist alongside the existing `azapi_resource`, which remains untouched for day-zero and advanced use cases.

## Source Material

AzAPI already embeds **31,175** resource type+version definitions covering **3,314** unique ARM resource types (2,631 with stable API versions) in `internal/azure/generated/index.json`. Each definition is a `types.json` file containing a full type graph: ObjectType, StringLiteralType, UnionType, ArrayType, IntegerType, BooleanType — with property flags (Required, ReadOnly, WriteOnly, etc.) and descriptions. This is the same data currently used for `schema_validation_enabled` runtime checks on the dynamic `body`.

The current provider binary is already **358 MB**, with **334 MB** from the embedded bicep type data in `internal/azure/generated/`. Generated Go schema code would add comparatively little to the binary size.

## How It Would Work

### Scope: Latest Stable API Version Only

Generate one Terraform resource per ARM resource type, using the latest stable (non-preview) API version. This gives **~2,631 resources** — complete coverage of all stable ARM services in a single generation pass. Preview-only resource types are excluded; users fall back to `azapi_resource` for those.

### Schema Generation Pipeline

1. **Read** `types.json` for the latest stable API version of each resource type
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

### Property Naming: camelCase → snake_case (Required)

The Terraform plugin framework SDK **enforces** lowercase attribute names via `ValidAttributeNameRegex = regexp.MustCompile("^[a-z_][a-z0-9_]*$")` in `fwschema/attribute_name_validation.go`. camelCase names like `minimumTlsVersion` produce a hard error:

```
Invalid Attribute/Block Name: "minimumTlsVersion" at schema path "properties.minimumTlsVersion"
is an invalid attribute/block name. Names must only contain lowercase alphanumeric characters
(a-z, 0-9) and underscores (_).
```

This means the generator **must** convert ARM camelCase to Terraform snake_case. The conversion must be:
- **Deterministic**: Same algorithm everywhere (generator, CRUD marshaling, import)
- **Reversible**: Reconstruct the ARM JSON key from the Terraform attribute name
- **Documented**: Users need to know `minimumTlsVersion` → `minimum_tls_version`

Note: Alignment with AzureRM naming is explicitly a non-goal. AzureRM uses heavily curated names (e.g., `account_replication_type` for `sku.name`). The generated resources use a mechanical, predictable conversion that directly mirrors the ARM property structure.

The conversion algorithm is straightforward for the vast majority of cases:
- Insert `_` before each uppercase letter, then lowercase: `minimumTlsVersion` → `minimum_tls_version`
- Consecutive uppercase (acronyms) need care: `isHnsEnabled` → `is_hns_enabled`, `iPRules` → `ip_rules`

AzAPI already has this algorithm implemented as `snake2Camel` / `camel2Snake` functions for the `provider::azapi::snake2camel` / `provider::azapi::camel2snake` Terraform functions. The reverse is needed for schema generation.

Edge cases where collisions occur are rare. The generator should detect them at build time and fail with an explicit error, forcing a manual resolution via an override table.

### Resource Registration

```go
// Generated: Microsoft.Storage/storageAccounts@2025-01-01
func init() {
    RegisterResource("azapi_microsoft_storage_storage_account", NewStorageAccountResource)
}
```

### Customization Layer

Generated resources accept overlay functions for custom logic:

```go
// Hand-written overlay in storage_account_overlay.go
func init() {
    OverlayResource("azapi_microsoft_storage_storage_account", StorageAccountOverlay{})
}

type StorageAccountOverlay struct{}

func (o StorageAccountOverlay) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
    // Custom ForceNew for SKU zone migration
}
```

## Pros

### 1. Full IDE Autocomplete and Type Safety

Every property is a named Terraform attribute. Users get autocomplete, type checking at plan time, and clear error messages pointing to the exact attribute — not a JSON path inside a dynamic blob.

```hcl
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

Terraform shows which specific property changed, not "body changed":

```diff
# azapi_storage_account.example will be updated in-place
  ~ resource "azapi_storage_account" "example" {
      ~ properties {
          ~ access_tier = "Hot" -> "Cool"
        }
    }
```

### 3. ForceNew / Computed / Sensitive at the Schema Level

No runtime lookup needed — the schema itself declares `ForceNew`, `Computed`, `Sensitive` per property. Terraform handles plan modification natively.

### 4. Cross-Field References Work Naturally

```hcl
output "tls_version" {
  value = azapi_storage_account.example.properties.minimum_tls_version
}
```

No `response_export_values` or `output` dynamic extraction needed.

### 5. Documentation Generated Automatically

Terraform registry docs can be generated from the schema descriptions embedded in `types.json`. Each property gets its own doc line with ARM description.

### 6. ARM API Coverage Is Complete by Construction

2,631 stable resource types are covered by the bicep type definitions. No manual extraction or knowledge transfer needed — the schema IS the API.

### 7. State Migration and Import Are Simpler

Each attribute has a known type and path. `terraform import` can populate the state from a GET response by walking the schema.

### 8. Coexists with azapi_resource

`azapi_resource` remains untouched. Users choose per resource:
- `azapi_resource` for day-zero coverage, preview APIs, advanced use cases, or when they prefer raw ARM JSON
- `azapi_storage_account` (etc.) for autocomplete, typed diffs, and AzureRM-like experience

No migration required. Users adopt static resources incrementally, one resource at a time.

## Cons

### 1. Property Name Translation (Required by Terraform SDK)

The Terraform framework SDK enforces `^[a-z_][a-z0-9_]*$` for attribute names — camelCase is rejected with a hard error. The generator must convert all ARM property names to snake_case.

This introduces a translation layer users must understand: ARM docs say `minimumTlsVersion`, Terraform config uses `minimum_tls_version`. The mapping is mechanical and predictable, but it's still a mental translation that `azapi_resource` doesn't require.

**Mitigations:**
- The conversion algorithm is deterministic — users can predict it
- The provider already ships `snake2camel` / `camel2snake` functions
- Generated documentation maps each Terraform name to its ARM path
- Collision detection at build time catches the rare ambiguous cases

### 2. Code Generation Volume

2,631 stable resource types with ~500-2,000 LOC per resource = ~1.3-5.2M lines of generated Go. However:
- The bicep type data (334 MB) is already embedded — generated code adds comparatively little
- Generated code is boilerplate and can be organized into sub-packages for parallel compilation
- Only the latest stable version per resource type is generated, not all 31K definitions
- Build tags or conditional compilation can exclude generated resources from development builds

### 3. Read-Only Flag Inaccuracy in Bicep Types

Some properties are flagged ReadOnly in the spec but are actually settable in PUT. The generator would reject valid configurations for these properties.

**Mitigations:**
- The generator can accept an override file listing properties where the ReadOnly flag should be ignored
- Runtime discovery: if a property is in the Create model but flagged ReadOnly, the generator can emit it as `Optional + Computed` instead of `Computed`-only
- These overrides are a finite, knowable set that can be curated and shared

### 4. Discriminated Unions

ARM uses polymorphic types keyed by a discriminator field (e.g., `type`). Terraform has no native discriminated union type.

**Mitigations:**
- The generator can flatten all variant properties into a single block with mutual-exclusivity validators
- For simple cases (2-3 variants), separate blocks with `ExactlyOneOf` work well
- For complex cases, fall back to a `types.Dynamic` attribute for that specific property, preserving the static schema for everything else
- The bicep type graph already encodes discriminators — the generator has full type information

### 5. Maintenance of Custom Overlays

Custom logic (conditional ForceNew, cross-field validation, non-standard defaults) must be maintained per resource type as overlay files. This is necessary work to provide a better user experience — the same work azwise does today, but with the advantage of being schema-level rather than runtime heuristics.

The overlay system is additive: a resource works correctly without any overlay (the generated schema handles standard cases), and overlays add polish for resources that need it.

## Scale Estimate

| Metric | Value |
|---|---|
| Unique ARM resource types (stable) | 2,631 |
| Preview-only resource types | 683 |
| Total type+version definitions | 31,175 |
| Current provider binary size | 358 MB (334 MB embedded bicep types) |
| Estimated generated Go LOC (latest stable only) | ~1.3-5.2M lines |
| Estimated additional binary size | Marginal (type data already embedded) |

## Key Technical Risks

1. **snake_case naming edge cases**: Consecutive uppercase, acronyms, single-letter segments. Build-time collision detection catches problems, but override curation is needed for a minority of resources.
2. **Discriminated union UX**: Flattened variants may be confusing for complex polymorphic types. Fall back to dynamic attributes for the worst cases.
3. **State round-tripping**: Generated schema must handle the snake_case ↔ camelCase conversion losslessly when reading/writing ARM JSON.
4. **Compilation time**: Large generated codebase may need sub-packaging or build-tag gating to keep development builds fast.

## Conclusion

Static schema generation from bicep types is feasible and delivers a dramatically better user experience for the ~2,631 stable ARM resource types. The main engineering challenge is the mandatory snake_case conversion (enforced by Terraform SDK) and discriminated union modeling. Both are solvable with known techniques.

The approach is additive — `azapi_resource` remains the escape hatch for day-zero, preview, and advanced use cases. Generated resources are a strictly better experience for stable, well-defined ARM resource types.

Azwise knowledge is not wasted: ForceNew rules, validation constraints, and computed field identification feed directly into the overlay system that customizes generated resources beyond what the bicep types alone can express.
