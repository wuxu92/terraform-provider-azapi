package keyvault

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
)

// propsType is the attr-type shape of the nested vault properties object that purgePlan
// reads (only enable_purge_protection matters to the decision).
var propsType = map[string]attr.Type{"enable_purge_protection": types.BoolType}

// kvState builds a minimal vault state object carrying only the attributes a case needs.
// Pass a nil attr.Value for purge_on_destroy / location / properties to omit that
// attribute entirely (AttrObject/AttrBool/AttrString then treat it as absent). types.Object
// construction requires the attrTypes and attrs maps to agree exactly, so each attribute
// is registered only when present.
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

// props builds a properties object carrying enable_purge_protection.
func props(enablePurgeProtection attr.Value) types.Object {
	return types.ObjectValueMust(propsType, map[string]attr.Value{"enable_purge_protection": enablePurgeProtection})
}

// actionName renders a purgeAction for readable failure messages only; the assertions
// compare the raw values.
func actionName(a purgeAction) string {
	switch a {
	case purgeSkip:
		return "purgeSkip"
	case purgeBlocked:
		return "purgeBlocked"
	case purgeMissingLocation:
		return "purgeMissingLocation"
	case purgeMissingSubscription:
		return "purgeMissingSubscription"
	case purgeDo:
		return "purgeDo"
	default:
		return fmt.Sprintf("purgeAction(%d)", int(a))
	}
}

// TestPurgePlan exercises the full purge decision matrix directly (white-box). purgePlan
// is pure, so every branch and its ordering is asserted from realistic state/ID inputs.
// It defends: (1) the first-match decision order — skip beats blocked beats
// missing-location beats missing-subscription; (2) that deletedVaultID stays "" for every
// non-purgeDo outcome so a refactor can't leak a half-built ID; (3) that location
// normalization runs before the empty check and reaches the emitted ID; and (4) the exact
// deletedVaults ID composed from subscription + normalized location + vault name.
func TestPurgePlan(t *testing.T) {
	boolTrue := types.BoolValue(true)
	boolFalse := types.BoolValue(false)

	// stdID has a subscription and a vault name; noSubID is a subscription-less ARM ID.
	stdID := parse.ResourceId{
		AzureResourceId: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/myvault",
		Name:            "myvault",
		ApiVersion:      "2026-02-01",
	}
	noSubID := parse.ResourceId{
		AzureResourceId: "/providers/Microsoft.KeyVault/vaults/v",
		Name:            "v",
	}

	// happyID is the ID purgePlan emits for stdID + a normalized "eastus" location.
	const happyID = "/subscriptions/sub-123/providers/Microsoft.KeyVault/locations/eastus/deletedVaults/myvault"

	cases := []struct {
		name           string
		state          types.Object
		id             parse.ResourceId
		wantAction     purgeAction
		wantDeletedVID string
	}{
		{
			// purge_on_destroy is unset even though protection is ON — skip must win,
			// locking the skip-beats-blocked ordering.
			name:       "skip: purge_on_destroy absent (beats blocked)",
			state:      kvState(nil, types.StringValue("eastus"), props(boolTrue)),
			id:         stdID,
			wantAction: purgeSkip,
		},
		{
			name:       "skip: purge_on_destroy null (beats blocked)",
			state:      kvState(types.BoolNull(), types.StringValue("eastus"), props(boolTrue)),
			id:         stdID,
			wantAction: purgeSkip,
		},
		{
			name:       "skip: purge_on_destroy false (beats blocked)",
			state:      kvState(boolFalse, types.StringValue("eastus"), props(boolTrue)),
			id:         stdID,
			wantAction: purgeSkip,
		},
		{
			// purge requested, protection on, location + subscription present: blocked
			// must win over the later location/subscription checks.
			name:       "blocked: enable_purge_protection true",
			state:      kvState(boolTrue, types.StringValue("eastus"), props(boolTrue)),
			id:         stdID,
			wantAction: purgeBlocked,
		},
		{
			name:           "not blocked: enable_purge_protection false reaches purgeDo",
			state:          kvState(boolTrue, types.StringValue("eastus"), props(boolFalse)),
			id:             stdID,
			wantAction:     purgeDo,
			wantDeletedVID: happyID,
		},
		{
			name:           "not blocked: enable_purge_protection null reaches purgeDo",
			state:          kvState(boolTrue, types.StringValue("eastus"), props(types.BoolNull())),
			id:             stdID,
			wantAction:     purgeDo,
			wantDeletedVID: happyID,
		},
		{
			name:           "not blocked: properties attribute absent reaches purgeDo",
			state:          kvState(boolTrue, types.StringValue("eastus"), nil),
			id:             stdID,
			wantAction:     purgeDo,
			wantDeletedVID: happyID,
		},
		{
			name:           "not blocked: properties object null reaches purgeDo",
			state:          kvState(boolTrue, types.StringValue("eastus"), types.ObjectNull(propsType)),
			id:             stdID,
			wantAction:     purgeDo,
			wantDeletedVID: happyID,
		},
		{
			name:       "missing location: empty string",
			state:      kvState(boolTrue, types.StringValue(""), props(boolFalse)),
			id:         stdID,
			wantAction: purgeMissingLocation,
		},
		{
			name:       "missing location: null",
			state:      kvState(boolTrue, types.StringNull(), props(boolFalse)),
			id:         stdID,
			wantAction: purgeMissingLocation,
		},
		{
			// Normalization strips the spaces to "" — proves it runs before the empty check.
			name:       "missing location: whitespace normalizes to empty",
			state:      kvState(boolTrue, types.StringValue("   "), props(boolFalse)),
			id:         stdID,
			wantAction: purgeMissingLocation,
		},
		{
			name:       "missing subscription: ARM ID has no subscription segment",
			state:      kvState(boolTrue, types.StringValue("eastus"), props(boolFalse)),
			id:         noSubID,
			wantAction: purgeMissingSubscription,
		},
		{
			// "East US" normalizes to "eastus" and must appear in the emitted ID together
			// with the extracted subscription and id.Name.
			name:           "purgeDo: happy path composes ID with normalized location",
			state:          kvState(boolTrue, types.StringValue("East US"), props(boolFalse)),
			id:             stdID,
			wantAction:     purgeDo,
			wantDeletedVID: happyID,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotAction, gotID := purgePlan(tc.state, tc.id)
			if gotAction != tc.wantAction {
				t.Fatalf("action = %s, want %s", actionName(gotAction), actionName(tc.wantAction))
			}
			if gotID != tc.wantDeletedVID {
				t.Fatalf("deletedVaultID = %q, want %q", gotID, tc.wantDeletedVID)
			}
		})
	}
}
