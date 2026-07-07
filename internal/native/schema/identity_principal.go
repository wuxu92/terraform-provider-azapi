package schema

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// systemAssignedComponent is the substring in an identity `type` that indicates
// a system-assigned identity is present ("SystemAssigned" or the combined
// "SystemAssigned, UserAssigned"). Only then does Azure populate the top-level
// principal_id / tenant_id.
const systemAssignedComponent = "SystemAssigned"

// UseStateForSystemPrincipal is the plan modifier for the identity block's
// top-level principal_id / tenant_id — the SYSTEM-assigned identity's coordinates.
// Azure populates them only when the identity type includes SystemAssigned; for a
// UserAssigned-only (or None) identity they are absent from every response, so they
// stay null in state.
//
// Plain UseNonNullStateForUnknown no-ops on a null prior state, leaving the framework's
// "(known after apply)" unknown in the plan, which never resolves to a value the server
// supplies — a perpetual diff on every UserAssigned-only identity. This modifier splits
// on the planned identity type:
//
//   - type includes SystemAssigned: Azure fills these values, so preserve the
//     UseNonNullStateForUnknown behavior — copy a known, non-null prior state into the
//     plan; leave the value unknown on first create so Azure can populate it.
//   - type is UserAssigned / None (no system component): the server never populates
//     these, so pin the plan to a KNOWN null. This both settles the create/no-op case
//     (null == null, no drift) and clears a stale system principal when transitioning
//     SystemAssigned -> UserAssigned (Azure drops it; state must not keep the old GUID).
func UseStateForSystemPrincipal() planmodifier.String {
	return systemPrincipalPlanModifier{}
}

type systemPrincipalPlanModifier struct{}

func (systemPrincipalPlanModifier) Description(context.Context) string {
	return "Hold the system-assigned principal/tenant ID stable, pinning null when the identity type has no system-assigned component."
}

func (m systemPrincipalPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (systemPrincipalPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// A known planned value (rare for a Computed-only field) or an unknown config
	// interpolation is left untouched.
	if !req.PlanValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}

	// Read the sibling identity `type` from the proposed plan. It is user-configured,
	// so it is known here regardless of plan-modifier traversal order.
	var idType types.String
	diags := req.Plan.GetAttribute(ctx, req.Path.ParentPath().AtName("type"), &idType)
	if diags.HasError() || idType.IsNull() || idType.IsUnknown() {
		// Type indeterminate: fall back to UseNonNullStateForUnknown semantics so we
		// never regress the SystemAssigned path.
		if !req.StateValue.IsNull() {
			resp.PlanValue = req.StateValue
		}
		return
	}

	if strings.Contains(idType.ValueString(), systemAssignedComponent) {
		// Azure populates this field; hold a known prior value, otherwise leave it
		// unknown for the create that first fills it.
		if !req.StateValue.IsNull() {
			resp.PlanValue = req.StateValue
		}
		return
	}

	// No system-assigned component: the server never returns these, so pin a known
	// null. Avoids the perpetual "(known after apply)" and clears any stale value
	// carried from a prior SystemAssigned type.
	resp.PlanValue = types.StringNull()
}

var _ planmodifier.String = systemPrincipalPlanModifier{}
