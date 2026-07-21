package planmodifiers

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNormalizeLocation(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  string
	}{
		{input: "West US", want: "westus"},
		{input: "westus", want: "westus"},
		{input: "North Europe", want: "northeurope"},
		{input: "  East US 2 ", want: "eastus2"},
	} {
		if got := NormalizeLocation(tc.input); got != tc.want {
			t.Fatalf("NormalizeLocation(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestUseStateForEquivalentLocationSuppressesEquivalentPlan(t *testing.T) {
	modifier := UseStateForEquivalentLocation()
	resp := &planmodifier.StringResponse{}

	modifier.PlanModifyString(context.Background(), planmodifier.StringRequest{
		ConfigValue: types.StringValue("West US"),
		PlanValue:   types.StringValue("West US"),
		StateValue:  types.StringValue("westus"),
	}, resp)

	if !resp.PlanValue.Equal(types.StringValue("westus")) {
		t.Fatalf("PlanValue = %s, want prior state westus", resp.PlanValue.String())
	}
}

func TestUseStateForEquivalentLocationLeavesDifferentPlan(t *testing.T) {
	modifier := UseStateForEquivalentLocation()
	resp := &planmodifier.StringResponse{PlanValue: types.StringValue("East US")}

	modifier.PlanModifyString(context.Background(), planmodifier.StringRequest{
		ConfigValue: types.StringValue("East US"),
		PlanValue:   types.StringValue("East US"),
		StateValue:  types.StringValue("westus"),
	}, resp)

	if !resp.PlanValue.Equal(types.StringValue("East US")) {
		t.Fatalf("PlanValue = %s, want configured East US", resp.PlanValue.String())
	}
}
