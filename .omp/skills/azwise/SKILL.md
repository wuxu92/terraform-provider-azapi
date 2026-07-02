---
name: azwise
description: Extract operational knowledge from terraform-provider-azurerm to generate Go knowledge files for AzAPI's azwise package
---

# azwise — AzureRM Knowledge Extraction

The `azwise` package (`internal/azure/azwise/`) encodes per-resource operational knowledge extracted from `terraform-provider-azurerm`. Each resource type gets a Go file that declares ForceNew rules, timeouts, soft-delete behavior, naming validation, sensitive fields, computed properties, server-default properties, required fields, and other constraints. AzAPI's plan/validate/apply hooks call into azwise so users get the same guardrails AzureRM provides — without maintaining a full AzureRM-style provider.

## Knowledge Categories

Scan the AzureRM resource file (e.g. `internal/services/<svc>/<resource>_resource.go`) for these patterns:
### API Versions
The `azwise_extract` tool extracts API versions from the Go source import paths (e.g. `resource-manager/keyvault/2023-02-01/vaults`). Set `ApiVersions` on `BaseKnowledge` so the registry can version-match knowledge to the AzAPI user's chosen API version. Leave empty only for data-plane resources where AzureRM doesn't use ARM API versions.
```go
ApiVersions: []string{"2023-02-01"}
```


### ForceNew
Look for `ForceNew: true` in `pluginsdk.Schema` fields. Map the schema path to the ARM JSON property path.
```go
ForceNew: []ForceNewRule{{PropertyPath: "name"}}
```

### Timeouts
Look for `pluginsdk.DefaultTimeout(N * time.Minute)` inside the `Timeouts` block of `Create`/`Read`/`Update`/`Delete`.
```go
TimeoutsConfig: &Timeouts{
    Create: 30 * time.Minute,
    Read:   5 * time.Minute,
    Update: 30 * time.Minute,
    Delete: 30 * time.Minute,
}
```

### SoftDelete
Look for purge/recover patterns in the CRUD functions, `features.KeyVault.PurgeSoftDeleteOnDestroy` feature flags, or explicit soft-delete API calls.
```go
SoftDelete: true
```

### Naming Validation
Look for `validate/` packages with regex patterns or length constraints for the resource name. Use a `StringRule` with an **empty** `PropertyPath` (empty means it validates the resource name, not a body property).
```go
StringRules: []StringRule{{
    Regex:     `^[a-z0-9]{3,24}$`,
    MinLength: 3, MaxLength: 24,
    Message:   "must be lowercase alphanumeric, 3-24 characters",
}}
```

### Sensitive Fields
Look for `Sensitive: true` in schema definitions. Map to ARM JSON property paths.
```go
SensitiveFields: []string{"properties.primaryAccessKey"}
```

### Computed Properties (Read-Only)
ARM properties that Azure populates in the response but that the user cannot set.
Use `azwise_extract` with `category=schema` and `resource_name=azurerm_<resource>` to get the list of computed-only fields from `provider-schema.json`.
Only include fields that map to ARM body properties — skip provider-internal fields like `resource_id`, `versionless_id`, `public_key_pem`.
```go
ComputedFields: []string{"properties.vaultUri"}
```

### Default Values (AzureRM-Recommended Defaults)
ARM properties where AzureRM or Azure fills in a default value if the user omits them. Stores the actual ARM-format default alongside the path.
```go
DefaultValues: []DefaultValue{
    {PropertyPath: "properties.minimumTlsVersion", Value: "TLS1_2"},
    {PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
    {PropertyPath: "properties.accessTier"},  // nil Value = server decides
}
```

### Required Fields
ARM body properties that must be present for a valid resource creation. Derived from `Required: true` in AzureRM schema — excludes envelope fields (name, location, resource_group_name). Include fields that AzureRM hardcodes but the ARM API requires (e.g. `sku.family = "A"` for KeyVault).
```go
RequiredFields: []string{"properties.tenantId", "properties.sku.name", "properties.sku.family"}
```
During `ValidateConfig`, missing required fields emit an error diagnostic with default-value hints when available.

### Property Validation
Look for enum values (often from SDK `constants.go`), `validation.IntBetween`, `validation.FloatBetween`, `validation.StringInSlice`, `MaxItems` in schema.
```go
StringRule{PropertyPath: "sku.name", AllowedValues: []string{"Standard_LRS", "Premium_LRS"}}
FloatRule{PropertyPath: "properties.weight", Min: ptrFloat(0), Max: ptrFloat(1000)}
IntRule{PropertyPath: "properties.capacity", Min: ptrInt(1), Max: ptrInt(100)}
ArrayRule{PropertyPath: "properties.ipRules", MaxItems: ptrInt(200)}
```

### Custom / semantic validators (→ azapin validators)

`StringRule`/`IntRule`/`FloatRule` only cover enum, regex, length, and range.
Transfer **every** other `ValidateFunc` too, routed by reusability:

- **Generic / cross-resource** (`validation.IsUUID`, `azure.ValidateResourceID`, …):
  use a shared validator in `internal/native/schema` (`UUID()`, `AzureResourceID()`,
  …) via `generator.SharedValidator("UUID()")`. Add a new one there only when the
  rule is genuinely cross-resource.
- **Resource-specific** (e.g. `StorageAccountIpRule` — regex *plus* a public-vs-private
  IP check): a `validator.String` in `internal/native/services/<service>/validators/<rule>.go`
  (package `validators`, e.g. `services/storage/validators/ip_rules.go`), one per
  file, referenced via `generator.CustomValidator("StorageAccountIPRule()")`.
- **Sub-service** (e.g. `BlobPropertiesDefaultServiceVersion`): goes on the
  sub-service resource, not the parent. **Non-mappable** (composite Key Vault key
  URI, map-key validators): skip with a note.

Attach by mapping the AzureRM field to its ARM body path and appending to
`p.Validators` via `generator.FindProperty(def, "<arm.path>")` in the resource
customizer (`internal/native/generator/customizers/<resource>.go`); call
`generator.IsolateArrayElement` first for a value in an array whose element type may
be shared with a sibling (e.g. `ipRules`/`ipv6Rules`, `resourceAccessRules`). See
GENERATOR.md "Schema Customization Plugins".

## Output Format

Each resource type gets a Go file in `internal/azure/azwise/`. The file:

1. Defines a struct embedding `BaseKnowledge`
2. Satisfies the `ResourceKnowledge` interface (via embedding; override methods only when needed)
3. Provides a `New<Type>()` constructor populating `BaseKnowledge` fields

### Template

```go
package azwise

import "time"

// <Type> provides resource knowledge for <ARM resource type>.
//
// Sources:
//   - AzureRM <resource>_resource.go schema + CRUD functions
//   - <other sources: validate packages, SDK constants>
type <Type> struct {
    BaseKnowledge
}

var _ ResourceKnowledge = (*<Type>)(nil)

func New<Type>() *<Type> {
    return &<Type>{
        BaseKnowledge: BaseKnowledge{
            ResourceType: "<Microsoft.Provider/resourceType>",
            ApiVersions:  []string{"<from azwise_extract schema output>"},
            ForceNew:     []ForceNewRule{{PropertyPath: "name"}},
            SoftDelete:   false,
            TimeoutsConfig: &Timeouts{
                Create: 30 * time.Minute,
                Read:   5 * time.Minute,
                Update: 30 * time.Minute,
                Delete: 30 * time.Minute,
            },
            StringRules:     []StringRule{/* naming + enum rules */},
            IntRules:        []IntRule{},
            FloatRules:      []FloatRule{},
            ArrayRules:      []ArrayRule{},
            SensitiveFields: []string{},
            ComputedFields:  []string{/* read-only ARM properties */},
            DefaultValues:   []DefaultValue{/* {PropertyPath, Value} pairs */},
            RequiredFields:  []string{/* ARM body paths required for creation */},
        },
    }
}
```

Override methods only for non-trivial logic (see `storage_account.go` `CheckForceNew` for conditional SKU zone-migration). **Always call the base method first** — e.g. `s.BaseKnowledge.CheckForceNew(oldBody, newBody)` — so that the declarative rules in `ForceNew`, `StringRules`, etc. are still evaluated. The override adds behavior on top, it does not replace the base.

## Verification checklist

Before finalizing a knowledge file:

1. **Rule type vs field type**: Every `StringRule.AllowedValues` must target a string/enum ARM field. Every `IntRule` must target a numeric ARM field. Cross-check against the ARM SDK model struct in `vendor/github.com/hashicorp/go-azure-sdk/resource-manager/<service>/<api-version>/<resource>/model_*.go`.
2. **Enum completeness**: Use the full ARM SDK enum values from `PossibleValuesFor*()` in `constants.go`, not the AzureRM-restricted subset. AzAPI users send raw ARM values.
3. **ARM path accuracy**: Verify each `PropertyPath` traces correctly through the nested struct chain (e.g., `properties.immutableStorageWithVersioning.immutabilityPolicy.state` not `...immutabilityPeriodSinceCreationInDays`). Check the `json:"..."` tags in the SDK model files.
4. **Computed vs user-settable**: `ComputedFields` must only include fields that are absent from the ARM Create/Update model (truly read-only). Verify against the SDK's `model_*createparameters.go` — if a field is present there, it is settable by AzAPI users and MUST NOT be in `ComputedFields` (even if AzureRM doesn't support it yet). Fields that are optional-but-server-defaulted belong in `DefaultValues`, not `ComputedFields`. `StripComputedFields()` removes these from the body, so a false positive silently discards user-provided values.
5. **Method override base calls**: Every method override (e.g., `func (s *Type) CheckForceNew(...)`) must call `s.BaseKnowledge.<Method>(...)` first. An override that omits the base call silently disables the declarative rules.
6. **Sub-service API separation**: Verify every `PropertyPath` belongs to the resource's own ARM API, not a sub-service API. AzureRM bundles sub-service settings (e.g., `blob_properties`, `share_properties`, `queue_properties` in `azurerm_storage_account`) but ARM manages them as separate resources (`Microsoft.Storage/storageAccounts/blobServices/default`, etc.). Rules for sub-service properties must go in their own knowledge file.

## Registration

Add the constructor call to `RegisterAll()` in `register.go`:

```go
func RegisterAll() {
    Register(NewStorageAccount())
    Register(NewKeyVault())
    Register(New<Type>())  // ← add here
}
```

## Data Sources

### `.release/provider-schema.json`
A pre-built JSON file in the AzureRM repo containing the full schema of every resource — field types, computed/optional/required/forceNew/sensitive flags, nesting, and MaxItems. This is the **primary source** for field flags; use it before grepping Go source.

Call `azwise_extract` with `category=schema` and `resource_name=azurerm_<resource>` to get a classified breakdown (computed, defaults, forceNew, sensitive, required, optional fields). Use `category=schema` without `resource_name` to list all resources (optionally filtered by `service`).

### Go source (`internal/services/<svc>/`)
The source code is needed for details the schema file doesn't capture: validation regexes, enum values, soft-delete behavior, conditional ForceNew logic, and the Terraform→ARM field name mapping.

## Using azwise_extract

The tool supports five modes, each with per-resource and service-wide variants:

1. **Schema mode** (`category=schema`): reads `provider-schema.json` for field flags + API versions from Go imports. Returns classified fields, API version, and a list of nested `blocks`. Use `block=blob_properties` to narrow to a sub-tree.
2. **Automap mode** (`category=automap`): combines schema + d.Set/d.Get mappings + mechanical snake→camelCase to produce ARM paths for all fields. Partitions results into `mapped` (verified from source), `automapped` (mechanical, needs checking), and `skipped` (envelope/provider-internal). Emits a `recommendation` when the resource has >30 fields, suggesting block-by-block processing.
3. **Mapping mode** (`category=mapping`): extracts Terraform→ARM property path mappings from Go source. With `resource_name`, detects property aliases, discovers expand/flatten functions, and reports **unmapped fields**. Use `block=` to filter.
4. **Validation mode** (`category=validation`): extracts `ValidateFunc`/`ValidateDiagFunc` patterns. Returns structured rules: enums, ranges, regexes, UUID, custom. Use `block=` to filter.
5. **Grep mode** (`category=forcenew|sensitive|timeouts|softdelete`): scans Go source for specific code patterns.

Use `category=all` to run all modes at once.

### `block` parameter

Narrows schema/mapping/validation/automap to a nested block sub-tree (e.g. `block=blob_properties` or `block=network_rules.private_link_access`). Use this for large resources where processing everything at once exceeds context limits. The schema output's `blocks` list shows all available block paths.

### Recommended workflow

**Small resources** (≤30 top-level fields, ≤5 blocks):
1. `category=schema` → classified fields + API version + blocks list
2. `category=automap` → ARM paths for all fields
3. `category=validation` → validation rules
4. `category=timeouts` + `category=softdelete` → timeout values + soft-delete
5. Read Go source only for `automapped` fields that need verification
6. Write the Go knowledge file

**Large resources** (>30 fields or >5 blocks, e.g. storage_account):
1. `category=schema` → get the `blocks` list and top-level field count
2. `category=automap` → ARM paths for top-level scalar fields
3. For each block: `category=automap, block=<name>` + `category=validation, block=<name>` → focused extraction
4. Read the expand/flatten function for each block to verify ARM paths
5. `category=timeouts` + `category=softdelete`
6. Aggregate all results into the knowledge file

### Mapping patterns supported
- `d.Set("field", model.Properties.X)` — direct property access
- `d.Set("field", props.X)` — aliased property access (alias auto-detected)
- `d.Set("field", flatten...(model.Properties.X))` — flatten with direct or aliased access
- `d.Set("field", model.Properties.X.Y)` — nested property access (2 levels)
- `d.Set("field", model.X)` / `account.X` — top-level model fields
- `d.Set("field", key.X)` / `attributes.X` — data-plane patterns
- `.Properties.X = d.Get("field")` — Create/Update direction

### What automap CANNOT verify mechanically (LLM must read source)
- **Non-conventional ARM paths**: `customer_managed_key` → `properties.encryption` (not `properties.customerManagedKey`)
- **Intermediate variables**: `vaultUri := model.Properties.VaultUri; d.Set("vault_uri", vaultUri)`
- **Boolean logic**: `publicNetworkAccessEnabled := strings.EqualFold(*props.PublicNetworkAccess, "Enabled")`
- **One-to-many**: one Terraform field expanding into multiple ARM properties
- **Separate API resources**: `blob_properties` comes from BlobService API, not the storage account body

The automap output's `automapped` list shows exactly which fields were mechanically guessed. Focus source reading there.

## Validation

### azwise_validate tool

After generating or modifying knowledge files, run `azwise_validate` to cross-reference rules against the ARM SDK. The tool reads each Go knowledge file, parses all StringRules/IntRules/FloatRules/ComputedFields/ForceNew/DefaultValues, then traces every PropertyPath through the SDK struct chain via `json:"..."` tags.

**Usage:**
- `azwise_validate` — validate all knowledge files
- `azwise_validate resource=storage_account` — validate only `storage_account.go` (and `storage_account_*.go`)

**What it checks:**
1. **Type mismatches**: StringRule on a `*int64` field (should be IntRule), IntRule on a `*string` enum field
2. **Invalid paths**: PropertyPath segments that don't match any `json:"..."` tag in the SDK struct chain
3. **Incomplete enums**: AllowedValues missing values from `PossibleValuesFor*()` in the SDK constants
4. **Extra enum values**: AllowedValues containing values not present in the SDK

**Issue severities:**
- `error` — definite bug, must fix (type mismatches, invalid paths)
- `warning` — needs review (incomplete enums, unresolvable top-level paths)

### azwise-validate agent

The `azwise-validate` agent automates the full validation + fix cycle. Invoke it to validate all knowledge files, triage issues, apply fixes, and verify with `go test`.