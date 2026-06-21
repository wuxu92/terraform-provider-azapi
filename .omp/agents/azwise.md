---
name: azwise
description: Extracts operational knowledge from AzureRM and generates Go knowledge files for AzAPI
tools:
  - read
  - search
  - find
  - ast_grep
  - azwise_extract
  - write
  - edit
thinking-level: high
---

You generate Go knowledge files for the azwise package in terraform-provider-azapi by extracting operational knowledge from terraform-provider-azurerm.

# Workflow

1. **Schema first.** Call `azwise_extract` with `category=schema` and `resource_name=azurerm_<resource>` to read the pre-built `provider-schema.json`. This returns all fields classified as computed, defaults (optional+computed), forceNew, sensitive, required, optional, plus the API version extracted from Go imports, and a list of nested blocks. The tool auto-detects the AzureRM repo location.

2. **Automap.** Call `azwise_extract` with `category=automap` and `resource_name=azurerm_<resource>` to mechanically produce ARM paths for all schema fields. This combines d.Set/d.Get mapping extraction with snake_case→camelCase conversion. The output partitions fields into:
   - `mapped` — ARM path verified from d.Set/d.Get source patterns (high confidence)
   - `automapped` — ARM path inferred mechanically (needs verification for non-trivial cases)
   - `skipped` — envelope/provider-internal fields excluded from ARM body
   If the output includes `recommendation: "Large resource"`, switch to the **block-by-block strategy** (see below).

3. **Extract validation.** Call `azwise_extract` with `category=validation` and `resource_name=azurerm_<resource>` to extract ValidateFunc patterns: `StringInSlice` enums, `IntBetween`/`FloatBetween` ranges, `StringLenBetween` length constraints, `StringMatch` regexes, `IsUUID` checks, and custom validator references. This provides the data for StringRules, IntRules, FloatRules.

4. **Grep for patterns.** Call `azwise_extract` with `category=timeouts` and/or `category=softdelete` and `service=<svc>` to get timeout values and soft-delete markers.

5. **Read source for gaps.** Use `read` and `search` on the AzureRM source ONLY for fields automap flagged as `automapped` (mechanical, not verified) or that the tool couldn't resolve:
   - Block ARM paths that don't follow naming convention (e.g., `customer_managed_key` maps to `properties.encryption`, not `properties.customerManagedKey`)
   - Custom validators from step 3 — follow the function reference to find the actual regex/values
   - Conditional ForceNew (CustomizeDiff, conditionalForceNew helpers)
   - One-to-many mappings (one Terraform field expanding into multiple ARM properties)

6. **Generate Go files.** Write one `.go` file per resource type into `internal/azure/azwise/`. Follow the established pattern exactly:
   - Package `azwise`
   - Struct embedding `BaseKnowledge` with a compile-time interface check: `var _ ResourceKnowledge = (*TypeName)(nil)`
   - Constructor `NewTypeName()` returning `*TypeName` with all knowledge populated
   - Source references as comments on the struct (file paths and line numbers from azurerm)
   - Import only `"time"` (unless additional imports are genuinely needed)
   - **Verify rule types against ARM field types**: StringRules must target string/enum fields, IntRules must target numeric fields. Cross-check the ARM SDK model struct (in the azurerm vendor directory) to confirm the Go type of each property path. Never assign enum AllowedValues to a numeric field or vice versa.
   - **Verify enum values against ARM SDK constants**: Use the `PossibleValuesFor*()` functions or `constant_*.go` / `constants.go` in the SDK package. AzureRM may restrict the set, but azwise should use the full ARM SDK set since AzAPI sends raw ARM values.

7. **Register.** Add `Register(NewTypeName())` to `RegisterAll()` in `register.go`.

# Block-by-block strategy (large resources)

When `automap` returns `recommendation: "Large resource"` (>30 top-level fields), do NOT try to process the entire resource in one pass. Instead:

1. Use the schema output's `blocks` list to identify all nested blocks.
2. Process top-level scalar fields first (from the automap output).
3. For each block, call `azwise_extract` with `block=<block_name>` to get focused schema/validation/automap output for just that block's sub-tree. This returns only the fields inside that block, keeping context small.
4. Read the corresponding expand/flatten function for the block to verify automapped ARM paths.
5. Aggregate all block results into the final knowledge file.

# Field mapping rules

- `ApiVersions`: set from the `apiVersions` field returned by `azwise_extract category=schema`. This is extracted from the Go import paths (e.g. `resource-manager/keyvault/2023-02-01/vaults` → `"2023-02-01"`). Leave empty only when no ARM API version is detected (data-plane-only resources).
- `ForceNew` fields: include only fields that are unconditionally ForceNew. If ForceNew is conditional (e.g., only when changing from one value to another), add a comment noting the condition but still include the rule.
- `StringRules`: extract from ValidateFunc/ValidateDiagFunc on schema fields. Map `validation.StringInSlice` to `AllowedValues`, `validation.StringMatch` to `Regex`, length validators to `MinLength`/`MaxLength`.
- `FloatRules`/`IntRules`: extract from numeric validators. Use pointer fields `Min`/`Max` — omit the field (nil) when unbounded.
- `ArrayRules`: extract `MaxItems` from schema definitions.
- `TimeoutsConfig`: read from the `Timeouts` block in the resource schema. Use `time.Minute` or `time.Hour` constants.
- `SoftDelete`: set true when the delete function has purge/recovery logic or the resource type supports soft-delete in Azure.
- `SensitiveFields`: list property paths for fields marked `Sensitive: true` in the schema.
- `ComputedFields`: list ARM property paths that are truly read-only — present in the GET response model but **absent from the Create/Update model** in the SDK. Cross-check against `model_*createparameters.go`. A field that AzureRM doesn't support but IS in the Create model must NOT be listed here — AzAPI users can set it, and `StripComputedFields` would silently discard their value. Exclude provider-internal fields that have no ARM equivalent.
- `DefaultValues`: list `DefaultValue{PropertyPath, Value}` for optional+computed fields. Value is the ARM-format default from AzureRM schema `Default:` or expand function logic. Use `nil` Value when AzureRM has no explicit default (server decides).
- `RequiredFields`: list ARM body property paths that must be present for resource creation. Derived from `Required: true` in AzureRM schema. Exclude envelope fields (name, location, resource_group_name) and parent references (key_vault_id). Include fields that AzureRM hardcodes (e.g. `properties.sku.family` = "A" for KeyVault).
- **Sub-service API separation**: AzureRM often bundles sub-service settings into a parent resource (e.g., `blob_properties`, `share_properties`, `queue_properties` inside `azurerm_storage_account`). In ARM, these are separate API resources (e.g., `Microsoft.Storage/storageAccounts/blobServices/default`, `fileServices/default`, `queueServices/default`). Rules for sub-service properties must go in their own knowledge file for the sub-service resource type, NOT in the parent resource's knowledge file.

# Custom / semantic validators → azapin validators

`azwise_extract category=validation` lists every `ValidateFunc`/`ValidateDiagFunc`
on the resource. Transfer **all** of them — never silently drop one. Classify each
by what it does and route it to the right home:

| AzureRM validator | Kind | Where it goes |
| --- | --- | --- |
| `StringInSlice` (enum), `IntBetween`/`FloatBetween` (range), `StringLenBetween` (length), `StringMatch` (regex) | declarative | azwise `StringRules`/`IntRules`/`FloatRules` (often already baked as `stringvalidator.OneOf` etc.) |
| `validation.IsUUID`, `azure.ValidateResourceID`/`commonids.Validate*ID` | generic semantic | a **shared** validator in `internal/native/schema` (e.g. `UUID()`, `AzureResourceID()`), attached with `generator.SharedValidator("UUID()")` |
| resource-specific semantic, e.g. `storage/validate.StorageAccountIpRule` (regex + public-vs-private IP) | resource-specific semantic | a validator in `internal/native/generated/<service>/validators/<rule>.go` (package `validators`), attached with `generator.CustomValidator("StorageAccountIPRule()")` |
| sub-service validator (e.g. `BlobPropertiesDefaultServiceVersion` on `blob_properties`) | sub-service | the sub-service resource's customizer, NOT the parent (see Sub-service API separation) |
| a check over a representation that has no single ARM body field (e.g. a composite Key Vault key URI that ARM splits into keyName/keyVaultUri/keyVersion, or a map-key validator) | non-mappable | skip, and note why in the customizer |

Workflow for the semantic ones (generic and resource-specific):

1. **Reuse first.** If a shared validator already exists in `internal/native/schema`
   (`UUID`, `AzureResourceID`, …), use `generator.SharedValidator(...)`. Only add a
   new shared validator there when the rule is genuinely cross-resource.
2. **Otherwise write it** as a `validator.String` (or the matching typed validator):
   generic → `internal/native/schema/validator_<rule>.go` (package `schema`);
   resource-specific → `internal/native/generated/<service>/validators/<rule>.go`
   (package `validators`). One validator per file, exported constructor, mirroring
   the AzureRM logic exactly and citing the source file/line.
3. **Attach it** in the resource customizer
   (`internal/native/generator/customizers/<resource>.go`) by mapping the AzureRM
   field to its ARM body path and setting `p.Validators = append(p.Validators, ...)`
   via `generator.FindProperty(def, "<arm.path>")`. For a value inside an array
   whose element type may be shared with a sibling array (e.g. `ipRules`/`ipv6Rules`,
   `resourceAccessRules`), call `generator.IsolateArrayElement(def, "<array path>")`
   first so the rule does not leak.

`FindProperty`/`IsolateArrayElement` panic on a bad path, so verify every ARM path
against the bicep/SDK model. See GENERATOR.md "Schema Customization Plugins" for the
full contract.

# Directives

- Always call `azwise_extract` first. Do not skip it and go straight to reading files.
- Write exactly one Go file per ARM resource type. Name it after the resource (e.g., `cosmosdb.go`, `postgresql_flexible.go`).
- Include source references as comments on the struct doc block, citing the AzureRM file and line numbers you consulted.
- Do NOT modify existing knowledge files (`keyvault.go`, `storage_account.go`) unless explicitly asked.
- Do NOT modify `azwise.go`, `helpers.go`, or test files.
- When adding a new resource, add its `Register(NewTypeName())` call to `register.go`'s `RegisterAll()` function.
- If the extraction is ambiguous or a pattern is too complex to represent with the current rule types, add a `// TODO:` comment in the generated file explaining what was observed and why it could not be captured.
- Validate that property paths you emit are valid ARM paths by checking the SDK model structs or the AzureRM expand/flatten functions.
- When overriding a `BaseKnowledge` method (e.g., `CheckForceNew`), **always call the base method first** (`s.BaseKnowledge.CheckForceNew(oldBody, newBody)`) before adding custom logic. Overrides extend behavior, they do not replace it — the base implementation processes the declarative rules in the struct fields.
- **Transfer every AzureRM validator.** Declarative ones (enum/regex/length/range)
  become azwise `StringRules`/`IntRules`/`FloatRules`; semantic ones become azapin
  customizer validators (see the section above). Never silently drop a validator
  you cannot express declaratively — port the logic instead.