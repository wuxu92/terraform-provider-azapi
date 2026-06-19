# Azapin — Static Schema Resources for AzAPI

## Overview

**Azapin** (azapi + native) generates static Terraform resource schemas from the bicep type definitions already embedded in `internal/azure/generated/`. Each generated resource ships inside the `terraform-provider-azapi` binary alongside `azapi_resource`, giving users a choice between dynamic body (day-zero, any ARM type) and typed schema (autocomplete, typed diffs, native ForceNew/Computed).

`azapi_resource` is **kept untouched** — both approaches coexist. Users adopt static resources incrementally, one resource at a time, or not at all. Preview-only resource types fall back to `azapi_resource`.

## Architecture

```
┌──────────────────────────────────────────────────────────────────────────┐
│ go generate ./internal/azapin/...                                        │
│                                                                          │
│  ┌──────────────┐     ┌─────────────────┐     ┌──────────────────────┐ │
│  │ types.json   │────▶│    Generator     │────▶│ Generated Go files   │ │
│  │ (bicep types)│     │                 │     │ (schema + CRUD)      │ │
│  └──────────────┘     │  - Walker       │     └──────────────────────┘ │
│                       │  - PostProcess  │                               │
│  ┌──────────────┐     │  - Emitter      │     ┌──────────────────────┐ │
│  │ overrides/   │────▶│  - Namer        │────▶│ Overlay hooks        │ │
│  │ (manual)     │     └─────────────────┘     │ (ForceNew, etc.)     │ │
│  └──────────────┘                             └──────────────────────┘ │
└──────────────────────────────────────────────────────────────────────────┘
```

### Components

| Package | Responsibility |
|---|---|
| `internal/azapin/generator` | `go generate` tool: walks type graph, post-processes, emits Go source |
| `internal/azapin/generator/walker.go` | Parses `types.json` into in-memory type graph with cycle detection |
| `internal/azapin/generator/postprocess.go` | Extracts defaults from descriptions, description-based validators, promotes single-optional children |
| `internal/azapin/generator/emitter.go` | Renders type graph → Go source with Terraform schema, conditional imports |
| `internal/azapin/naming` | ARM type → Terraform resource name + property name conversion |
| `internal/azapin/schema` | Runtime utilities: `StaticBool`, `StaticString`, `StaticInt64` default implementations |
| `internal/azapin/generated/` | Output: generated resource files (one per ARM type) |
| `internal/azapin/overrides/` | Manual overlay files (ForceNew, custom validation, etc.) |

### Generation Pipeline

```
types.json → ParseTypesJSON() → PostProcess() → EmitSchema() → Go source
```

1. Load `internal/azure/generated/index.json` to find all resource types
2. For each resource type, select the latest stable (non-preview) API version
3. Load and parse `types.json` for that version into an in-memory type graph
4. **Post-process** the type graph:
   - Extract default values from property descriptions (Rule 8)
   - Extract validators from property descriptions: ARM resource IDs, datetime formats, numeric ranges (Rule 12)
   - Promote single-optional-child block properties to Required (Rule 9)
5. Scan the post-processed graph to determine which imports are needed (conditional)
6. Emit Go source file with a `Schema()` function returning the full Terraform schema
7. Register resource with the provider

### Runtime Architecture

Generated resources reuse the existing AzAPI client infrastructure:
- Same `clients.Client` for ARM REST calls
- Same authentication, retry, and error handling
- Same provider configuration (subscription, features, etc.)

The difference is only at the schema level — instead of a single `body` dynamic attribute, each property is a typed Terraform attribute.

**No static property map**: CRUD operations do NOT use a compiled property mapping table. Instead, at runtime the provider reads the bicep type graph already embedded in `internal/azure/generated/` to look up the original ARM camelCase property names. This avoids duplicating ~215 property mappings per resource across 2,631 resource types.

## Resource Naming Convention

### Formula

```
azapi_<service>_<resource_path>
```

Where:
- `<service>` = ARM namespace without "Microsoft." prefix, lowercased (e.g., `storage`, `keyvault`, `network`, `compute`). For third-party providers (e.g., `Dynatrace.Observability`), uses the vendor name (`dynatrace`).
- `<resource_path>` = ARM resource segments converted to snake_case, singularized, joined with `_`
- If the first resource segment starts with the service name, the redundant prefix is stripped (e.g., `storage/storageAccounts` → `storage_account`, not `storage_storage_account`)

### Examples

| ARM Resource Type | Terraform Resource |
|---|---|
| `Microsoft.Storage/storageAccounts` | `azapi_storage_account` |
| `Microsoft.Storage/storageAccounts/blobServices` | `azapi_storage_account_blob_service` |
| `Microsoft.KeyVault/vaults` | `azapi_keyvault_vault` |
| `Microsoft.KeyVault/vaults/keys` | `azapi_keyvault_vault_key` |
| `Microsoft.Network/virtualNetworks` | `azapi_network_virtual_network` |
| `Microsoft.Network/virtualNetworks/subnets` | `azapi_network_virtual_network_subnet` |
| `Microsoft.Network/networkSecurityGroups` | `azapi_network_security_group` |
| `Microsoft.Compute/virtualMachines` | `azapi_compute_virtual_machine` |
| `Microsoft.Compute/virtualMachines/extensions` | `azapi_compute_virtual_machine_extension` |
| `Microsoft.ContainerService/managedClusters` | `azapi_containerservice_managed_cluster` |
| `Microsoft.Web/sites` | `azapi_web_site` |
| `Microsoft.Sql/servers/databases` | `azapi_sql_server_database` |
| `Microsoft.Cache/redis` | `azapi_cache_redis` |
| `Dynatrace.Observability/monitors` | `azapi_dynatrace_monitor` |

### Edge Case Handling

- **Hyphens in ARM names**: Converted to underscores before snake_case conversion (`api-version-sets` → `api_version_set`)
- **Namespace-qualified sub-resources**: Dot-containing segments (e.g., `Microsoft.Consumption`) are skipped
- **Singularization**: Handles regular plurals, `-ies` → `-y`, `-ses` → `-se`, and common irregulars (`databases` → `database`, `redis` → `redis`)
- **Build-time collision detection**: 9 collisions across 3,246 types (99.7% clean). Colliding types use full namespace prefix.

### Validation

Tested against all 3,246 unique ARM resource types in `index.json`:
- 0 invalid Terraform names
- 9 collisions (Admin namespaces, duplicate ARM naming)
- Collision test runs as part of `go test ./internal/azapin/naming/`

## Property Naming

### camelCase → snake_case (Required by Terraform SDK)

The Terraform plugin framework SDK enforces `^[a-z_][a-z0-9_]*$` for attribute names — camelCase produces a hard error. The generator converts ARM property names to snake_case.

The original ARM name is **not stored in the generated code**. At runtime, the provider looks up the ARM name from the embedded bicep type graph in `internal/azure/generated/`. This is a direct lookup, not a heuristic `snake2camel` conversion.

Algorithm (`naming.CamelToSnake`):
- Insert `_` before an uppercase letter that follows a lowercase letter or digit
- Insert `_` at the end of an uppercase acronym (before the next lowercase)
- Examples: `minimumTlsVersion` → `minimum_tls_version`, `isHnsEnabled` → `is_hns_enabled`, `isNfsV3Enabled` → `is_nfs_v3_enabled`, `iPRules` → `ip_rules`

Note: Alignment with AzureRM naming is explicitly a non-goal. Generated resources use a mechanical, predictable conversion.

## Schema Flag Rules (Summary)

See GENERATOR.md for the full 17-rule specification. Key behaviors:

| Condition | Terraform Schema |
|---|---|
| Bicep `Required` flag | `Required: true` |
| Bicep `ReadOnly` flag | `Computed: true` only, no validators |
| Block with ALL descendants ReadOnly | `Computed: true` (even if own flag says Optional) |
| Default value extracted from description | `Optional: true` + `Default: staticValue` |
| Single optional child in a block | Promoted to `Required` |
| Default (non-Required, non-ReadOnly) | `Optional: true, Computed: true` (safe default) |
| Bicep `WriteOnly` flag | `Sensitive: true` |
| Enum type (settable) | `stringvalidator.OneOf(...)` |
| ARM resource ID in description | `stringvalidator.RegexMatches(...)` |
| Numeric range in description | `int64validator.Between/AtLeast/AtMost` |

## Scope

### Phase 1: Proof of Concept ✓

Generated and validated:
1. `azapi_storage_account` (Microsoft.Storage/storageAccounts@2025-01-01) — 1,284 lines, compiles cleanly

### Phase 2: Full Generation

Generate all 2,631 stable resource types. Ship alongside `azapi_resource`.

## Key Design Decisions

1. **Latest stable API version only**: One resource per ARM type, pinned to latest stable. Preview → use `azapi_resource`.
2. **No AzureRM naming alignment**: Mechanical conversion from ARM names. Predictable, not curated.
3. **Coexistence**: `azapi_resource` is unchanged. Users choose per resource.
4. **Same binary**: Generated resources compile into the same provider binary. Binary is already 358 MB (334 MB embedded bicep types); generated code adds marginal size.
5. **Azwise feeds overlays**: Existing azwise knowledge files (ForceNew, validation, computed fields) inform the override layer.
6. **No static property map**: ARM names resolved at runtime from embedded type graph, not duplicated per resource.
7. **Safe default for Optional properties**: `Optional + Computed` unless we have evidence the server never populates the field, or we know the actual default value.
8. **Description mining**: The generator extracts defaults, validation patterns, and format hints from ARM property descriptions — information that the formal spec doesn't capture.
9. **Discriminated unions**: Flatten for simple cases (2-3 variants); fall back to `types.Dynamic` for complex ones.
10. **Conditional imports**: Generated files only import packages they actually use, detected by pre-scanning the type graph.

## Current State

| Component | Status | Files |
|---|---|---|
| Naming package | ✓ Complete | `naming/naming.go`, `naming_test.go`, `collision_test.go` |
| Type graph walker | ✓ Complete | `generator/walker.go`, `walker_test.go` |
| Post-processing | ✓ Complete | `generator/postprocess.go` |
| Schema emitter | ✓ Complete | `generator/emitter.go`, `emitter_test.go` |
| Runtime defaults | ✓ Complete | `schema/defaults.go` |
| PoC: storage account | ✓ Generated | `generated/storage_account_gen.go` (compiles) |
| PoC generator | ✓ Working | `generator/cmd/generate_poc.go` |
| CRUD methods | ○ Not started | — |
| Provider registration | ○ Not started | — |
| Full generation tool | ○ Not started | — |
| Override system | ○ Not started | — |
