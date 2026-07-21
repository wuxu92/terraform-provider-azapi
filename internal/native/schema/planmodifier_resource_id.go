package schema

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// UseStateForEquivalentResourceID keeps the prior state value when the configured
// and remote ARM resource ID differ only by case. Azure frequently echoes a
// resource ID back with different casing than it was submitted (most often the
// subscription GUID and resource-provider namespace), which would otherwise plan a
// perpetual diff. This is the native equivalent of AzureRM's case-insensitive
// resource-ID diff suppression, for a generated string attribute that carries an
// ARM resource ID.
//
// Attach it from a customizer:
//
//	def.AddPlanModifiersFor("properties.someResourceId",
//	    typegraph.PlanModifier(nativeschema.UseStateForEquivalentResourceID))
func UseStateForEquivalentResourceID() planmodifier.String {
	return equivalentResourceIDPlanModifier{}
}

type equivalentResourceIDPlanModifier struct{}

func (equivalentResourceIDPlanModifier) Description(context.Context) string {
	return "Use the prior state when the ARM resource ID differs only by case."
}

func (m equivalentResourceIDPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (equivalentResourceIDPlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	if strings.EqualFold(req.ConfigValue.ValueString(), req.StateValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}

var _ planmodifier.String = equivalentResourceIDPlanModifier{}
