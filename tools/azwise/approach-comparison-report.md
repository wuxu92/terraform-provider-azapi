# Comparison: Dynamic Body + Azwise vs Static Schema Generation

## The Two Approaches

### Approach A: Dynamic Body + Azwise Knowledge (Current)

`azapi_resource` accepts a freeform `body` (dynamic JSON). The **azwise** subsystem provides a plugin layer that injects operational knowledge (ForceNew, validation, computed fields, defaults) at runtime via Go registry lookups. The user writes raw ARM JSON; azwise silently improves the experience behind the scenes.

```hcl
resource "azapi_resource" "storage" {
  type      = "Microsoft.Storage/storageAccounts@2025-01-01"
  name      = "mystorage"
  parent_id = azurerm_resource_group.example.id
  location  = "westus2"

  body = {
    kind = "StorageV2"
    sku  = { name = "Standard_LRS" }
    properties = {
      minimumTlsVersion = "TLS1_2"
      accessTier        = "Hot"
    }
  }
}
```

### Approach B: Static Schema Generation

Each ARM resource type gets a dedicated Terraform resource with a fully typed schema generated from the bicep type definitions. Users write HCL with named attributes, autocomplete, and plan-time type checking. `azapi_resource` is **kept untouched** — both approaches coexist.

```hcl
resource "azapi_storage_account" "example" {
  name      = "mystorage"
  parent_id = azurerm_resource_group.example.id
  location  = "westus2"

  kind = "StorageV2"

  sku {
    name = "Standard_LRS"
  }

  properties {
    minimum_tls_version = "TLS1_2"
    access_tier         = "Hot"
  }
}
```

## Naming Constraint: Terraform SDK Enforces snake_case

The Terraform plugin framework SDK validates attribute names with `^[a-z_][a-z0-9_]*$` — camelCase names produce a **hard error**. This is enforced in `fwschema/attribute_name_validation.go` and cannot be bypassed. Any static schema generation must convert ARM camelCase to snake_case.

This is the primary UX difference between the two approaches: `azapi_resource` users write exact ARM names (`minimumTlsVersion`), while generated resources use Terraform-style names (`minimum_tls_version`). The conversion is mechanical and predictable, but it is mandatory.

## Head-to-Head Comparison

| Dimension | Dynamic Body + Azwise | Static Schema Generation |
|---|---|---|
| **User autocomplete** | None — raw JSON keys | Full IDE/editor autocomplete |
| **Type checking** | Runtime (schema_validation_enabled) | Plan-time, per attribute |
| **Plan diff quality** | "body changed" (opaque) | Per-property diff |
| **ForceNew detection** | Azwise runtime lookup → plan modifier | Schema-level `RequiresReplace` per attribute |
| **Computed fields** | Azwise strips from diff | Schema `Computed: true` — native |
| **Validation** | Azwise validates at plan time | Schema validators — native |
| **Day-zero coverage** | Instant — any ARM type works | Falls back to `azapi_resource` for unsupported types |
| **API version handling** | User changes `@version`, body is freeform | Each resource pins to latest stable; preview uses `azapi_resource` |
| **Property naming** | Exact ARM camelCase | snake_case (SDK requirement); documented mapping |
| **Cross-field references** | `output` + `response_export_values` | Native `resource.attr` references |
| **Code volume** | ~1K LOC core + ~300 LOC per resource | ~500-2K LOC per resource × 2,631 types |
| **Binary size impact** | Negligible | Marginal (334 MB type data already embedded) |
| **Migration needed** | No — transparent improvements | No — coexists with `azapi_resource` |
| **Discriminated unions** | JSON handles naturally | Generator flattens or falls back to dynamic |
| **Nested depth** | Same depth in JSON body | Same depth in HCL blocks |
| **Sensitive/write-only** | `sensitive_body` attribute | Schema `Sensitive: true`; handled by generator |

## Detailed Analysis

### Where Static Schema Wins Clearly

**1. User experience for known resources.**
For the ~2,631 stable resource types, a typed schema is dramatically better. Autocomplete prevents typos, plan diffs are readable, and ForceNew/Computed behavior is transparent.

**2. Terraform ecosystem integration.**
Static schemas work with `terraform validate`, `tflint`, IDE plugins, and documentation generators out of the box. Dynamic bodies are opaque to all of these tools.

**3. No invisible behavior.**
Azwise silently strips computed fields, validates properties, and detects ForceNew — but users don't see this happening. Static schemas make all constraints visible and inspectable.

**4. State management.**
Terraform knows every attribute's type. No `ignore_missing_property`, `ignore_casing`, or custom diff suppression needed.

### Where Dynamic Body + Azwise Wins Clearly

**1. Day-zero coverage.**
Any ARM resource type works immediately — no generation, compilation, or release. This is the core value proposition of AzAPI and remains critical for preview APIs and newly launched services.

**2. No naming translation.**
Users write the exact ARM property names from Azure docs, Portal, and ARM templates. No mental translation between `minimumTlsVersion` (ARM) and `minimum_tls_version` (Terraform).

**3. Escape hatch for ARM edge cases.**
ARM APIs sometimes accept properties not in their spec, or have behavior that differs from the spec. Dynamic body lets users work around these. Static schemas would reject valid configurations when the spec is wrong.

### Non-Issues (Previously Overstated Cons)

Several concerns from the initial analysis are less significant than originally assessed:

**Schema drift across API versions**: This already exists for `azapi_resource` — users must update their `body` content when changing `@version`. A static schema makes the same changes explicit at plan time rather than as runtime API errors. This is arguably better, not worse.

**Breaking changes between API versions**: Same as above — these are an ARM API reality, not a schema generation problem. Static schemas surface them earlier (plan time vs apply time).

**Binary size**: The provider binary is already 358 MB, of which 334 MB is embedded bicep type data. Generated Go code adds comparatively little.

**Nested depth**: The same nesting exists whether properties are JSON objects in `body = {}` or HCL blocks in a static schema. Neither approach changes the ARM API structure.

**Existing user migration**: `azapi_resource` remains untouched. Users adopt static resources incrementally, one resource at a time, or not at all.

**Maintenance of overlays**: This is the same work azwise does today — curating ForceNew rules, validation, defaults. The difference is that overlays enhance a typed schema (schema-level constraints) rather than patching a dynamic body (runtime heuristics). The investment in better UX is the point, not a drawback.

## The Real Challenges

### 1. Property Name Conversion (Mandatory)

The Terraform SDK's `^[a-z_][a-z0-9_]*$` regex is a hard constraint. The generator must convert every ARM property name to snake_case and maintain a lossless reverse mapping for CRUD operations.

The conversion is mechanical (`minimumTlsVersion` → `minimum_tls_version`) and AzAPI already has `snake2camel` / `camel2snake` functions. Edge cases (acronyms, consecutive uppercase) need a build-time collision detector. Override tables handle the rare ambiguous cases.

This is solvable engineering, not a blocking risk. But it does mean generated resources use different property names than `azapi_resource` and ARM docs. Documentation must bridge this gap.

### 2. Discriminated Unions

ARM's polymorphic types need Terraform representation. Options:
- Flatten all variant properties with mutual-exclusivity validators (works for 2-3 variants)
- Fall back to `types.Dynamic` for complex unions (preserves static schema for everything else)

The generator has full type information from bicep types and can choose the best strategy per union.

### 3. ReadOnly Flag Accuracy

Some bicep type flags are wrong. The generator needs an override mechanism for properties that are flagged ReadOnly but are actually settable. This is a finite, curated set.

### 4. Compilation Time

~2,631 generated resources at 500-2,000 LOC each is significant. Sub-packaging, build tags, or incremental compilation strategies can keep development builds fast.

## Recommendation

**Build static schema generation as an additive layer on top of `azapi_resource`.**

The two approaches are complementary, not competing:

| User need | Best tool |
|---|---|
| Day-zero / preview API | `azapi_resource` |
| Stable resource, wants autocomplete | `azapi_storage_account` (generated) |
| Working around ARM spec bugs | `azapi_resource` |
| AzureRM migration target | Generated resources (closer UX) |
| CI/CD with strict validation | Generated resources (plan-time checks) |

Azwise knowledge is reusable in both approaches: ForceNew rules, validation, computed fields, and defaults feed into the overlay system that customizes generated resources. The investment is portable.

### Implementation Path

1. **Build the generator**: Convert `types.json` → Terraform resource schema for latest stable versions. Start with 5 resources as proof of concept.
2. **Solve naming**: Implement camelCase → snake_case with collision detection. Validate on all 2,631 resource types.
3. **Handle discriminated unions**: Choose strategy per union complexity. Measure how many resources have them.
4. **Ship alongside `azapi_resource`**: Both in the same provider binary. Users choose per resource.
5. **Feed azwise into overlays**: Reuse existing knowledge files as the override layer for generated resources.
