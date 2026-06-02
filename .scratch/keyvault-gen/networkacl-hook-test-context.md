# Goal
Add a focused white-box unit test for a new Key Vault runtime hook, `stripEmptyNetworkACLRuleLists`, in the AzAPI native provider on branch `feature/azapin` of `terraform-provider-azapi`.

Do NOT run TF_ACC / live acceptance — pure in-process unit test. Skip formatters/linters/full suite; the parent runs the gate.

# Where
- File under test: `internal/native/services/keyvault/key_vault_hooks.go`
- Add the test to the EXISTING file: `internal/native/services/keyvault/key_vault_hooks_test.go` (package `keyvault`). Follow the style of the existing `TestEnsureAccessPolicies` in that file — table/subtest white-box calls against a `nativeresource.CrudCtx` literal, asserting on the mutated map. Import `nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"`.

# The hook under test (verbatim contract)
```go
// stripEmptyNetworkACLRuleLists deletes an empty ipRules or virtualNetworkRules
// array from properties.networkAcls in the ARM response. Both are Optional+Computed
// lists that plan known-null when the user configures network_acls with only
// default_action/bypass, but the Key Vault RP echoes them back as [] — flattening
// that empty array over the planned-null base yields an "inconsistent result after
// apply" on create/update and perpetual drift on read. Deleting the empty key lets
// the mapper preserve the base value (planned/state) instead, so null stays null and
// a user-supplied [] stays []. A non-empty list is left untouched and flattens normally.
func stripEmptyNetworkACLRuleLists(c *nativeresource.CrudCtx) {
	props, ok := c.Response["properties"].(map[string]interface{})
	if !ok {
		return
	}
	acls, ok := props["networkAcls"].(map[string]interface{})
	if !ok {
		return
	}
	for _, key := range []string{"ipRules", "virtualNetworkRules"} {
		if arr, ok := acls[key].([]interface{}); ok && len(arr) == 0 {
			delete(acls, key)
		}
	}
}
```
IMPORTANT: the hook reads/mutates `c.Response` (NOT `c.Body` — that is a different hook). Response is the ARM GET response map.

# Cases to cover (name the test `TestStripEmptyNetworkACLRuleLists`, subtests)
1. **empty ipRules and empty virtualNetworkRules are both deleted** — Response `properties.networkAcls` = `{"defaultAction":"Deny","bypass":"AzureServices","ipRules":[],"virtualNetworkRules":[]}`. After the hook, both keys are ABSENT from the networkAcls map; `defaultAction`/`bypass` remain untouched.
2. **non-empty ipRules is preserved** — `ipRules` = `[{"value":"1.2.3.4/32"}]`, `virtualNetworkRules` = `[]`. After: `ipRules` still present with its one element intact; `virtualNetworkRules` deleted.
3. **non-empty virtualNetworkRules is preserved** — symmetric to case 2: `virtualNetworkRules` = `[{"id":"/subscriptions/.../subnets/s1"}]`, `ipRules` = `[]`. After: `virtualNetworkRules` present intact; `ipRules` deleted.
4. **missing networkAcls is a no-op** — Response `properties` = `{"tenantId":"x"}` (no networkAcls). Hook must not panic and must not add a networkAcls key.
5. **missing properties is a no-op** — Response = `{}` (empty map). Hook must not panic; Response stays empty (no "properties" key added).
6. (optional but nice) **absent ipRules/virtualNetworkRules keys are a no-op** — networkAcls present with only `{"defaultAction":"Allow"}`; hook leaves it exactly as-is.

# Acceptance
- New test `TestStripEmptyNetworkACLRuleLists` compiles and passes: `go test ./internal/native/services/keyvault/ -run 'TestStripEmptyNetworkACLRuleLists' -count=1`.
- Do NOT modify product code or the existing `TestEnsureAccessPolicies`.
- Report the final `go test` line you ran and its PASS output.
