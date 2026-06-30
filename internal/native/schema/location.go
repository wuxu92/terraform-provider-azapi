package schema

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// NormalizeLocation transforms human-readable Azure region names (for example,
// "West US") into Azure's canonical comparison form ("westus"). This mirrors
// AzureRM's location.Normalize helper so native resources do not plan changes
// when Azure returns a normalized location string.
func NormalizeLocation(input string) string {
	return strings.ReplaceAll(strings.ToLower(input), " ", "")
}

// UseStateForEquivalentLocation keeps the prior state value when the configured
// and remote location values normalize to the same Azure region. Terraform Plugin
// Framework has no SDKv2 DiffSuppressFunc hook; this plan modifier is the native
// equivalent for generated top-level location attributes.
func UseStateForEquivalentLocation() planmodifier.String {
	return equivalentLocationPlanModifier{}
}

type equivalentLocationPlanModifier struct{}

func (equivalentLocationPlanModifier) Description(context.Context) string {
	return "Use the prior state when Azure location names differ only by case or spaces."
}

func (m equivalentLocationPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (equivalentLocationPlanModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	if NormalizeLocation(req.ConfigValue.ValueString()) == NormalizeLocation(req.StateValue.ValueString()) {
		resp.PlanValue = req.StateValue
	}
}

var _ planmodifier.String = equivalentLocationPlanModifier{}
