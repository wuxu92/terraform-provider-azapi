# Tester context — Key Vault purge-on-destroy decision (`purgePlan`, `subscriptionOf`)

## Where
- Package `keyvault` (white-box, `package keyvault`), file to CREATE:
  `internal/native/services/keyvault/key_vault_purge_test.go`
- DO NOT touch `key_vault_hooks_test.go` (holds `TestEnsureAccessPolicies`) — add a new file.
- Follow the style of `internal/native/services/keyvault/key_vault_hooks_test.go`: table/subtest driven, `t.Run`, plain stdlib `testing`, no testify.

## System under test (all in `key_vault_hooks.go`, same package — call directly)

### `purgePlan(state types.Object, id parse.ResourceId) (action purgeAction, deletedVaultID string)`
PURE, no I/O. This is the real target — it encodes the entire purge decision matrix. `purgeOnDestroy` (the live wrapper) only turns the result into diagnostics + one HTTP POST, so it is NOT unit-tested here.

Decision order (first match wins):
1. `purge_on_destroy` state attr is missing / null / false  → `purgeSkip`, `""`
2. else if `properties.enable_purge_protection` is true      → `purgeBlocked`, `""`
3. else if normalized `location` is `""`                     → `purgeMissingLocation`, `""`
4. else if the ID's `AzureResourceId` has no subscription    → `purgeMissingSubscription`, `""`
5. else                                                       → `purgeDo`, and `deletedVaultID` =
   `/subscriptions/<sub>/providers/Microsoft.KeyVault/locations/<normLoc>/deletedVaults/<id.Name>`

`purgeAction` constants (unexported, same package): `purgeSkip`, `purgeBlocked`, `purgeMissingLocation`, `purgeMissingSubscription`, `purgeDo`.

Location is normalized via `location.Normalize` = lowercase + strip spaces (`"East US"` → `"eastus"`). Assert the `deletedVaultID` carries the NORMALIZED location.

Subscription is parsed by `subscriptionOf` (below) — case-insensitive `subscriptions` segment.

### `subscriptionOf(armResourceID string) string`
Splits the ARM ID on `/` and returns the segment after a case-insensitive `subscriptions` segment; `""` when absent. Test it directly too:
- `/subscriptions/abc-123/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/v` → `"abc-123"`
- case-insensitive: `/Subscriptions/abc-123/...` → `"abc-123"`
- no subscription segment (e.g. `/providers/Microsoft.Management/managementGroups/g`) → `""`
- empty string → `""`
- trailing `subscriptions` with nothing after it → `""` (guard: `i+1 < len`)

## Building the `types.Object` state (framework value construction)
Imports:
```go
"github.com/hashicorp/terraform-plugin-framework/attr"
"github.com/hashicorp/terraform-plugin-framework/types"
"github.com/Azure/terraform-provider-azapi/internal/services/parse"
```

Helper to build a vault state object — spell only the attrs a case needs; `types.ObjectValueMust` requires attrTypes and attrs maps to agree exactly. Suggested helper:

```go
// kvState builds a minimal vault state object. Pass properties=nil to omit the nested
// properties object (childObject then yields a null object).
func kvState(purgeOnDestroy attr.Value, location attr.Value, properties attr.Value) types.Object {
	attrTypes := map[string]attr.Type{}
	attrs := map[string]attr.Value{}
	if purgeOnDestroy != nil {
		attrTypes["purge_on_destroy"] = types.BoolType
		attrs["purge_on_destroy"] = purgeOnDestroy
	}
	if location != nil {
		attrTypes["location"] = types.StringType
		attrs["location"] = location
	}
	if properties != nil {
		attrTypes["properties"] = properties.Type(context.Background())
		attrs["properties"] = properties
	}
	return types.ObjectValueMust(attrTypes, attrs)
}

// props builds a properties object carrying enable_purge_protection. Pass a
// types.BoolNull()/types.BoolValue(...) as needed, or use types.ObjectNull for absent.
func props(enablePurgeProtection attr.Value) types.Object {
	return types.ObjectValueMust(
		map[string]attr.Type{"enable_purge_protection": types.BoolType},
		map[string]attr.Value{"enable_purge_protection": enablePurgeProtection},
	)
}
```
- `properties.Type(context.Background())` needs `"context"` import; alternatively hardcode `types.ObjectType{AttrTypes: map[string]attr.Type{"enable_purge_protection": types.BoolType}}` for the properties attr type.
- A NULL properties object: `types.ObjectNull(map[string]attr.Type{"enable_purge_protection": types.BoolType})` — pass this to exercise the "properties present but null" path (should behave like absent → not blocked).
- `id` is a plain struct: `parse.ResourceId{AzureResourceId: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/myvault", Name: "myvault", ApiVersion: "2026-02-01"}`.

## Required cases for `purgePlan` (assert BOTH returned values)
1. **skip — purge_on_destroy absent**: `kvState(nil, strVal("eastus"), props(boolTrue))` → `purgeSkip`, `""`. (Even with protection on, skip wins because purge_on_destroy is unset — order matters.)
2. **skip — purge_on_destroy null**: `purge_on_destroy = types.BoolNull()` → `purgeSkip`.
3. **skip — purge_on_destroy false**: `types.BoolValue(false)` → `purgeSkip`.
4. **blocked — protection on**: purge true, `props(types.BoolValue(true))`, location set, sub present → `purgeBlocked`, `""`.
5. **not blocked — protection false**: purge true, `props(types.BoolValue(false))`, location "eastus", sub present → `purgeDo` (proves false protection does NOT block).
6. **not blocked — protection null**: `props(types.BoolNull())` → proceeds past blocked (→ purgeDo given location+sub).
7. **not blocked — properties object null/absent**: `properties=nil` and `properties=types.ObjectNull(...)` both → proceed past blocked.
8. **missing location — empty string**: purge true, protection off, `location=types.StringValue("")` → `purgeMissingLocation`, `""`.
9. **missing location — null**: `location=types.StringNull()` → `purgeMissingLocation`.
10. **missing location — whitespace-only normalizes to empty**: `location=types.StringValue("   ")` → Normalize strips spaces → `""` → `purgeMissingLocation`. (Good edge: proves normalization runs before the empty check.)
11. **missing subscription**: purge true, protection off, location set, `id.AzureResourceId="/providers/Microsoft.KeyVault/vaults/v"` (no subscription) → `purgeMissingSubscription`, `""`.
12. **purgeDo happy path + normalized location + ID shape**: purge true, protection off, `location=types.StringValue("East US")`, `id={AzureResourceId:"/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/myvault", Name:"myvault"}` →
    `purgeDo`, `deletedVaultID == "/subscriptions/sub-123/providers/Microsoft.KeyVault/locations/eastus/deletedVaults/myvault"`.
    Assert the exact string: proves subscription extraction, location normalization, and `id.Name` all compose correctly.

## Notes / invariants to defend
- Decision ORDER is a real contract: skip beats blocked beats missing-location beats missing-subscription. Case 1 (unset purge + protection on returning skip, not blocked) locks that ordering.
- `deletedVaultID` is `""` for every non-`purgeDo` action — assert it, so a future refactor can't leak a half-built ID.
- Location normalization must happen BEFORE the empty check (case 10) and must reach the emitted ID (case 12).
- Do NOT write tests that merely restate the code (e.g. asserting the constant values). Assert observable decisions from realistic inputs.
- Keep it stdlib `testing`; no live Azure, no `TF_ACC`, no network.

## Verify
`go test ./internal/native/services/keyvault/... -run 'TestPurgePlan|TestSubscriptionOf' -count=1` must pass. Also ensure `gofmt -l` clean on the new file. Do NOT run the live acceptance suite.
