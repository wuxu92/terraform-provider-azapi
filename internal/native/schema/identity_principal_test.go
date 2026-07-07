package schema

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// identityPlanWithType builds a real tfsdk.Plan whose Raw carries an identity
// block with the given known `type`; every other attribute is null. The modifier
// under test reads the sibling identity.type via req.Plan.GetAttribute, so the
// Plan must be decodable at path identity.type against the identity schema.
func identityPlanWithType(ctx context.Context, t *testing.T, idType string) tfsdk.Plan {
	t.Helper()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"identity": ManagedServiceIdentity(false, ""),
		},
	}

	tfType := s.Type().TerraformType(ctx).(tftypes.Object)
	identTfType := tfType.AttributeTypes["identity"].(tftypes.Object)

	identVals := make(map[string]tftypes.Value, len(identTfType.AttributeTypes))
	for name, at := range identTfType.AttributeTypes {
		identVals[name] = tftypes.NewValue(at, nil) // null
	}
	identVals["type"] = tftypes.NewValue(tftypes.String, idType)

	raw := tftypes.NewValue(tfType, map[string]tftypes.Value{
		"identity": tftypes.NewValue(identTfType, identVals),
	})

	return tfsdk.Plan{Schema: s, Raw: raw}
}

// TestUseStateForSystemPrincipalPlanModifier locks Fix 2: the identity
// system-assigned principal/tenant plan modifier. It drives
// systemPrincipalPlanModifier.PlanModifyString directly against a real
// tfsdk.Plan so the modifier can resolve the sibling identity.type.
func TestUseStateForSystemPrincipalPlanModifier(t *testing.T) {
	ctx := context.Background()

	// principal_id is a top-level identity attribute; ParentPath().AtName("type")
	// must therefore resolve to identity.type.
	reqPath := path.Root("identity").AtName("principal_id")

	for _, tc := range []struct {
		name        string
		idType      string
		configValue types.String
		planValue   types.String
		stateValue  types.String
		// assert inspects the post-modify resp.PlanValue.
		assert func(t *testing.T, got types.String)
	}{
		{
			// Case 1: UserAssigned-only identity, nothing in state. Azure never
			// populates the system principal, so the modifier must pin a KNOWN
			// null instead of leaving "(known after apply)". This is the storage
			// account bug.
			name:        "user_assigned known-null pin (storage bug)",
			idType:      "UserAssigned",
			configValue: types.StringNull(),
			planValue:   types.StringUnknown(),
			stateValue:  types.StringNull(),
			assert: func(t *testing.T, got types.String) {
				if !got.IsNull() || got.IsUnknown() {
					t.Fatalf("PlanValue = %s, want KNOWN null", got.String())
				}
			},
		},
		{
			// Case 2: SystemAssigned -> UserAssigned transition. State still holds
			// the stale system-assigned GUID; the modifier must clear it to a
			// KNOWN null (the web site bug).
			name:        "user_assigned clears stale system principal (web bug)",
			idType:      "UserAssigned",
			configValue: types.StringNull(),
			planValue:   types.StringUnknown(),
			stateValue:  types.StringValue("stale-guid"),
			assert: func(t *testing.T, got types.String) {
				if !got.IsNull() || got.IsUnknown() {
					t.Fatalf("PlanValue = %s, want KNOWN null", got.String())
				}
			},
		},
		{
			// Case 3: SystemAssigned with a known prior state — hold it
			// (UseNonNullStateForUnknown behavior preserved).
			name:        "system_assigned holds known prior state",
			idType:      "SystemAssigned",
			configValue: types.StringNull(),
			planValue:   types.StringUnknown(),
			stateValue:  types.StringValue("guid"),
			assert: func(t *testing.T, got types.String) {
				if !got.Equal(types.StringValue("guid")) {
					t.Fatalf("PlanValue = %s, want prior state \"guid\"", got.String())
				}
			},
		},
		{
			// Case 4: SystemAssigned first create — no prior state. Leave the
			// value unknown so Azure fills it on apply.
			name:        "system_assigned leaves unknown on first create",
			idType:      "SystemAssigned",
			configValue: types.StringNull(),
			planValue:   types.StringUnknown(),
			stateValue:  types.StringNull(),
			assert: func(t *testing.T, got types.String) {
				if !got.IsUnknown() {
					t.Fatalf("PlanValue = %s, want UNKNOWN", got.String())
				}
			},
		},
		{
			// Case 5: a known planned value is never touched (early return before
			// any sibling lookup).
			name:        "known plan value untouched",
			idType:      "UserAssigned",
			configValue: types.StringNull(),
			planValue:   types.StringValue("x"),
			stateValue:  types.StringNull(),
			assert: func(t *testing.T, got types.String) {
				if !got.Equal(types.StringValue("x")) {
					t.Fatalf("PlanValue = %s, want untouched \"x\"", got.String())
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			modifier := UseStateForSystemPrincipal()

			req := planmodifier.StringRequest{
				Path:        reqPath,
				Plan:        identityPlanWithType(ctx, t, tc.idType),
				ConfigValue: tc.configValue,
				PlanValue:   tc.planValue,
				StateValue:  tc.stateValue,
			}

			// The framework pre-populates resp.PlanValue with the proposed plan
			// value before running modifiers; mirror that here.
			resp := &planmodifier.StringResponse{PlanValue: req.PlanValue}

			modifier.PlanModifyString(ctx, req, resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
			}

			tc.assert(t, resp.PlanValue)
		})
	}
}
