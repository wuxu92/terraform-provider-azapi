# Goal
Add regression tests for a new generator capability: Key Vault's `ip_rules` / `virtual_network_rules` nested lists now get an empty-list schema `Default` (via `listdefault.StaticValue(types.ListValueMust(<elemType>, nil))`), driven by a new `DefaultEmptyList` flag on `typegraph.Property` set by the keyvault customizer. Branch `feature/azapin` of `terraform-provider-azapi`.

Pure in-process unit tests (they call the generator, not Azure). Do NOT run TF_ACC/live. Skip formatters/linters/full suite; the parent runs the gate.

# Where
Add to the EXISTING file `internal/native/generator/customizers/key_vault_customizer_test.go` (package `customizers_test`). Mirror the two existing tests in that file:
- `TestKeyVaultCustomizerAttachesPermissionValidators` (graph contract, uses helper `keyVaultDef(t)` which loads real defs + runs PostProcess + customizers.Apply)
- `TestKeyVaultSchemaEmitsPermissionListValidators` (emit contract, calls `generator.EmitSchema(def)` and asserts on the source string)

Reuse the existing `keyVaultDef(t)` helper (already in the file) and imports (`reflect`, `strings`, `testing`, `generator`, `typegraph`).

# What the product code does (already implemented — do NOT modify it)
- `typegraph.Property` has a new bool field `DefaultEmptyList`.
- Customizer `customizeKeyVault` sets `typegraph.FindProperty(def, path).DefaultEmptyList = true` for paths `"properties.networkAcls.ipRules"` and `"properties.networkAcls.virtualNetworkRules"`.
- The emitter emits, for such a property, a line like:
  `Default: listdefault.StaticValue(types.ListValueMust(types.ObjectType{AttrTypes: map[string]attr.Type{"value": types.StringType}}, nil)),`
  and for virtual_network_rules:
  `Default: listdefault.StaticValue(types.ListValueMust(types.ObjectType{AttrTypes: map[string]attr.Type{"id": types.StringType, "ignore_missing_vnet_service_endpoint": types.BoolType}}, nil)),`
- The emitted import block gains `github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault` and `github.com/hashicorp/terraform-plugin-framework/attr`.

# Tests to add

## 1. Graph contract — `TestKeyVaultCustomizerFlagsEmptyListDefaults`
Using `def := keyVaultDef(t)`, assert that BOTH `typegraph.FindProperty(def, "properties.networkAcls.ipRules")` and `"properties.networkAcls.virtualNetworkRules")` have `.DefaultEmptyList == true`. Also assert a control: a sibling that should NOT be flagged — e.g. `typegraph.FindProperty(def, "properties.networkAcls.bypass").DefaultEmptyList == false` (bypass is a string with a StaticString default, not an empty-list default). This guards that the flag is set on exactly the intended array properties.

## 2. Emit contract — `TestKeyVaultSchemaEmitsEmptyListDefaults`
Call `src, err := generator.EmitSchema(def)` (fatal on err). Assert:
- `src` contains the `listdefault` import path `"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"`.
- `src` contains the `attr` import path `"github.com/hashicorp/terraform-plugin-framework/attr"`.
- `src` contains exactly 2 occurrences of `listdefault.StaticValue(types.ListValueMust(` (one per list) — use `strings.Count`.
- `src` contains the ip_rules element-type literal substring: `types.ObjectType{AttrTypes: map[string]attr.Type{"value": types.StringType}}`.
- `src` contains the virtual_network_rules element-type literal substring: `map[string]attr.Type{"id": types.StringType, "ignore_missing_vnet_service_endpoint": types.BoolType}`.
  NOTE: the emitted source is gofmt-normalized; the map key order in the literal is deterministic (sorted: "id" before "ignore_missing_vnet_service_endpoint"). If an exact-substring match is brittle against gofmt spacing, assert the presence of each of the three key/type fragments (`"value": types.StringType`, `"id": types.StringType`, `"ignore_missing_vnet_service_endpoint": types.BoolType`) plus the 2x `listdefault.StaticValue(` count instead.

# Acceptance
- `go test ./internal/native/generator/customizers/ -run 'TestKeyVault' -count=1` PASSES (all four keyvault tests, including the two new ones).
- Do NOT modify product code or the two existing tests or the `keyVaultDef` helper.
- Report the exact command and its PASS output.
