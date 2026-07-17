package schema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DiscriminatedVariant returns the plan modifier for one variant block of a
// discriminated type. A variant block is Optional+Computed (it may carry
// server-populated leaf values), so an unselected variant with a null config
// leaves the framework planning it as "(known after apply)". A plain
// UseStateForUnknown would fix the perpetual-diff for the steady state, but it
// would also pin the previously-selected variant when the practitioner switches
// to a sibling — leaving two variants populated and tripping the ExactlyOneOf /
// AtMostOneOf validator.
//
// This modifier threads that needle: when the framework leaves this variant
// unknown (null config), it reuses the prior state UNLESS a sibling variant is
// selected in config, in which case the practitioner is switching away from this
// variant, so it clears to null. siblings are the sibling variant blocks' snake
// attribute names, resolved relative to this attribute's parent so it works for a
// discriminated root and a nested discriminated block alike.
func DiscriminatedVariant(siblings ...string) planmodifier.Object {
	return discriminatedVariantPlanModifier{siblings: siblings}
}

type discriminatedVariantPlanModifier struct {
	siblings []string
}

func (discriminatedVariantPlanModifier) Description(context.Context) string {
	return "Reuses prior state for an unselected discriminated variant, clearing it when a sibling variant is selected."
}

func (m discriminatedVariantPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m discriminatedVariantPlanModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// Only act when the framework left the value unknown: this variant is not set
	// in config (Optional+Computed). A variant selected in config is known and
	// must be left exactly as the practitioner wrote it.
	if !resp.PlanValue.IsUnknown() {
		return
	}
	// A sibling variant selected in config means this variant is being switched
	// away from (or was never selected): it must clear to null rather than pin
	// stale state.
	parent := req.Path.ParentPath()
	for _, sibling := range m.siblings {
		var sib types.Object
		if diags := req.Config.GetAttribute(ctx, parent.AtName(sibling), &sib); diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if !sib.IsNull() && !sib.IsUnknown() {
			resp.PlanValue = types.ObjectNull(resp.PlanValue.AttributeTypes(ctx))
			return
		}
	}
	// No sibling selected: preserve the prior state (a null steady state, or a
	// server-populated default variant on an AtMostOneOf block).
	resp.PlanValue = req.StateValue
}

var _ planmodifier.Object = discriminatedVariantPlanModifier{}
