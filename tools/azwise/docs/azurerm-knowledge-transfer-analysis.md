# Analysis: Leveraging AzureRM Provider Knowledge to Enhance AzAPI

## Executive Summary

The AzureRM provider (`terraform-provider-azurerm`) encodes **operational knowledge** about 1,134 Azure resources across 131 service packages — knowledge accumulated from years of real-world usage and Azure API quirks. The AzAPI provider validates resources against embedded Bicep type schemas (31,175 resource types), but treats all Azure resources uniformly at runtime.

This report catalogs the categories of knowledge in AzureRM, provides concrete examples for each, and assesses feasibility and benefit of transferring it to AzAPI.

---

## 1. ForceNew (Replace-on-Change) Rules

**Scale**: 4,954 `ForceNew` annotations across 1,319 resource files.

AzureRM marks specific properties as requiring resource destruction and recreation when changed. Without this, Terraform would attempt an in-place update that Azure rejects.

### Example 1: Key Vault — `name` and `location`

```go
// internal/services/keyvault/key_vault_resource.go:66-71
"name": {
    Type:         pluginsdk.TypeString,
    Required:     true,
    ForceNew:     true,    // ← cannot rename a Key Vault in-place
    ValidateFunc: validate.VaultName,
},
```

Azure's Key Vault API does not support renaming. If a user changes the `name`, Terraform must destroy the old vault and create a new one. AzAPI would instead attempt a PUT with the new name, receiving a cryptic 400/409 error.

### Example 2: Storage Account — Conditional ForceNew on `account_replication_type`

```go
// internal/services/storage/storage_account_resource.go:1183-1197
pluginsdk.ForceNewIfChange("account_replication_type", func(ctx context.Context, old, new, meta interface{}) bool {
    newAccRep := strings.ToUpper(new.(string))
    switch strings.ToUpper(old.(string)) {
    case "LRS", "GRS", "RAGRS":
        if newAccRep == "GZRS" || newAccRep == "RAGZRS" || newAccRep == "ZRS" {
            return true  // ← cross-zone migration requires recreation
        }
    case "ZRS", "GZRS", "RAGZRS":
        if newAccRep == "LRS" || newAccRep == "GRS" || newAccRep == "RAGRS" {
            return true  // ← zone-to-non-zone requires recreation
        }
    }
    return false
}),
```

This is **conditional** ForceNew — you can switch `LRS → GRS` in place, but switching `LRS → ZRS` requires recreation because Azure cannot migrate data between zonal and non-zonal replication in place.

### Impact for AzAPI

AzAPI has no ForceNew knowledge. When a user changes a property that requires replacement, AzAPI attempts an update, which either:
- Fails with an opaque API error at apply time (not plan time)
- Silently succeeds but creates a new resource with the old still existing (orphan)

**Benefit**: Plan-time "must be replaced" signals would prevent apply-time surprises.

---

## 2. Soft Delete / Purge Handling

**Scale**: 36 resource files handle soft-delete/purge logic.

Some Azure resources enter a "soft-deleted" state on DELETE. They continue to exist and block creation of new resources with the same name until purged (or auto-purged after a retention period).

### Example 1: Key Vault — Feature-flag controlled purge

```go
// internal/services/keyvault/key_vault_resource.go:858-883 (simplified)
if meta.(*clients.Client).Features.KeyVault.PurgeSoftDeleteOnDestroy && softDeleteEnabled {
    deletedVaultId := vaults.NewDeletedVaultID(id.SubscriptionId, location, id.VaultName)

    // KeyVaults with Purge Protection Enabled cannot be deleted unless done by Azure
    if purgeProtectionEnabled {
        log.Printf("[DEBUG] The Key Vault %q has Purge Protection Enabled and was deleted on %q. "+
            "Azure will purge this on %q", id.VaultName, deletedInfo.deleteDate, deletedInfo.purgeDate)
        return nil
    }

    log.Printf("[DEBUG] KeyVault %q marked for purge - executing purge", id.VaultName)
    if err := client.PurgeDeletedThenPoll(ctx, deletedVaultId); err != nil {
        return fmt.Errorf("purging %s: %+v", *id, err)
    }
}
```

AzureRM: (1) deletes the vault, (2) checks if purge protection is enabled, (3) if not, purges it. Also handles the case where create discovers a soft-deleted vault and recovers it.

### Example 2: Cognitive Services Account — Delete + Purge + Wait for subnet cleanup

```go
// internal/services/cognitive/cognitive_account_resource.go:747-758
if err := accountsClient.AccountsDeleteThenPoll(ctx, *id); err != nil {
    return fmt.Errorf("deleting %s: %+v", *id, err)
}

if meta.(*clients.Client).Features.CognitiveAccount.PurgeSoftDeleteOnDestroy {
    log.Printf("[DEBUG] Purging %s..", *id)
    if err := deletedAccountsClient.DeletedAccountsPurgeThenPoll(ctx, deletedAccountId); err != nil {
        return fmt.Errorf("purging %s: %+v", *id, err)
    }
}
```

Additionally, Cognitive Services with network injection must wait for Azure to remove the Service Association Link from the subnet after deletion — a workaround for a known service issue.

### Impact for AzAPI

AzAPI sends a single DELETE. If the resource has soft-delete enabled:
- `terraform destroy` succeeds, but the resource is not truly gone
- `terraform apply` on a new resource with the same name fails: "resource already exists"
- User must manually purge via CLI or wait for auto-purge (7-90 days)

**Benefit**: A table of soft-delete-enabled resource types would allow AzAPI to automatically issue purge calls or warn users.

---

## 3. Custom Polling / Async Operations

**Scale**: 18 custom poller directories, 159 files with explicit `StateRefreshFunc` / `WaitForState`.

Some Azure resources have non-standard async patterns — the standard LRO (Long Running Operation) header-based polling doesn't apply, or the API returns "succeeded" before the resource is actually usable.

### Example 1: Synapse Managed Private Endpoint — Custom provisioning state poller

```go
// internal/services/synapse/custompollers/synapse_managed_private_endpoint_create_poller.go:26-51
func (s synapseManagedPrivateEndpointCreatePoller) Poll(ctx context.Context) (*pollers.PollResult, error) {
    resp, err := s.client.Get(ctx, s.id)
    if err != nil {
        return nil, err
    }
    if resp.Model == nil || resp.Model.Properties == nil || resp.Model.Properties.ProvisioningState == nil {
        return nil, fmt.Errorf("checking `provisioningState` for %s", s.id)
    }
    switch *resp.Model.Properties.ProvisioningState {
    case string(pollers.PollingStatusSucceeded):
        return &pollers.PollResult{Status: pollers.PollingStatusSucceeded}, nil
    case string(pollers.PollingStatusFailed):
        return nil, fmt.Errorf("provisioningState was `%s`", pollers.PollingStatusFailed)
    case string(pollers.PollingStatusCancelled):
        return nil, fmt.Errorf("provisioningState was `%s`", pollers.PollingStatusCancelled)
    }
    return &pollers.PollResult{PollInterval: 10 * time.Second, Status: pollers.PollingStatusInProgress}, nil
}
```

The Synapse data-plane API doesn't return standard Azure LRO headers. AzureRM implements a custom poller that GETs the resource and checks `provisioningState` until it reaches a terminal state.

### Example 2: Key Vault — Polling DNS availability after creation

```go
// internal/services/keyvault/key_vault_resource.go:890-910
func keyVaultRefreshFunc(vaultUri string) pluginsdk.StateRefreshFunc {
    return func() (interface{}, string, error) {
        client := &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}}
        conn, err := client.Get(vaultUri)
        if err != nil {
            return nil, "pending", fmt.Errorf("connecting to %q: %s", vaultUri, err)
        }
        defer conn.Body.Close()
        return "available", "available", nil
    }
}
```

After ARM reports Key Vault creation as "Succeeded", the vault's DNS endpoint (`https://<name>.vault.azure.net`) may not be resolvable for several seconds. AzureRM polls the endpoint directly before attempting data-plane operations (creating secrets, keys, etc.).

### Impact for AzAPI

AzAPI relies on the generic Azure LRO polling mechanism. When that's insufficient:
- Resources may appear created but aren't usable yet (subsequent operations fail)
- Users must configure manual `retry` blocks with `error_message_regex` patterns

**Benefit**: Per-resource polling configuration (interval, field to check, DNS readiness) as metadata.

---

## 4. Diff Suppression

**Scale**: 221 files with `DiffSuppressFunc` implementations.

Azure APIs sometimes return values that differ from what was submitted but are semantically identical. Without suppression, Terraform shows phantom diffs on every plan.

### Example 1: Key Vault Key — Ignore version in key reference

```go
// internal/services/keyvault/suppress/key_vault_key.go:11-23
func DiffSuppressIgnoreKeyVaultKeyVersion(k, old, new string, _ *pluginsdk.ResourceData) bool {
    oldKey, err := keyvault.ParseNestedItemID(old, keyvault.VersionTypeAny, keyvault.NestedItemTypeAny)
    if err != nil {
        return false
    }
    newKey, err := keyvault.ParseNestedItemID(new, keyvault.VersionTypeAny, keyvault.NestedItemTypeAny)
    if err != nil {
        return false
    }
    return (oldKey.KeyVaultBaseURL == newKey.KeyVaultBaseURL) && (oldKey.Name == newKey.Name)
}
```

When a Key Vault key auto-rotates, its version changes. Resources referencing that key (e.g., storage encryption) would show a diff every plan. This suppresses the version component.

### Example 2: Virtual Network Gateway — Case-insensitive subnet ID

```go
// internal/services/network/virtual_network_gateway_resource.go:164
"subnet_id": {
    ...
    DiffSuppressFunc: suppress.CaseDifference,
},
```

Azure normalizes resource IDs with different casing than what the user submitted (e.g., `/subscriptions/...` vs `/Subscriptions/...`). Without suppression, every plan shows a change.

### Impact for AzAPI

AzAPI offers `ignore_casing = true` (resource-wide) and Terraform's `lifecycle { ignore_changes }`. However:
- `ignore_casing` is all-or-nothing per resource
- Users don't know which fields need it without trial and error
- Semantic suppressions (like ignoring key versions) have no equivalent

**Benefit**: Per-property diff suppression metadata (casing, version stripping) would reduce phantom diffs out of the box.

---

## 5. Custom Timeouts

**Scale**: 3,022 `DefaultTimeout` annotations across resource implementations.

AzureRM sets per-resource, per-operation timeouts based on real-world behavior.

### Example 1: Storage Account — 60-minute create/update/delete

```go
// internal/services/storage/storage_account_resource.go:88-93
Timeouts: &pluginsdk.ResourceTimeout{
    Create: pluginsdk.DefaultTimeout(60 * time.Minute),
    Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
    Update: pluginsdk.DefaultTimeout(60 * time.Minute),
    Delete: pluginsdk.DefaultTimeout(60 * time.Minute),
},
```

Storage account operations (especially with network rules, encryption, and geo-replication) can take substantial time.

### Impact for AzAPI

AzAPI uses Terraform's default timeouts (typically 30 or 60 minutes depending on operation type). Without resource-specific tuning:
- Short-lived resources may waste time waiting on generous defaults
- Long-lived resources (large VM scale sets) may timeout prematurely

**Benefit**: A timeout lookup table would improve both correctness and user experience.

---

## 6. Property Validation (Naming Rules)

**Scale**: 101 `validate/` directories with custom validators.

AzureRM validates resource names and property values at plan time with resource-specific rules.

### Example 1: Storage Account name — lowercase alphanumeric, 3-24 chars

```go
// internal/services/storage/validate/storage_account_name.go:11-19
func StorageAccountName(v interface{}, _ string) (warnings []string, errors []error) {
    input := v.(string)
    if !regexp.MustCompile(`\A([a-z0-9]{3,24})\z`).MatchString(input) {
        errors = append(errors, fmt.Errorf(
            "name (%q) can only consist of lowercase letters and numbers, "+
            "and must be between 3 and 24 characters long", input))
    }
    return warnings, errors
}
```

Storage account names are globally unique DNS labels — only lowercase letters and digits, 3-24 characters.

### Example 2: Key Vault name — alphanumeric + hyphens, no consecutive hyphens

```go
// internal/services/keyvault/validate/vault_name.go:12-27
func VaultName(v interface{}, k string) (warnings []string, errors []error) {
    value := v.(string)
    if matched := regexp.MustCompile(`^[a-zA-Z0-9-]{3,24}$`).Match([]byte(value)); !matched {
        errors = append(errors, fmt.Errorf(
            "%q may only contain alphanumeric characters and dashes "+
            "and must be between 3-24 chars", k))
    }
    if matched2 := regexp.MustCompile(`^[a-zA-Z].*[a-zA-Z0-9]$`).Match([]byte(value)); !matched2 {
        errors = append(errors, fmt.Errorf(
            "%q must start with a letter and end with a letter or number", k))
    }
    if strings.Contains(value, "--") {
        errors = append(errors, fmt.Errorf(
            "%q cannot contain consecutive hyphens (\"--\")", k))
    }
    return warnings, errors
}
```

Key Vault names have three rules: character set, start/end constraints, and no consecutive hyphens.

### Impact for AzAPI

AzAPI validates property **types** (string vs number vs object) and **required/optional** status via Bicep schemas, but not naming rules. Users discover naming violations only at apply time from Azure API errors like: `"The Vault name must be between 3-24 alphanumeric characters."`.

**Benefit**: Representable as `{regex, minLength, maxLength}` tuples per resource type's `name` property. Plan-time validation catches errors before any API call.

---

## 7. API Workarounds / Cross-Field Constraints (CustomizeDiff)

**Scale**: 175 files with `CustomizeDiff` logic.

AzureRM encodes cross-field validation rules and conditional logic that Azure's API doesn't enforce until apply time.

### Example 1: Storage Account — `account_kind` downgrade forces recreation

```go
// internal/services/storage/storage_account_resource.go:1102-1114
if d.HasChange("account_kind") {
    accountKind, changedKind := d.GetChange("account_kind")
    if accountKind != "" {
        if accountKind != string(storageaccounts.KindStorage) && changedKind != string(storageaccounts.KindStorageVTwo) {
            log.Printf("[DEBUG] recreate storage account, can't be migrated from %q to %q", accountKind, changedKind)
            d.ForceNew("account_kind")
        }
    }
}
```

You can upgrade from `Storage` → `StorageV2`, but cannot downgrade or change between other kinds without recreation. This is undocumented in the API spec.

### Example 2: Storage Account — Immutability policy state transitions

```go
// internal/services/storage/storage_account_resource.go:1151-1173
if d.HasChange("immutability_policy.0.state") {
    old, new := d.GetChange("immutability_policy.0.state")
    // Initial value can be either "Disabled" or "Unlocked"
    if old == "" && (new.(string) != "Disabled" && new.(string) != "Unlocked") {
        return fmt.Errorf("initial value of `immutability_policy.0.state` can be either Disabled or Unlocked")
    }
    // Only "Unlocked" state can be updated to "Locked"
    if new.(string) == "Locked" && old.(string) != "Unlocked" {
        return fmt.Errorf("`immutability_policy.0.state` can only be set to Locked from Unlocked")
    }
    // Once "Locked", can't be changed.
    if old.(string) == "Locked" {
        d.ForceNew("immutability_policy.0.state")
    }
}
```

This encodes a state machine: `Disabled` → `Unlocked` → `Locked` (terminal). The API only returns an error *after* you attempt the invalid transition.

### Impact for AzAPI

AzAPI has no cross-field validation or state transition knowledge. Users hitting these constraints see apply-time errors like `"InvalidStorageAccountImmutabilityPolicy"` with no guidance.

**Benefit**: State machine metadata and cross-field rules could produce better plan-time errors.

---

## 8. No-Downtime Operations

**Scale**: Specialized logic in compute package.

AzureRM implements complex decision trees to determine if an operation can be performed without service interruption.

### Example: Managed Disk — Resize without stopping VM

```go
// internal/services/compute/no_downtime_resize.go:27-66
func determineIfDataDiskSupportsNoDowntimeResize(disk *disks.Disk, requiresDetaching bool) bool {
    // Only supported for data disks (not OS disks)
    // Not supported for shared disks (maxShares > 1)
    // Not supported if requires detaching
    ...
}

func determineIfDataDiskRequiresDetaching(disk *disks.Disk, oldSizeGb, newSizeGb int) bool {
    // Premium SSD v2 and Ultra Disks: no limit
    // Other disk types: can't cross the 4 TiB boundary without detaching
    diskTypeIsSupported := ... PremiumV2LRS || UltraSSDLRS
    if !diskTypeIsSupported && oldSizeGb < 4096 && newSizeGb >= 4096 {
        return true
    }
    return false
}
```

Then it checks if the VM SKU supports no-downtime resize by querying the SKU capabilities API for `EphemeralOSDiskSupported`, `PremiumIO`, or `HyperVGenerations` containing "V2".

### Impact for AzAPI

AzAPI sends a PATCH to resize a disk. Depending on conditions, Azure may:
- Resize in place (no downtime)
- Require VM deallocation first (operation fails with error)
- Require disk detachment (silent data risk)

**Benefit**: This is hard to encode as static metadata (requires runtime queries to SKU API). Best suited for documentation/warnings rather than automatic enforcement.

---

## 9. Sensitive / Write-Only Properties

**Scale**: 1,004 `Sensitive: true` annotations across 376 resource files.

AzureRM marks properties that contain secrets, keys, or credentials as `Sensitive`, preventing them from appearing in plan output, state diffs, or logs.

### Example: Storage Account — Access Keys

```go
// internal/services/storage/storage_account_resource.go
"primary_access_key": {
    Type:      pluginsdk.TypeString,
    Computed:  true,
    Sensitive: true,
},
"secondary_access_key": {
    Type:      pluginsdk.TypeString,
    Computed:  true,
    Sensitive: true,
},
```

Access keys, connection strings, and passwords are automatically hidden from plan output. Users see `(sensitive value)` instead of the raw secret.

### Impact for AzAPI

AzAPI treats all body properties uniformly. Users must manually use `sensitive_body` to prevent secrets from appearing in plan output and state. Without per-property sensitivity metadata:
- Secrets may appear in `terraform plan` output or CI logs
- Users must discover which fields are sensitive through trial and error
- No warning is given when a secret-bearing property is placed in `body` rather than `sensitive_body`

**Benefit**: A per-resource-type list of sensitive property paths would allow AzAPI to warn users (or automatically route) when secret-bearing fields are used in `body` instead of `sensitive_body`.

---

## Feasibility Summary

| Category | Extractable? | Delivery Mechanism | Priority |
|---|---|---|---|
| ForceNew rules | ✅ Low effort | Generated JSON lookup table | **P0** |
| Soft-delete resources | ✅ Low effort | Resource type boolean flag | **P1** |
| Timeout defaults | ✅ Low effort | Generated JSON lookup table | **P1** |
| Naming validation | ✅ Medium effort | `{regex, min, max}` tuples | **P2** |
| Sensitive properties | ✅ Medium effort | Per-property path list | **P2** |
| Cross-field constraints | ⚠️ Hard | State machine DSL or hard-coded | **P3** |
| Diff suppression | ⚠️ Hard | Per-property flags (casing, partial) | **P3** |
| Custom pollers | ❌ Not generalizable | Stay with user-configured `retry` | Low |
| No-downtime operations | ❌ Requires runtime queries | Documentation only | Low |

---

## Key Challenges

1. **Property path mapping**: AzureRM uses `snake_case` Terraform fields (`account_replication_type`); AzAPI uses Azure ARM `camelCase` JSON paths (`properties.encryption.keySource`). A mapping table (or convention-based conversion) is required.

2. **Conditional rules**: Much of the highest-value knowledge (conditional ForceNew, state transitions) isn't static — it depends on current vs. desired values. Encoding this requires a rules engine, not just a lookup table.

3. **Scope gap**: AzureRM covers ~1,134 resource types. The Bicep type index has 31,175. Transfer only helps the intersection.

4. **Version pinning**: AzureRM pins API versions per resource. AzAPI lets users choose. Knowledge may not apply across versions.

5. **Automation requirement**: Manual transfer doesn't scale. Any solution needs a generation pipeline that runs against AzureRM source periodically.

6. **Licensing**: AzureRM is MPL-2.0. Extracting derived metadata for embedding in AzAPI (also MPL-2.0) needs legal review.

---

## Recommendation

The clearest win is a **generated metadata file** per resource type containing:
- ForceNew property paths (static subset only)
- Timeout defaults
- Soft-delete flag
- Naming rules

This can be produced by an AST analysis tool scanning AzureRM source, mapping Terraform field names to ARM property paths via the expand/flatten functions. The output would sit alongside the Bicep type schemas in `internal/azure/generated/`.

Cross-field constraints and conditional ForceNew require a more expressive mechanism (possibly a declarative rules file per resource type, hand-curated for high-value resources). This is a separate effort with higher cost but substantial value for frequently-used resources like Storage Accounts and Key Vaults.
---

## Technical Implementation Methods

This section details three concrete approaches for integrating AzureRM knowledge into the AzAPI provider, based on source code analysis of `terraform-provider-azapi`.

### AzAPI Architecture Overview

The AzAPI provider's resource type system is built on embedded Bicep schemas:

- **Schema source**: `internal/azure/generated/` (334 MB, 357 service directories) synced from `ms-henglu/bicep-types-az`. These files are embedded at compile time via `//go:embed generated` in `loader.go` and **must not be modified** — they are upstream-controlled.
- **Type resolution**: `loader.go` exposes `GetResourceDefinition(resourceType, apiVersion)` which lazily deserializes a 194K-line `generated/index.json` into a `Schema` struct. All resource type lookups flow through this.
- **ForceNew scope**: `name` and `parent_id` use schema-level `RequiresReplace` plan modifiers (in `Schema`, lines 184 and 193). `location` is checked dynamically in `ModifyPlan` (line 669–672). All body properties are treated as mutable — AzAPI has **zero** per-resource-type ForceNew rules today.
- **User escape hatches**: `replace_triggers_external_values` (arbitrary dynamic values) and `replace_triggers_refs` (JMESPath expressions evaluated against the body). These require users to discover which properties need replacement themselves.
- **Timeouts**: Hardcoded for all resource types — Create 30m, Read 5m, Update 30m, Delete 30m. No per-resource-type override mechanism exists.
- **No-op change suppression**: When `IgnoreNoOpChanges` is enabled, `ModifyPlan` issues a GET to compare plan body against remote state, suppressing diffs that don't represent real changes (lines 591–632).
- **Preflight validation**: Optional (`EnablePreflight` feature flag) ARM preflight validation at plan time (lines 716–736).

#### Existing Per-Resource-Type Knowledge Patterns

AzAPI already has several mechanisms that encode per-resource-type knowledge in Go (not generated):

1. **`skipApiVersions`** in `loader.go` — A `map[string]map[string]bool` that blocks specific resource type + API version combinations known to cause `NoRegisteredProviderFound` errors (currently 2 entries: `microsoft.keyvault/vaults/keys` and `microsoft.keyvault/vaults/secrets` for two API versions each — one GA `2026-02-01` and one preview `2026-03-01-preview`).

2. **`volatileFieldList()`** in `common.go` — A hardcoded list of 22 field names (e.g., `etag`, `updatedBy`, `lastModifiedAt`) stripped from default output to prevent phantom diffs from server-side timestamps.

3. **`fix.GetWriteOnlyFix()`** in `fix/fix.go` — Hardcoded logic to fix write-only properties in the resource body (currently handles `userAssignedIdentities` normalization).

4. **`customization/` package** with `registration.go` — An `init()` registration pattern for data-plane resource customizations. Currently has exactly 2 entries:
   - `KeyVaultKeyCustomization` (key_vault_key_customization.go)
   - `FoundryAgentCustomization` (foundry_agent_customization.go)

   Each implements a `DataPlaneResource` interface. Lookup is via `GetCustomization(resourceType)`, case-insensitive.

These patterns demonstrate that AzAPI's maintainers already accept per-resource-type Go code when needed. The question is how to scale this for the knowledge categories identified in this report.

---

### Approach A: Separate JSON Metadata Files

Create a new `internal/azure/knowledge/` directory containing JSON files per resource type (or per service), produced by an AST extraction tool scanning AzureRM source.

#### Structure

```
internal/azure/knowledge/
├── index.json                          # manifest of all resource types with metadata
├── keyvault/
│   ├── microsoft.keyvault_vaults.json
│   └── microsoft.keyvault_vaults_keys.json
├── storage/
│   └── microsoft.storage_storageaccounts.json
└── ...
```

Each JSON file encodes static knowledge:

```json
{
  "resourceType": "Microsoft.Storage/storageAccounts",
  "forceNew": [
    {"path": "properties.encryption.keySource", "condition": null},
    {"path": "sku.name", "condition": "cross-zone-migration"}
  ],
  "timeouts": {"create": "60m", "read": "5m", "update": "60m", "delete": "60m"},
  "softDelete": false,
  "naming": {"regex": "^[a-z0-9]{3,24}$", "minLength": 3, "maxLength": 24},
  "sensitiveFields": ["properties.primaryAccessKey", "properties.secondaryAccessKey"],
  "diffSuppress": [
    {"path": "properties.encryption.keyVaultProperties.keyVersion", "type": "ignore"}
  ]
}
```

#### Loading mechanism

A second `//go:embed knowledge` directive in a new `knowledge_loader.go`, mirroring `loader.go`'s pattern. Lookup by resource type returns the metadata struct, consumed by `ModifyPlan` and other lifecycle methods.

#### Advantages

- Clean separation between upstream Bicep schemas (read-only) and extracted AzureRM knowledge (provider-owned)
- JSON is tooling-friendly — easier to generate, diff, and review
- Could be produced by a CI pipeline scanning AzureRM on each release

#### Disadvantages

- **Cannot express conditional logic**: The Storage Account replication type ForceNew rule depends on comparing old vs. new values at runtime — JSON can encode the property path but not the transition matrix (`LRS→ZRS` = replace, `LRS→GRS` = update)
- **Property path mapping is hard**: AzureRM's `account_replication_type` maps to `sku.name` via expand/flatten functions that aren't 1:1. Building a reliable mapping pipeline is a project in itself
- **Second sync lifecycle**: Must track AzureRM releases and re-extract. Stale metadata is worse than no metadata (wrong ForceNew → unnecessary destruction)
- **No access to Terraform plan context**: JSON metadata can't inspect `d.GetChange()` values or query Azure APIs

---

### Approach B: Go Registry Package

Create a new `internal/azure/resourceknowledge/` package with Go structs per service, following the existing `customization/` and `skipApiVersions` patterns.

#### Structure

```
internal/azure/resourceknowledge/
├── registry.go                         # type definitions + lookup
├── keyvault.go                         # Microsoft.KeyVault/* knowledge
├── storage.go                          # Microsoft.Storage/* knowledge
├── compute.go                          # Microsoft.Compute/* knowledge
└── ...
```

#### Registry definition

```go
// registry.go
package resourceknowledge

type ForceNewRule struct {
    PropertyPath string
    // Condition is optional; when non-nil, it receives old and new body values
    // and returns true if the change requires replacement.
    Condition func(oldBody, newBody map[string]interface{}) bool
}

type ResourceKnowledge struct {
    ResourceType    string
    ForceNew        []ForceNewRule
    Timeouts        *Timeouts    // nil = use defaults
    SoftDelete      bool
    NamingRegex     string       // empty = no validation
    SensitiveFields []string     // ARM JSON paths containing secrets
}

type Timeouts struct {
    Create, Read, Update, Delete time.Duration
}

var registry = map[string]*ResourceKnowledge{}

func Register(k *ResourceKnowledge) {
    registry[strings.ToLower(k.ResourceType)] = k
}

func Get(resourceType string) *ResourceKnowledge {
    return registry[strings.ToLower(resourceType)]
}
```

#### Per-service knowledge file

```go
// storage.go
package resourceknowledge

func init() {
    Register(&ResourceKnowledge{
        ResourceType: "Microsoft.Storage/storageAccounts",
        ForceNew: []ForceNewRule{
            {PropertyPath: "sku.name", Condition: func(old, new map[string]interface{}) bool {
                // Cross-zone migration requires recreation
                oldSku := strings.ToUpper(extractString(old, "sku.name"))
                newSku := strings.ToUpper(extractString(new, "sku.name"))
                zonal := map[string]bool{"ZRS": true, "GZRS": true, "RAGZRS": true}
                nonZonal := map[string]bool{"LRS": true, "GRS": true, "RAGRS": true}
                return (zonal[oldSku] && nonZonal[newSku]) || (nonZonal[oldSku] && zonal[newSku])
            }},
        },
        Timeouts: &Timeouts{
            Create: 60 * time.Minute, Read: 5 * time.Minute,
            Update: 60 * time.Minute, Delete: 60 * time.Minute,
        },
        SoftDelete:      false,
        NamingRegex:     `^[a-z0-9]{3,24}$`,
        SensitiveFields: []string{"properties.primaryAccessKey", "properties.secondaryAccessKey"},
}
```

#### Integration points in AzAPI

**1. `ModifyPlan`** (`azapi_resource.go`, after line 671 — location RequiresReplace block):

```go
// After existing location/name/parent_id RequiresReplace checks:
if state != nil && dynamic.IsFullyKnown(plan.Body) {
    if knowledge := resourceknowledge.Get(azureResourceType); knowledge != nil {
        for _, rule := range knowledge.ForceNew {
            if rule.Condition != nil {
                oldBody := extractBody(state.Body)
                newBody := extractBody(plan.Body)
                if rule.Condition(oldBody, newBody) {
                    response.RequiresReplace.Append(path.Root("body"))
                    break
                }
            } else {
                // Static ForceNew: check if the property at path changed
                if propertyChanged(state.Body, plan.Body, rule.PropertyPath) {
                    response.RequiresReplace.Append(path.Root("body"))
                    break
                }
            }
        }
    }
}
```

**2. `Schema`** (for timeout defaults): Override the hardcoded 30m/5m defaults with per-resource values from the knowledge registry, read at plan time from the `type` attribute.

**3. `ValidateConfig`**: Add naming regex validation when the `name` attribute is known and knowledge exists for the resource type.

#### Advantages

- **Follows existing patterns**: `customization/registration.go` already uses `init()` + `map` + interface lookup. `skipApiVersions` is literally a `map[string]map[string]bool`. This is the established idiom.
- **Supports conditional logic**: Go functions can express arbitrarily complex ForceNew conditions, state machine transitions, and cross-field validation — exactly what the highest-value knowledge categories require.
- **Single authoring modality**: All knowledge is Go code, reviewed and tested through existing CI.
- **No mapping pipeline needed**: Property paths are written directly against the ARM JSON body shape (e.g., `sku.name`, `properties.encryption.keySource`) because that's what AzAPI users write in their `body` attribute.

#### Disadvantages

- **Manual curation**: Doesn't scale to all 1,134 AzureRM resource types. However, the Pareto principle applies — 20–30 high-traffic resource types (Storage Accounts, Key Vaults, VMs, Virtual Networks, App Services, SQL, AKS, Cosmos DB, etc.) would cover the vast majority of user pain.
- **Maintenance burden**: Knowledge must be updated when Azure changes API behavior. But this is true for any approach — the question is whether updates are JSON file regeneration or Go code PRs.

---

### Approach C: Hybrid (Generated JSON + Go Overrides)

Combine Approaches A and B: generated JSON files for bulk static metadata (timeouts, naming rules, soft-delete flags) with Go registry overrides for complex conditional logic.

#### Structure

```
internal/azure/knowledge/
├── generated/                          # AST-extracted, CI-regenerated
│   ├── index.json
│   └── ...per-service JSON files...
├── overrides/                          # Hand-curated Go
│   ├── registry.go
│   ├── storage.go
│   └── keyvault.go
└── loader.go                           # Merges generated + overrides
```

#### Loading priority

1. Load generated JSON metadata for all resource types (bulk coverage)
2. Apply Go overrides on top — override wins for any field it specifies
3. Expose merged result via `knowledge.Get(resourceType)`

#### Advantages

- Maximum coverage: generated JSON handles the long tail of simple cases (timeouts, naming)
- Maximum expressiveness: Go overrides handle complex cases (conditional ForceNew, state machines)
- Incremental adoption: start with generated JSON, add Go overrides for high-value resources as needed

#### Disadvantages

- **Highest complexity**: Two authoring modalities, two sync lifecycles, merge semantics to define and test
- **Cognitive overhead**: Contributors must understand both systems and know when to use which
- **Overkill for initial scope**: If the practical scope is 20–30 resources (Approach B), the generated JSON layer adds complexity without proportional value

---

### Comparison Matrix

| Criterion | A: JSON Files | B: Go Registry | C: Hybrid |
|---|---|---|---|
| **Conditional ForceNew** | ❌ Cannot express | ✅ Go functions | ✅ Go overrides |
| **Bulk coverage** | ✅ Generated | ⚠️ Manual per-type | ✅ Generated + overrides |
| **Property path mapping** | ❌ Requires pipeline | ✅ Direct ARM paths | ⚠️ Partial (JSON needs mapping) |
| **Follows existing patterns** | ⚠️ New pattern | ✅ Matches `customization/`, `skipApiVersions` | ⚠️ Mixed |
| **Maintenance burden** | Medium (regen pipeline) | Low (targeted PRs) | High (two systems) |
| **Time to first value** | High (build pipeline first) | **Low** (one PR per resource) | High (build both layers) |
| **Expressiveness** | Low (static only) | High (arbitrary Go) | High |
| **Complexity** | Low | Low | High |

### Recommendation: Approach B (Go Registry Package)

**Approach B is recommended** for the following reasons:

1. **Fastest path to value**: A single PR adding `internal/azure/resourceknowledge/` with Storage Account knowledge immediately improves the experience for AzAPI's most common resource type. No pipeline, no tooling, no infrastructure.

2. **Conditional logic is the highest-value knowledge**: The analysis in this report shows that the most impactful knowledge categories — conditional ForceNew (§1, §7), soft-delete with purge protection checks (§2), state machine transitions (§7) — all require runtime evaluation. Static JSON cannot express these. Starting with the approach that handles them is more pragmatic than starting with bulk coverage of low-impact static metadata.

3. **Proven pattern**: The codebase already has `customization/registration.go` (2 entries), `skipApiVersions` (2 entries), and `volatileFieldList()` (23 entries) — all hand-curated, per-resource-type Go code. Approach B is a natural extension, not a new paradigm.

4. **Scope is manageable**: The 20–30 highest-traffic Azure resource types represent the vast majority of AzAPI usage. Manual curation of this set is a bounded, reviewable effort — unlike building and maintaining a generation pipeline for 1,134 types.

5. **No mapping problem**: AzAPI users write ARM JSON property paths directly in their `body` attribute. Knowledge rules reference the same paths. There is no snake_case → camelCase translation layer to build.

6. **Escape hatch to C**: If bulk coverage later proves necessary, Approach B's registry can serve as the override layer in a future Hybrid setup (Approach C) without rewriting the Go code.

### Implementation Plan

#### Phase 1: Foundation

| File | Action | Description |
|---|---|---|
| `internal/azure/resourceknowledge/registry.go` | Create | Type definitions (`ResourceKnowledge`, `ForceNewRule`, `Timeouts`), registry map, `Register()`, `Get()` |
| `internal/azure/resourceknowledge/helpers.go` | Create | `extractString(body, path)`, `propertyChanged(old, new, path)` utilities |
| `internal/azure/resourceknowledge/registry_test.go` | Create | Unit tests for registration and lookup |

#### Phase 2: First Resource Type

| File | Action | Description |
|---|---|---|
| `internal/azure/resourceknowledge/storage.go` | Create | `Microsoft.Storage/storageAccounts` — conditional ForceNew for SKU zone migration, 60m timeouts, naming regex, sensitive field paths |
| `internal/azure/resourceknowledge/storage_test.go` | Create | Test zone migration matrix: `LRS→ZRS` (replace), `LRS→GRS` (update), `ZRS→LRS` (replace) |

#### Phase 3: Integration

| File | Action | Description |
|---|---|---|
| `internal/services/azapi_resource.go` | Modify | `ModifyPlan`: after line ~672 (location RequiresReplace), add knowledge-based ForceNew checks |
| `internal/services/azapi_resource.go` | Modify | `ValidateConfig`: add naming regex validation when knowledge exists; warn when sensitive fields are used in `body` instead of `sensitive_body` |

#### Phase 4: Expand Coverage

Add knowledge files for high-traffic resource types in priority order:

1. `Microsoft.KeyVault/vaults` — ForceNew (name), soft-delete + purge, naming regex, sensitive fields (access policies)
2. `Microsoft.Compute/virtualMachines` — ForceNew (numerous properties), custom timeouts
3. `Microsoft.Network/virtualNetworks` — ForceNew (address space changes), diff suppression
4. `Microsoft.ContainerService/managedClusters` — ForceNew (network profile), long timeouts
5. `Microsoft.Sql/servers` — ForceNew (name, location), soft-delete, sensitive fields (admin password)
