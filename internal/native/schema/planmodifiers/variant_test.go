package planmodifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// A discriminated body with two variant blocks (read_write / read_only_following),
// each an Optional+Computed SingleNestedAttribute — the shape the generator emits
// for a discriminated root such as azapi_kusto_database.
var variantChildAttrTypes = map[string]attr.Type{"x": types.StringType}

func variantSchema() rschema.Schema {
	child := map[string]rschema.Attribute{"x": rschema.StringAttribute{Optional: true, Computed: true}}
	variant := rschema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: child}
	return rschema.Schema{Attributes: map[string]rschema.Attribute{
		"read_write":          variant,
		"read_only_following": variant,
	}}
}

// variantConfig builds a tfsdk.Config setting each named variant block to a value
// (nil entry -> the block is null in config).
func variantConfig(set map[string]bool) tfsdk.Config {
	variantTFType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"x": tftypes.String}}
	bodyTFType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"read_write":          variantTFType,
		"read_only_following": variantTFType,
	}}
	block := func(selected bool) tftypes.Value {
		if !selected {
			return tftypes.NewValue(variantTFType, nil)
		}
		return tftypes.NewValue(variantTFType, map[string]tftypes.Value{"x": tftypes.NewValue(tftypes.String, "v")})
	}
	raw := tftypes.NewValue(bodyTFType, map[string]tftypes.Value{
		"read_write":          block(set["read_write"]),
		"read_only_following": block(set["read_only_following"]),
	})
	return tfsdk.Config{Raw: raw, Schema: variantSchema()}
}

func populatedVariant(t *testing.T) types.Object {
	t.Helper()
	o, d := types.ObjectValue(variantChildAttrTypes, map[string]attr.Value{"x": types.StringValue("server")})
	if d.HasError() {
		t.Fatalf("build populated variant: %v", d)
	}
	return o
}

// TestDiscriminatedVariantClearsWhenSiblingSelected covers the reported drift: the
// unselected variant of an ExactlyOneOf root (read_only_following, null config,
// Optional+Computed => planned unknown) must resolve to null — not
// "(known after apply)" — because its sibling read_write is selected in config.
func TestDiscriminatedVariantClearsWhenSiblingSelected(t *testing.T) {
	ctx := context.Background()
	resp := &planmodifier.ObjectResponse{PlanValue: types.ObjectUnknown(variantChildAttrTypes)}

	DiscriminatedVariant("read_write").PlanModifyObject(ctx, planmodifier.ObjectRequest{
		Path:       path.Root("read_only_following"),
		Config:     variantConfig(map[string]bool{"read_write": true}),
		StateValue: types.ObjectNull(variantChildAttrTypes),
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if !resp.PlanValue.IsNull() {
		t.Fatalf("PlanValue = %s, want null (sibling read_write selected clears this variant)", resp.PlanValue)
	}
}

// TestDiscriminatedVariantClearsPriorOnSwitch covers a variant switch: this variant
// was populated in prior state, config now selects the sibling, so it must clear to
// null rather than pin the stale prior value (which would leave both variants set
// and trip the mutual-exclusion validator).
func TestDiscriminatedVariantClearsPriorOnSwitch(t *testing.T) {
	ctx := context.Background()
	resp := &planmodifier.ObjectResponse{PlanValue: types.ObjectUnknown(variantChildAttrTypes)}

	DiscriminatedVariant("continuous").PlanModifyObject(ctx, planmodifier.ObjectRequest{
		Path:       path.Root("periodic"),
		Config:     variantConfigNamed(map[string]bool{"continuous": true}, "periodic", "continuous"),
		StateValue: populatedVariant(t),
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if !resp.PlanValue.IsNull() {
		t.Fatalf("PlanValue = %s, want null (switched to sibling continuous)", resp.PlanValue)
	}
}

// TestDiscriminatedVariantReusesStateWhenNoSiblingSelected covers an AtMostOneOf
// nested block where the practitioner selects no variant and the server returns a
// default one: with no sibling selected in config, the prior (server-populated)
// state is preserved so it does not perpetually drift.
func TestDiscriminatedVariantReusesStateWhenNoSiblingSelected(t *testing.T) {
	ctx := context.Background()
	prior := populatedVariant(t)
	resp := &planmodifier.ObjectResponse{PlanValue: types.ObjectUnknown(variantChildAttrTypes)}

	DiscriminatedVariant("read_only_following").PlanModifyObject(ctx, planmodifier.ObjectRequest{
		Path:       path.Root("read_write"),
		Config:     variantConfig(map[string]bool{}),
		StateValue: prior,
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if !resp.PlanValue.Equal(prior) {
		t.Fatalf("PlanValue = %s, want prior state %s (no sibling selected)", resp.PlanValue, prior)
	}
}

// TestDiscriminatedVariantLeavesSelectedVariantKnown proves the modifier never
// touches a variant the practitioner set: when the plan value is already known
// (the variant is selected in config), it is returned unchanged.
func TestDiscriminatedVariantLeavesSelectedVariantKnown(t *testing.T) {
	ctx := context.Background()
	known := populatedVariant(t)
	resp := &planmodifier.ObjectResponse{PlanValue: known}

	DiscriminatedVariant("read_only_following").PlanModifyObject(ctx, planmodifier.ObjectRequest{
		Path:       path.Root("read_write"),
		Config:     variantConfig(map[string]bool{"read_write": true}),
		StateValue: known,
	}, resp)

	if !resp.PlanValue.Equal(known) {
		t.Fatalf("PlanValue = %s, want the configured value untouched", resp.PlanValue)
	}
}

// variantConfigNamed is variantConfig for an arbitrary pair of variant attribute
// names (the periodic/continuous nested-block case).
func variantConfigNamed(set map[string]bool, a, b string) tfsdk.Config {
	variantTFType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"x": tftypes.String}}
	bodyTFType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{a: variantTFType, b: variantTFType}}
	child := map[string]rschema.Attribute{"x": rschema.StringAttribute{Optional: true, Computed: true}}
	variant := rschema.SingleNestedAttribute{Optional: true, Computed: true, Attributes: child}
	schema := rschema.Schema{Attributes: map[string]rschema.Attribute{a: variant, b: variant}}
	block := func(selected bool) tftypes.Value {
		if !selected {
			return tftypes.NewValue(variantTFType, nil)
		}
		return tftypes.NewValue(variantTFType, map[string]tftypes.Value{"x": tftypes.NewValue(tftypes.String, "v")})
	}
	raw := tftypes.NewValue(bodyTFType, map[string]tftypes.Value{a: block(set[a]), b: block(set[b])})
	return tfsdk.Config{Raw: raw, Schema: schema}
}
