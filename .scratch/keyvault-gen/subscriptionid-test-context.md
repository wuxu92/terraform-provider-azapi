# Tester context — contract test for `parse.ResourceId.SubscriptionId()`

## Where
- Package `parse`, file: `internal/services/parse/resource.go` gained a new method.
- Existing tests live in `internal/services/parse/resource_test.go` (package `parse`, white-box, plain stdlib `testing` — see `TestResourceID`-style table tests there). ADD a new test function to that file (or a new `_test.go` in the same package if you prefer; matching the existing file is fine). Do NOT modify unrelated tests.

## System under test
```go
// SubscriptionId returns the subscription GUID the resource lives under, or "" when the
// ID is not subscription-scoped (e.g. tenant/management-group scope) or cannot be
// parsed. It reads AzureResourceId via the standard arm parser rather than string
// slicing, so it stays correct across ID casing and scope shapes.
func (id ResourceId) SubscriptionId() string {
	armId, err := arm.ParseResourceID(id.AzureResourceId)
	if err != nil {
		return ""
	}
	return armId.SubscriptionID
}
```
It is a thin wrapper over `github.com/Azure/azure-sdk-for-go/sdk/azcore/arm`.`ParseResourceID`. Construct `parse.ResourceId` values directly (exported struct); only `AzureResourceId` matters to this method.

## Verified upstream behavior (probed live, use as ground truth)
- `/subscriptions/cb563ee9-7df0-468e-81d5-166968d1f89a/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/myvault` → `"cb563ee9-7df0-468e-81d5-166968d1f89a"`
- `/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/myvault` → `"sub-123"` (non-GUID accepted)
- `/Subscriptions/sub-123/resourceGroups/rg/...` (capital S) → `"sub-123"` (case-insensitive)
- `/providers/Microsoft.KeyVault/vaults/v` (no subscription) → `""`
- `""` → parse error → `""`
- `/subscriptions/` → parse error → `""`

## Required test: `TestResourceId_SubscriptionId`
Table/subtest driven, mirroring the existing table style in `resource_test.go`. Cases (assert exact returned string):
1. standard subscription-scoped ID → the GUID
2. non-GUID subscription token → that token (proves it doesn't GUID-validate)
3. capital-S `Subscriptions` segment → the token (case-insensitive)
4. subscription-less scope (`/providers/Microsoft.KeyVault/vaults/v`) → `""`
5. empty `AzureResourceId` → `""`
6. malformed `/subscriptions/` (trailing, nothing after) → `""`
7. a resource-group-scoped ID (e.g. `/subscriptions/<guid>/resourceGroups/rg`) → the GUID (sanity that RG scope still yields the subscription)

## Contract defended
The `""`-on-unparseable/absent branch is the load-bearing one: `purgePlan` (keyvault) relies on `SubscriptionId()` returning `""` (not a panic, not a partial) to route to `purgeMissingSubscription`. The GUID happy path must return the exact subscription so the purge endpoint targets the right subscription.

## Constraints
- Pure stdlib `testing`; no testify, no network, no TF_ACC. No production-code edits.
- Reuse the exported `parse.ResourceId` struct; no new helpers needed.

## Verify
`go test ./internal/services/parse/... -run TestResourceId_SubscriptionId -count=1` passes; `gofmt -l` clean on the edited file. Report the passing output tail.
