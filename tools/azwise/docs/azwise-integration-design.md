# Azwise Integration Design Plan

## Goal

Design and implement the `azwise` package — a per-resource-type knowledge registry that integrates into `terraform-provider-azapi` to surface AzureRM-derived operational knowledge (ForceNew rules, timeouts, naming validation, property constraints, sensitive field paths, soft-delete awareness) during `terraform plan` and `terraform validate`.

## User Decisions

- **Severity**: Errors (strict — block apply for known-bad configurations)
- **Feature flag**: Enabled by default with `disable_resource_knowledge = true` opt-out
- **Scope**: Storage Account + Key Vault in first PR
- **API versioning**: Knowledge entries specify which API versions they apply to; registry lookup takes `(resourceType, apiVersion)`
- **Validation logic location**: All validation/checking logic lives in the `azwise` package; `azapi_resource.go` only calls thin package-level functions
- **Registration pattern**: Each resource file exports a `XxxKnowledge()` function returning `[]ResourceKnowledge`; a central `register.go` calls them all explicitly — no `init()` magic
- **Architecture**: `ResourceKnowledge` interface with methods (`CheckForceNew`, `ValidateName`, etc.) + `BaseKnowledge` struct providing default implementations. Simple resources configure `BaseKnowledge`; complex ones embed and override. Matches existing `customization/DataPlaneResource` interface pattern.

## Background

### Current State of AzAPI

AzAPI treats all 31,175 Azure resource types uniformly at runtime. The only per-resource-type knowledge today lives in:

1. **`skipApiVersions`** (`internal/azure/loader.go`) — 2-entry map blocking broken API versions
2. **`volatileFieldList()`** (`internal/services/common.go`) — 22 field names stripped from output (not per-resource)
3. **`fix.GetWriteOnlyFix()`** (`internal/azure/fix/fix.go`) — hardcoded `userAssignedIdentities` normalization
4. **`customization/`** (`internal/services/customization/`) — 2 data-plane resource customizations via `init()` registration + interface

### AzAPI Lifecycle Hooks (Integration Points)

| Hook | File | What it does today | Azwise opportunity |
|---|---|---|---|
| `ValidateConfig` | `azapi_resource.go:486-525` | Validates `parent_id`, duplicated defs, schema validation | Add naming regex validation, property constraints, sensitive field errors |
| `ModifyPlan` | `azapi_resource.go:527-737` | ForceNew for `name`/`parent_id`/`location`, `IgnoreNoOpChanges`, `replace_triggers_refs`, preflight | Add body-property ForceNew rules |
| `CreateUpdate` | `azapi_resource.go:772-999` | Reads `plan.Timeouts.Create(ctx, 30*time.Minute)` | Override default timeout with per-resource value |
| `Read` | `azapi_resource.go:1001-1100` | Reads `model.Timeouts.Read(ctx, 5*time.Minute)` | Override default timeout |
| `Delete` | `azapi_resource.go:1162-1204` | Reads `model.Timeouts.Delete(ctx, 30*time.Minute)` | Override default timeout; future: soft-delete/purge |

### Key Architectural Facts

- `AzapiResource` is a flat struct with one field: `ProviderData *clients.Client`
- Resource type string is parsed via `utils.GetAzureResourceTypeApiVersion(config.Type.ValueString())` → returns `(azureResourceType, apiVersion, error)` — available in both `ValidateConfig` and `ModifyPlan`
- `name`/`parent_id` use schema-level `RequiresReplace` plan modifiers (always replace). `location` is checked dynamically in `ModifyPlan` (line 669-672)
- Timeouts use `hashicorp/terraform-plugin-framework-timeouts` with hardcoded defaults: Create/Update/Delete 30m, Read 5m. User-configured values always win over defaults.
- `azapi_update_resource` has its own `ModifyPlan` but much simpler — no ForceNew, no location check. It doesn't need azwise integration.

---

## Design

### Package Location

```
tools/azwise/docs/                         # Design docs (dev-time)
  azurerm-knowledge-transfer-analysis.md   # Analysis document
  azwise-integration-design.md             # This file

internal/azure/azwise/                     # Runtime Go package (ships in binary)
  azwise.go                                # Interface, BaseKnowledge, registry, package-level functions
  helpers.go                               # extractNestedValue, extractStringValue, etc.
  register.go                              # RegisterAll() — explicit registration of all resources
  azwise_test.go                           # Unit tests for registry + interface + BaseKnowledge
  storage.go                               # StorageKnowledge() — embed + override for custom ForceNew
  storage_test.go                          # Storage-specific tests
  keyvault.go                              # KeyVaultKnowledge() — BaseKnowledge directly
  keyvault_test.go                         # KeyVault-specific tests
```

### Interface

```go
package azwise

import "time"

// ResourceKnowledge is the interface each resource type implements.
// Simple resources use BaseKnowledge directly; complex ones embed it and override methods.
type ResourceKnowledge interface {
    // GetResourceType returns the Azure resource type (e.g. "Microsoft.Storage/storageAccounts").
    GetResourceType() string

    // GetApiVersions returns API versions this knowledge applies to.
    // Empty means all versions.
    GetApiVersions() []string

    // CheckForceNew returns true if the body change requires resource replacement.
    CheckForceNew(oldBody, newBody map[string]interface{}) bool

    // ValidateName returns an error if the name violates naming constraints.
    ValidateName(name string) error

    // ValidateProperties returns errors for body properties that violate constraints.
    ValidateProperties(body map[string]interface{}) []error

    // TimeoutDefault returns the recommended timeout for the given operation,
    // or fallback if no override. operation: "create"|"read"|"update"|"delete".
    TimeoutDefault(operation string, fallback time.Duration) time.Duration

    // GetSensitiveFields returns ARM JSON paths containing secrets, or nil.
    GetSensitiveFields() []string

    // IsSoftDelete returns whether the resource supports soft-delete.
    IsSoftDelete() bool
}
```

### Supporting Types

```go
// ForceNewRule describes a body property that triggers resource replacement.
type ForceNewRule struct {
    PropertyPath string // dot-separated ARM JSON path (e.g. "sku.name")
}

// StringRule validates a string value — resource name (PropertyPath == "") or body property.
// Supports regex pattern matching, length constraints, and enum (AllowedValues).
type StringRule struct {
    PropertyPath  string   // dot-separated ARM JSON path; empty = resource name attribute
    Regex         string   // pattern the value must match; empty = no pattern check
    MinLength     int      // 0 = no minimum
    MaxLength     int      // 0 = no maximum
    AllowedValues []string // if non-empty, value must be one of these (case-insensitive)
    Message       string   // human-readable explanation
}

// PropertyRule validates a numeric body property.
type PropertyRule struct {
    PropertyPath string   // dot-separated ARM JSON path
    MinValue     *float64 // nil = no minimum
    MaxValue     *float64 // nil = no maximum
    Message      string
}

// Timeouts overrides default timeout values. Zero = use provider default.
type Timeouts struct {
    Create time.Duration
    Read   time.Duration
    Update time.Duration
    Delete time.Duration
}
```

### BaseKnowledge (default implementation)

`BaseKnowledge` implements `ResourceKnowledge` from data fields. Simple resources use it directly; complex ones embed and override.

```go
type BaseKnowledge struct {
    ResourceType    string
    ApiVersions     []string
    ForceNew        []ForceNewRule
    TimeoutsConfig  *Timeouts
    SoftDelete      bool
    StringRules     []StringRule     // name + body string property validation
    PropertyRules   []PropertyRule   // numeric body property validation
    SensitiveFields []string
}

func (b *BaseKnowledge) GetResourceType() string    { return b.ResourceType }
func (b *BaseKnowledge) GetApiVersions() []string    { return b.ApiVersions }
func (b *BaseKnowledge) IsSoftDelete() bool           { return b.SoftDelete }
func (b *BaseKnowledge) GetSensitiveFields() []string { return b.SensitiveFields }

func (b *BaseKnowledge) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
    for _, rule := range b.ForceNew {
        oldVal := extractNestedValue(oldBody, rule.PropertyPath)
        newVal := extractNestedValue(newBody, rule.PropertyPath)
        if !reflect.DeepEqual(oldVal, newVal) {
            return true
        }
    }
    return false
}

// ValidateName checks StringRules with empty PropertyPath against the resource name.
func (b *BaseKnowledge) ValidateName(name string) error {
    for i := range b.StringRules {
        if b.StringRules[i].PropertyPath != "" { continue }
        if err := b.StringRules[i].Validate(name); err != nil { return err }
    }
    return nil
}

// ValidateProperties checks both numeric PropertyRules and string StringRules against body.
func (b *BaseKnowledge) ValidateProperties(body map[string]interface{}) []error {
    var errs []error
    for i := range b.PropertyRules {
        if err := b.PropertyRules[i].Validate(body); err != nil {
            errs = append(errs, err)
        }
    }
    for i := range b.StringRules {
        if b.StringRules[i].PropertyPath == "" { continue }
        v := extractNestedValue(body, b.StringRules[i].PropertyPath)
        if v == nil { continue }
        s, ok := v.(string)
        if !ok { continue }
        if err := b.StringRules[i].Validate(s); err != nil {
            errs = append(errs, err)
        }
    }
    return errs
}

func (b *BaseKnowledge) TimeoutDefault(operation string, fallback time.Duration) time.Duration {
    if b.TimeoutsConfig == nil { return fallback }
    switch operation {
    case "create":
        if b.TimeoutsConfig.Create > 0 { return b.TimeoutsConfig.Create }
    case "read":
        if b.TimeoutsConfig.Read > 0 { return b.TimeoutsConfig.Read }
    case "update":
        if b.TimeoutsConfig.Update > 0 { return b.TimeoutsConfig.Update }
    case "delete":
        if b.TimeoutsConfig.Delete > 0 { return b.TimeoutsConfig.Delete }
    }
    return fallback
}
```

### Registry (API-Version-Aware)

```go
type registryEntry []ResourceKnowledge

var registry = map[string]registryEntry{}

// Register adds a ResourceKnowledge implementation to the registry.
func Register(k ResourceKnowledge) {
    key := strings.ToLower(k.GetResourceType())
    registry[key] = append(registry[key], k)
}

// Get returns the best-matching ResourceKnowledge for a resource type + API version.
// Priority: version-specific match > catch-all (empty ApiVersions) > nil.
func Get(resourceType, apiVersion string) ResourceKnowledge {
    key := strings.ToLower(resourceType)
    entries := registry[key]
    var fallback ResourceKnowledge
    for _, k := range entries {
        if len(k.GetApiVersions()) == 0 {
            fallback = k
            continue
        }
        for _, v := range k.GetApiVersions() {
            if v == apiVersion {
                return k
            }
        }
    }
    return fallback
}
```

### Package-Level Functions (called from azapi_resource.go)

Thin functions that look up the registry and delegate to the interface:

```go
func CheckForceNew(resourceType, apiVersion string, oldBody, newBody map[string]interface{}) bool {
    EnsureRegistered()
    k := Get(resourceType, apiVersion)
    if k == nil { return false }
    return k.CheckForceNew(oldBody, newBody)
}

func ValidateName(resourceType, apiVersion, name string) error {
    EnsureRegistered()
    k := Get(resourceType, apiVersion)
    if k == nil { return nil }
    return k.ValidateName(name)
}

func ValidateProperties(resourceType, apiVersion string, body map[string]interface{}) []error {
    EnsureRegistered()
    k := Get(resourceType, apiVersion)
    if k == nil { return nil }
    return k.ValidateProperties(body)
}

func TimeoutDefault(resourceType, apiVersion, operation string, fallback time.Duration) time.Duration {
    EnsureRegistered()
    k := Get(resourceType, apiVersion)
    if k == nil { return fallback }
    return k.TimeoutDefault(operation, fallback)
}

func GetSensitiveFields(resourceType, apiVersion string) []string {
    EnsureRegistered()
    k := Get(resourceType, apiVersion)
    if k == nil { return nil }
    return k.GetSensitiveFields()
}
```

### Integration in azapi_resource.go (Thin Calls Only)

`azapi_resource.go` makes minimal calls. All logic stays in the azwise package.

#### ModifyPlan (after line ~672):

```go
if r.ProviderData != nil && !r.ProviderData.Features.DisableResourceKnowledge {
    if state != nil && dynamic.IsFullyKnown(plan.Body) {
        oldBody := make(map[string]interface{})
        newBody := make(map[string]interface{})
        if err := unmarshalBody(state.Body, &oldBody); err == nil {
            if err := unmarshalBody(plan.Body, &newBody); err == nil {
                if azwise.CheckForceNew(azureResourceType, apiVersion, oldBody, newBody) {
                    response.RequiresReplace.Append(path.Root("body"))
                }
            }
        }
    }
}
```

#### CreateUpdate (line ~794):

```go
createDefault := azwise.TimeoutDefault(azureResourceType, apiVersion, "create", 30*time.Minute)
timeout, diags = plan.Timeouts.Create(ctx, createDefault)
```

#### Read (line ~1007) / Delete (line ~1176): same pattern with `"read"` / `"delete"`.

#### ValidateConfig (after line ~518):

```go
if r.ProviderData != nil && !r.ProviderData.Features.DisableResourceKnowledge {
    azureResourceType, apiVersion, _ := utils.GetAzureResourceTypeApiVersion(resourceType)
    if !config.Name.IsNull() && !config.Name.IsUnknown() {
        if err := azwise.ValidateName(azureResourceType, apiVersion, config.Name.ValueString()); err != nil {
            response.Diagnostics.AddError("Invalid resource name",
                fmt.Sprintf("The name %q is not valid for %s: %s",
                    config.Name.ValueString(), azureResourceType, err.Error()))
        }
    }
    if !config.Body.IsNull() && !config.Body.IsUnknown() {
        var body map[string]interface{}
        if err := unmarshalBody(config.Body, &body); err == nil {
            for _, e := range azwise.ValidateProperties(azureResourceType, apiVersion, body) {
                response.Diagnostics.AddError("Invalid property value", e.Error())
            }
        }
    }
    if fields := azwise.GetSensitiveFields(azureResourceType, apiVersion); len(fields) > 0 {
        if !config.Body.IsNull() && !config.Body.IsUnknown() && config.SensitiveBody.IsNull() {
            response.Diagnostics.AddError("Sensitive properties should use sensitive_body",
                fmt.Sprintf("%s has known sensitive fields (%s). Move these from body to sensitive_body.",
                    azureResourceType, strings.Join(fields, ", ")))
        }
    }
}
```

### Per-Resource Registration

#### register.go (central, explicit)

```go
package azwise

func RegisterAll() {
    for _, k := range StorageKnowledge() { Register(k) }
    for _, k := range KeyVaultKnowledge() { Register(k) }
    // Future: for _, k := range ComputeKnowledge() { Register(k) }
}
```

#### storage.go — embed BaseKnowledge, override CheckForceNew

```go
package azwise

import (
    "strings"
    "time"
)

type storageAccountKnowledge struct{ *BaseKnowledge }

// CheckForceNew overrides BaseKnowledge — only cross-zone SKU migration triggers replacement.
func (s *storageAccountKnowledge) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
    oldSku := strings.ToUpper(extractStringValue(oldBody, "sku.name"))
    newSku := strings.ToUpper(extractStringValue(newBody, "sku.name"))
    if oldSku == "" || newSku == "" || oldSku == newSku {
        return false
    }
    zonal := map[string]bool{"STANDARD_ZRS": true, "STANDARD_GZRS": true, "STANDARD_RAGZRS": true}
    nonZonal := map[string]bool{"STANDARD_LRS": true, "STANDARD_GRS": true, "STANDARD_RAGRS": true}
    return (zonal[oldSku] && nonZonal[newSku]) || (nonZonal[oldSku] && zonal[newSku])
}

func StorageKnowledge() []ResourceKnowledge {
    return []ResourceKnowledge{
        &storageAccountKnowledge{
            BaseKnowledge: &BaseKnowledge{
                ResourceType: "Microsoft.Storage/storageAccounts",
                TimeoutsConfig: &Timeouts{
                    Create: 60 * time.Minute, Read: 5 * time.Minute,
                    Update: 60 * time.Minute, Delete: 60 * time.Minute,
                },
                StringRules: []StringRule{
                    {Regex: `^[a-z0-9]{3,24}$`, MinLength: 3, MaxLength: 24, Message: "must be lowercase alphanumeric, 3-24 characters"},
                    {PropertyPath: "sku.name", AllowedValues: []string{"Standard_LRS", "Standard_GRS", "Standard_RAGRS", "Standard_ZRS", "Standard_GZRS", "Standard_RAGZRS", "Premium_LRS", "Premium_ZRS"}, Message: "must be a valid storage SKU"},
                    {PropertyPath: "kind", AllowedValues: []string{"BlobStorage", "BlockBlobStorage", "FileStorage", "Storage", "StorageV2"}, Message: "must be a valid storage kind"},
                    {PropertyPath: "properties.accessTier", AllowedValues: []string{"Hot", "Cool", "Premium"}, Message: "must be Hot, Cool, or Premium"},
                },
                SensitiveFields: []string{"properties.primaryAccessKey", "properties.secondaryAccessKey"},
            },
        },
    }
}
```

#### keyvault.go — BaseKnowledge directly (no override needed)

```go
package azwise

func KeyVaultKnowledge() []ResourceKnowledge {
    return []ResourceKnowledge{
        &BaseKnowledge{
            ResourceType: "Microsoft.KeyVault/vaults",
            ForceNew:     []ForceNewRule{{PropertyPath: "name"}},
            SoftDelete:   true,
            StringRules: []StringRule{
                {Regex: `^[a-zA-Z][a-zA-Z0-9-]{1,22}[a-zA-Z0-9]$`, MinLength: 3, MaxLength: 24, Message: "must start with a letter, end with letter/digit, alphanumeric or hyphens, 3-24 characters"},
                {PropertyPath: "properties.sku.name", AllowedValues: []string{"standard", "premium"}, Message: "must be standard or premium"},
            },
            PropertyRules: []PropertyRule{{
                PropertyPath: "properties.softDeleteRetentionInDays",
                MinValue: floatPtr(7), MaxValue: floatPtr(90),
                Message: "soft delete retention must be between 7 and 90 days",
            }},
        },
    }
}
```

### Initialization

```go
var once sync.Once

func EnsureRegistered() {
    once.Do(RegisterAll)
}
```

Each package-level function calls `EnsureRegistered()` at the top.

---

## Key Design Decisions

### 1. Errors, not warnings

Azwise-derived validation emits **errors** that block apply. Strict enforcement.

### 2. Opt-out feature flag

`disable_resource_knowledge = true` top-level provider attribute. All azwise calls gated on `r.ProviderData != nil && !r.ProviderData.Features.DisableResourceKnowledge`.

### 3. API-version-aware registry

`Get(resourceType, apiVersion)` returns: version-specific match > catch-all > nil.

### 4. Interface + BaseKnowledge pattern

- `ResourceKnowledge` interface defines the contract (matches existing `customization/DataPlaneResource` pattern)
- `BaseKnowledge` struct provides default implementations from data fields
- Simple resources (Key Vault) → use `BaseKnowledge` directly, zero custom code
- Complex resources (Storage Account) → embed `BaseKnowledge`, override one method (`CheckForceNew`)
- Future extensibility: override `ValidateName`, `ValidateProperties`, or any other method for resource-specific logic without touching the registry or base implementation

### 5. Logic lives in azwise, not azapi_resource.go

`azapi_resource.go` calls only package-level functions (`azwise.CheckForceNew(...)`, etc.). All validation logic, body traversal, regex, numeric comparison live in `azwise`.

### 6. Explicit registration, no init()

`register.go` has `RegisterAll()` calling each resource's exported function. `EnsureRegistered()` wraps in `sync.Once`. No `init()` magic.

### 7. `azapi_update_resource` excluded

No ForceNew, no create/delete lifecycle, no naming validation. Not worth integration.

### 8. Two resource types in first PR

Storage Account (conditional ForceNew via override) and Key Vault (BaseKnowledge directly).

---

## Files Created/Modified

### New Files
| File | Description |
|---|---|
| `internal/azure/azwise/azwise.go` | `ResourceKnowledge` interface, `BaseKnowledge` struct + methods, `StringRule.Validate`, `PropertyRule.Validate`, registry (`Register`, `Get`), `EnsureRegistered()`, package-level functions |
| `internal/azure/azwise/helpers.go` | `extractNestedValue`, `extractStringValue`, `toFloat64`, `floatPtr` |
| `internal/azure/azwise/register.go` | `RegisterAll()` |
| `internal/azure/azwise/azwise_test.go` | Unit tests: interface compliance, BaseKnowledge defaults, registry lookup, API version matching |
| `internal/azure/azwise/storage.go` | `storageAccountKnowledge` (embed + override) + `StorageKnowledge()` |
| `internal/azure/azwise/storage_test.go` | SKU zone migration matrix, naming, body string property (sku, kind, accessTier) validation |
| `internal/azure/azwise/keyvault.go` | `KeyVaultKnowledge()` using `BaseKnowledge` directly |
| `internal/azure/azwise/keyvault_test.go` | Name ForceNew, naming, softDeleteRetentionDays range |

### Modified Files
| File | Change |
|---|---|
| `internal/services/azapi_resource.go` | Import `azwise`. Thin calls in `ModifyPlan`, `CreateUpdate`/`Read`/`Delete`, `ValidateConfig`. All gated on feature flag. |
| `internal/features/user_flags.go` | Add `DisableResourceKnowledge bool` field, default `false`. |
| `internal/provider/provider.go` | Add `disable_resource_knowledge` to provider schema + `providerData` model; wire to `UserFeatures`. |
