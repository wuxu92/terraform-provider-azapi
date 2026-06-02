# Key Vault generation — test context

Branch `feature/azapin` in `terraform-provider-azapi`. Two brand-new behaviors need
high-signal unit tests. Do NOT run TF_ACC / live acceptance — pure in-process unit
tests only. Skip formatters/linters/full suite; the parent runs the gate.

## Behavior 1 — Key Vault permission list-element validators (case-insensitive OneOf)

Key Vault access-policy permission lists (`certificates`, `keys`, `secrets`, `storage`)
are bicep `string[]` with no enum. AzureRM constrains each to a fixed, case-insensitive
value set. Two new mechanisms implement this:

- **typegraph validator kind** `typegraph.ValidatorStringOneOfCaseInsensitive` and its
  constructor `typegraph.OneOfCaseInsensitiveValidator(message string, allowed ...string) typegraph.DescriptionValidator`
  (file `internal/native/typegraph/envelope.go`; kind enum in `walker.go`).
- **emitter branch**: a primitive array/set property (`Type.Kind == typegraph.KindArray`,
  element `Kind != KindObject`) that is settable (`!EffectiveComputed`) and carries
  `Validators` emits them as list-element validators:
  `Validators: []validator.List{ listvalidator.ValueStringsAre( stringvalidator.OneOfCaseInsensitive( "Val1", "Val2", ... ) ) }`
  plus the import `github.com/hashicorp/terraform-plugin-framework-validators/listvalidator`.
  (file `internal/native/generator/emitter.go`.)
- **customizer** `customizeKeyVault` (file `internal/native/generator/customizers/key_vault.go`)
  attaches the four value sets by ARM dot-path via `typegraph.FindProperty(def, path)` +
  `p.Validators = append(p.Validators, typegraph.OneOfCaseInsensitiveValidator(...))`.
  Registered at init in `customizers/register.go` as `Register(armtypes.KeyVault, customizeKeyVault)`.

Exact ARM paths and allowed value sets (from `customizers/key_vault.go`, MUST match):
- `properties.accessPolicies.permissions.certificates`:
  Backup, Create, Delete, DeleteIssuers, Get, GetIssuers, Import, List, ListIssuers,
  ManageContacts, ManageIssuers, Purge, Recover, Restore, SetIssuers, Update
- `properties.accessPolicies.permissions.keys`:
  Backup, Create, Decrypt, Delete, Encrypt, Get, Import, List, Purge, Recover, Restore,
  Sign, UnwrapKey, Update, Verify, WrapKey, Release, Rotate, GetRotationPolicy, SetRotationPolicy
- `properties.accessPolicies.permissions.secrets`:
  Backup, Delete, Get, List, Purge, Recover, Restore, Set
- `properties.accessPolicies.permissions.storage`:
  Backup, Delete, DeleteSAS, Get, GetSAS, List, ListSAS, Purge, Recover, RegenerateKey,
  Restore, Set, SetSAS, Update

Distinctive per-list values (good for no-leak assertions): `DeleteIssuers`/`ManageContacts`
appear ONLY in certificates; `WrapKey`/`UnwrapKey` ONLY in keys; `GetSAS`/`RegenerateKey`
ONLY in storage. secrets is the smallest set.

### Test package + def loading

`generator` does NOT import `customizers` (would cycle), so this test MUST live in
`package customizers_test` (external), e.g. new file
`internal/native/generator/customizers/key_vault_customizer_test.go`. It can import
`generator`, `customizers`, `typegraph`, and `azure`.

Real Key Vault defs load exactly like `internal/native/generator/fixtures_test.go` does
(that helper is unexported in `package generator`, so replicate its body here):

```go
version, err := azure.GetLatestStableApiVersion("Microsoft.KeyVault/vaults")   // skip on err
location, err := azure.GetResourceTypeLocation("Microsoft.KeyVault/vaults", version) // skip on err
data, err := azure.StaticFiles.ReadFile("generated/" + location)               // skip on err
defs, err := typegraph.ParseTypesJSON(data)                                    // fatal on err
```
Imports: `"github.com/Azure/terraform-provider-azapi/internal/azure"`,
`".../internal/native/generator"`, `".../internal/native/generator/customizers"`,
`".../internal/native/typegraph"`.

Then: `typegraph.PostProcess(defs)`; `customizers.Apply(defs)`; find the KeyVault def
(its `.Name` starts with `"Microsoft.KeyVault/vaults@"`, or use `typegraph.ARMTypeOf(def) == "Microsoft.KeyVault/vaults"`).

`typegraph.FindProperty(def, path) *typegraph.Property` — `.Validators` is
`[]typegraph.DescriptionValidator`; each has `.Kind` (compare to
`typegraph.ValidatorStringOneOfCaseInsensitive`) and `.Allowed []string`.

`generator.EmitSchema(def) (string, error)` returns the generated Go source.

### What to assert (two focused tests)

**TestKeyVaultCustomizerAttachesPermissionValidators** (customizer contract, on the graph
after `customizers.Apply`): for each of the four paths, the property carries exactly one
`DescriptionValidator` whose `Kind == ValidatorStringOneOfCaseInsensitive` and whose
`Allowed` equals the expected set above (order matters — the customizer preserves it).

**TestKeyVaultSchemaEmitsPermissionListValidators** (emit contract, on `EmitSchema` output):
- source contains the listvalidator import path.
- `strings.Count(source, "stringvalidator.OneOfCaseInsensitive(")` == 4 (one per list).
- source contains `listvalidator.ValueStringsAre(`.
- distinctive values present: `"DeleteIssuers"`, `"WrapKey"`, `"GetSAS"`.
- no-leak sanity: the secrets set is a subset — assert e.g. `"DeleteIssuers"` and
  `"WrapKey"` and `"GetSAS"` each appear exactly the number of times they occur across
  the four sets (DeleteIssuers=1, WrapKey=1, GetSAS=1) so a value from one list did not
  leak into another. (Use `strings.Count`.)

## Behavior 2 — shared azapi_client_config data-source instance (`config.ClientConfig`)

The acceptance-test config scaffolding moved to a dedicated package
`internal/native/services/config` (`package config`, file `config.go`). It exposes a
single shared data-source instance every scenario references instead of hardcoding the
`data "azapi_client_config" "current" {}` block:

```go
var ClientConfig = ClientConfigData{DataSourceConfigBase: DataSourceConfigBase{tfType: "azapi_client_config", label: "current"}}
```
`ClientConfigData` embeds `DataSourceConfigBase`, which provides:
- `Config() string`  → `data "azapi_client_config" "current" {}` (via `fmt.Sprintf("data %q %q {}", tfType, label)`)
- `RefOf(path string) string` → `data.azapi_client_config.current.<path>`
- `IDRef() string` → `data.azapi_client_config.current.id`
- `DataSourceType() string` → `azapi_client_config`; `Label() string` → `current`

Key Vault's config builder (`internal/native/services/keyvault/key_vault_config.go`) now
threads this instance: it prepends `config.ClientConfig.Config()` and references
`config.ClientConfig.RefOf("tenant_id")` / `RefOf("object_id")` instead of hardcoded
strings. The live acceptance suite already passes; this unit test guards the shared
instance's rendered contract that Key Vault depends on.

### Test package + what to assert

Add to `package config_test` (external) — the existing test file is
`internal/native/services/config/config_test.go` (imports
`".../internal/native/services/config"`). A sibling file
`internal/native/services/config/client_config_test.go` in `package config_test` is fine.

**TestClientConfigSharedInstance**:
- `config.ClientConfig.Config()` == `data "azapi_client_config" "current" {}`
- `config.ClientConfig.RefOf("tenant_id")` == `data.azapi_client_config.current.tenant_id`
- `config.ClientConfig.RefOf("object_id")` == `data.azapi_client_config.current.object_id`
- `config.ClientConfig.IDRef()` == `data.azapi_client_config.current.id`
- `config.ClientConfig.DataSourceType()` == `azapi_client_config`
- `config.ClientConfig.Label()` == `current`

Assert exact string equality (these are load-bearing HCL fragments other resources embed).
