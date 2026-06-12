# Azapin — Static Schema Resources for AzAPI

## Overview

**Azapin** (azapi + native) generates static Terraform resource schemas from the bicep type definitions already embedded in `internal/azure/generated/`. Each generated resource ships inside the `terraform-provider-azapi` binary alongside `azapi_resource`, giving users a choice between dynamic body (day-zero, any ARM type) and typed schema (autocomplete, typed diffs, native ForceNew/Computed).

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│ go generate ./internal/azapin/...                                    │
│                                                                      │
│  ┌──────────────┐     ┌──────────────┐     ┌─────────────────────┐ │
│  │ types.json   │────▶│  Generator   │────▶│ Generated Go files  │ │
│  │ (bicep types)│     │              │     │ (schema + CRUD)     │ │
│  └──────────────┘     │  - Walker    │     └─────────────────────┘ │
│                       │  - Namer     │                              │
│  ┌──────────────┐     │  - Emitter   │     ┌─────────────────────┐ │
│  │ overrides/   │────▶│              │────▶│ Overlay hooks       │ │
│  │ (manual)     │     └──────────────┘     │ (ForceNew, etc.)    │ │
│  └──────────────┘                          └─────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

### Components

| Package | Responsibility |
|---|---|
| `internal/azapin/generator` | `go generate` tool: reads types.json, walks type graph, emits Go source |
| `internal/azapin/naming` | ARM type → Terraform resource name + property name conversion |
| `internal/azapin/schema` | Runtime schema utilities shared by generated resources |
| `internal/azapin/generated/` | Output: generated resource files (one per ARM type) |
| `internal/azapin/overrides/` | Manual overlay files (ForceNew, validation, etc.) |

### Generation Flow

```
go generate ./internal/azapin/generator
```

1. Load `internal/azure/generated/index.json`
2. For each resource type, find the latest stable API version
3. Load and parse `types.json` for that version
4. Walk the type graph from the ResourceType body reference
5. Convert property names: camelCase → snake_case (store both)
6. Emit Go file with:
   - Resource struct
   - `Schema()` method (Terraform schema from type graph)
   - `Create`/`Read`/`Update`/`Delete` methods (ARM REST calls)
   - Plan modifiers (from property flags + overrides)
7. Register resource with the provider

### Runtime Architecture

Generated resources reuse the existing AzAPI client infrastructure:
- Same `clients.Client` for ARM REST calls
- Same authentication, retry, and error handling
- Same provider configuration (subscription, features, etc.)

The difference is only at the schema level — instead of a single `body` dynamic attribute, each property is a typed Terraform attribute. The CRUD methods marshal/unmarshal between Terraform state and ARM JSON using the stored property name mapping.

## Resource Naming Convention

### Formula

```
azapi_<service>_<resource_path>
```

Where:
- `<service>` = ARM namespace without "Microsoft." prefix, lowercased (e.g., `storage`, `keyvault`, `network`, `compute`)
- `<resource_path>` = ARM resource segments converted to snake_case, singularized, joined with `_`
- If the first resource segment starts with the service name, the redundant prefix is stripped

### Examples

| ARM Resource Type | Terraform Resource |
|---|---|
| `Microsoft.Storage/storageAccounts` | `azapi_storage_account` |
| `Microsoft.Storage/storageAccounts/blobServices` | `azapi_storage_account_blob_service` |
| `Microsoft.KeyVault/vaults` | `azapi_keyvault_vault` |
| `Microsoft.KeyVault/vaults/keys` | `azapi_keyvault_vault_key` |
| `Microsoft.Network/virtualNetworks` | `azapi_network_virtual_network` |
| `Microsoft.Network/virtualNetworks/subnets` | `azapi_network_virtual_network_subnet` |
| `Microsoft.Compute/virtualMachines` | `azapi_compute_virtual_machine` |
| `Microsoft.ContainerService/managedClusters` | `azapi_containerservice_managed_cluster` |
| `Microsoft.Web/sites` | `azapi_web_site` |

### Collision Resolution

~10 collisions exist across 3,246 case-normalized resource types (mostly Admin namespaces and third-party providers like Dynatrace vs NewRelic). These are resolved by including the full namespace for colliding types:

```
Dynatrace.Observability/monitors  → azapi_dynatrace_monitor
NewRelic.Observability/monitors   → azapi_newrelic_monitor
```

## Property Naming

### Conversion: camelCase → snake_case

The generator stores both names per attribute:

```go
type PropertyMapping struct {
    TerraformName string // "minimum_tls_version" (used in HCL)
    ARMName       string // "minimumTlsVersion" (used in JSON payload)
}
```

Algorithm:
- Insert `_` before uppercase letters that follow lowercase or precede lowercase: `minimumTlsVersion` → `minimum_tls_version`
- Acronyms stay grouped: `isHnsEnabled` → `is_hns_enabled`
- Build-time collision detection across all properties of a resource

CRUD operations use `ARMName` directly for JSON marshaling — no runtime heuristic.

## Scope

### Phase 1: Proof of Concept (5 resources)

Generate and validate:
1. `azapi_storage_account` (Microsoft.Storage/storageAccounts)
2. `azapi_keyvault_vault` (Microsoft.KeyVault/vaults)
3. `azapi_compute_virtual_machine` (Microsoft.Compute/virtualMachines)
4. `azapi_network_virtual_network` (Microsoft.Network/virtualNetworks)
5. `azapi_containerservice_managed_cluster` (Microsoft.ContainerService/managedClusters)

### Phase 2: Full Generation

Generate all 2,631 stable resource types. Ship alongside `azapi_resource`.

## Key Design Decisions

1. **Latest stable API version only**: One resource per ARM type, pinned to latest stable. Preview → use `azapi_resource`.
2. **No AzureRM naming alignment**: Mechanical conversion from ARM names. Users can predict the Terraform name from the ARM name.
3. **Coexistence**: `azapi_resource` is unchanged. Users choose per resource.
4. **Same binary**: Generated resources compile into the same provider binary. No separate provider.
5. **Azwise feeds overlays**: Existing azwise knowledge files inform the override layer (ForceNew, validation, computed fields).
6. **Discriminated unions**: Flatten for simple cases (2-3 variants); fall back to `types.Dynamic` for complex ones.
