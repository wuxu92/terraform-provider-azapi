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

Each ARM resource type gets a dedicated Terraform resource with a fully typed schema generated from the bicep type definitions in `internal/azure/generated/`. Users write HCL with named attributes, autocomplete, and plan-time type checking.

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

## Head-to-Head Comparison

| Dimension | Dynamic Body + Azwise | Static Schema Generation |
|---|---|---|
| **User autocomplete** | None — raw JSON keys | Full IDE/editor autocomplete |
| **Type checking** | Runtime (schema_validation_enabled) | Plan-time, per attribute |
| **Plan diff quality** | "body changed" (opaque) | Per-property diff |
| **ForceNew detection** | Azwise runtime lookup → plan modifier | Schema-level `RequiresReplace` per attribute |
| **Computed fields** | Azwise strips from diff | Schema `Computed: true` — native |
| **Validation** | Azwise validates at plan time | Schema validators — native |
| **Day-zero coverage** | Instant — any ARM type works with `body = {}` | Must generate + compile + release first |
| **API version flexibility** | User sets `type = "...@version"`, body is freeform | Each version needs a schema; changing version may break config |
| **ARM naming** | User writes exact ARM camelCase | Must map camelCase ↔ snake_case (error-prone) |
| **Cross-field references** | `output` + `response_export_values` | Native `resource.attr` references |
| **Code volume** | ~1K LOC for azwise core + ~300 LOC per resource knowledge | ~500-2,000 LOC per generated resource × 3,314 types |
| **Binary size** | Negligible (knowledge is data, not schema) | +50-200 MB estimated for full generation |
| **Maintenance** | Knowledge files updated manually or via extraction | Regenerate from bicep types (automated but needs review) |
| **Migration from existing** | No migration — `azapi_resource` unchanged | Every user must rewrite `body = {}` → named attributes |
| **AzureRM migration** | Different syntax than AzureRM | Closer to AzureRM feel (named attributes) |
| **Sensitive/write-only** | `sensitive_body` attribute | Schema `Sensitive: true` per attribute |
| **Discriminated unions** | JSON handles polymorphism naturally | Requires complex schema modeling |

## Detailed Analysis

### Where Static Schema Wins

**1. User experience for known resources.**
For the top ~100 resource types that most users interact with, a typed schema is dramatically better. Autocomplete prevents typos, plan diffs are readable, and ForceNew/Computed behavior is transparent without any hidden runtime logic.

**2. Terraform ecosystem integration.**
Static schemas work with `terraform validate`, `tflint`, IDE plugins, and documentation generators out of the box. Dynamic bodies are opaque to all of these tools.

**3. No invisible behavior.**
Azwise silently strips computed fields, validates properties, and detects ForceNew — but users don't know this is happening until it doesn't work. Static schemas make all constraints visible in the schema definition.

**4. State management.**
Terraform knows every attribute's type and can diff, import, and migrate state reliably. With dynamic bodies, state contains raw JSON that requires custom diff logic (`IgnoreNoOpChanges`, `ignore_missing_property`, `ignore_casing`).

### Where Dynamic Body + Azwise Wins

**1. Day-zero coverage.**
Any ARM resource type works immediately — no code generation, no compilation, no release cycle. Users can manage a new Azure service the day it launches. This is the core value proposition of azapi and the reason it exists.

**2. API version flexibility.**
Changing `@2023-01-01` to `@2025-01-01` just works. The body is freeform; new properties are accepted without schema changes. Static schemas would require regeneration and potentially config rewrites for each version upgrade.

**3. No naming translation.**
Users write the exact ARM property names they see in the Azure API documentation, Azure Portal, and ARM templates. No mental translation between `minimumTlsVersion` (ARM) and `minimum_tls_version` (Terraform). This eliminates an entire class of "what's the Terraform name for X?" questions.

**4. Polymorphic types work naturally.**
ARM's discriminated unions (e.g., virtual machine image reference types, encryption key sources) are just JSON. No need for complex schema modeling or "one of these blocks" validators.

**5. Minimal code volume.**
Azwise knowledge for storage account is ~220 lines of Go. A generated static schema for the same resource would be 1,000-2,000 lines, and there are 3,314 resource types to cover.

**6. No migration.**
Existing `azapi_resource` users are unaffected. Azwise improvements are invisible — they just get better validation, ForceNew detection, and diff suppression without changing their configs.

**7. Escape hatch for ARM edge cases.**
ARM APIs sometimes accept properties not in their spec, or have behavior that differs from the spec. Dynamic body lets users work around these issues. Static schemas would reject valid configurations when the spec is wrong.

### The Naming Problem

This deserves special attention because it's the single largest technical risk of static schema generation.

ARM property naming is inconsistent:
- `minimumTlsVersion` vs `minTlsVersion` (different services)
- `isHnsEnabled` → `is_hns_enabled` or `hns_enabled`?
- `iPRules` → `ip_rules` or `i_p_rules`?
- `allowBlobPublicAccess` → `allow_blob_public_access`
- `supportsHttpsTrafficOnly` → `supports_https_traffic_only` or `https_traffic_only`?

AzureRM spent 10+ years manually curating these names, with breaking changes when they got it wrong. Automated conversion will produce names that are correct-but-ugly or incorrect-and-confusing. There is no algorithm that handles all ARM naming conventions correctly.

### The Scale Problem

| What | Dynamic + Azwise | Static Generation |
|---|---|---|
| Resources to support "well" | ~20-50 high-priority | 3,314 (all or nothing) |
| LOC per resource | ~200-300 (knowledge file) | ~1,000-2,000 (generated schema) |
| Total new LOC | ~10K-15K | ~3-6M |
| Compilation time impact | None | Significant (+minutes) |
| Binary size impact | ~100KB | +50-200MB |
| CI/CD impact | Minimal | Major (build time, test matrix) |

## Hybrid Approach

The two approaches are not mutually exclusive. A practical path:

### Phase 1: Dynamic Body + Azwise (Current)
Continue enhancing azwise for high-priority resources. This delivers incremental value (validation, ForceNew, computed field handling) without any user-facing changes or migration.

### Phase 2: Static Schema for Top-N Resources
Generate typed resources for the ~20-50 most-used resource types. These coexist alongside `azapi_resource` — users choose the experience they want:
- `azapi_resource` for day-zero, flexibility, or raw ARM access
- `azapi_storage_account` for autocomplete, typed diffs, and AzureRM-like experience

### Phase 3: Azwise Knowledge Feeds Generation
The azwise knowledge (ForceNew, defaults, computed fields, validation) becomes input to the code generator. Instead of maintaining knowledge files AND generated schemas, knowledge files generate the custom overlays that augment the generated schemas.

### Value Matrix

| Approach | Coverage | UX | Effort | Risk |
|---|---|---|---|---|
| Azwise only | 20-50 resources, deep | Good (invisible improvements) | Low | Low |
| Static only | 3,314 resources, shallow | Excellent (autocomplete, typed diffs) | Very high | High (naming, scale, migration) |
| Hybrid (recommended) | 20-50 static + all dynamic | Best of both | Medium | Medium |

## Recommendation

**Start with azwise, prepare for hybrid.**

1. **Short term**: Complete azwise for the top 20 resource types. This delivers immediate value with minimal risk. Every azwise knowledge file is also a specification that a future code generator can consume.

2. **Medium term**: Build a proof-of-concept schema generator that converts one `types.json` → Terraform resource schema. Validate the naming algorithm, discriminated union handling, and plan modifier generation on 3-5 resource types.

3. **Long term**: If the PoC proves viable, generate static resources for the top 50 types and ship them alongside `azapi_resource`. Users opt into the typed experience per resource type.

The key insight: **azwise knowledge is not wasted work** — it's the operational knowledge that any approach needs. Whether the schema is dynamic or static, you still need to know which properties are ForceNew, what the valid enum values are, and which fields are computed. Azwise captures this knowledge in a portable format.
