package resource

import (
	"context"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/services"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// TestConfigValidatorsSourcedFromHooks locks the seam moved out of the generated
// schema: resource-level relational constraints are now declared in a resource's
// hand-written Hooks.Relational (not baked into the Descriptor / _gen.go), and
// Base.ConfigValidators builds one framework ConfigValidator per declared
// constraint. A resource with no hooks (or no Relational) contributes none.
func TestConfigValidatorsSourcedFromHooks(t *testing.T) {
	ctx := context.Background()
	desc := services.Descriptor{Name: "azapi_test_rel", Schema: testResourceGroupSchema}

	// No hooks at all: no relational validators.
	b := &Base{desc: desc}
	if got := b.ConfigValidators(ctx); got != nil {
		t.Errorf("nil hooks: ConfigValidators = %v, want nil", got)
	}

	// Hooks present but no Relational: still none.
	b = &Base{desc: desc, hooks: &Hooks{}}
	if got := b.ConfigValidators(ctx); got != nil {
		t.Errorf("empty hooks.Relational: ConfigValidators = %v, want nil", got)
	}

	// Two declared constraints -> two validators, preserving kind + paths.
	b = &Base{desc: desc, hooks: &Hooks{Relational: []services.RelationalConstraint{
		{Kind: services.RequiredWith, Paths: []string{"properties.a", "properties.b"}},
		{Kind: services.ExactlyOneOf, Paths: []string{"x", "y"}},
	}}}
	got := b.ConfigValidators(ctx)
	if len(got) != 2 {
		t.Fatalf("ConfigValidators len = %d, want 2", len(got))
	}
	rv, ok := got[0].(relationalValidator)
	if !ok {
		t.Fatalf("validator[0] type = %T, want relationalValidator", got[0])
	}
	if rv.constraint.Kind != services.RequiredWith {
		t.Errorf("validator[0].Kind = %v, want RequiredWith", rv.constraint.Kind)
	}
	// Hooks.ConfigValidators are appended verbatim, after the relational-derived
	// ones, so a resource can plug prebuilt/reusable validator objects.
	custom := stubConfigValidator{}
	b = &Base{desc: desc, hooks: &Hooks{
		Relational:       []services.RelationalConstraint{{Kind: services.ExactlyOneOf, Paths: []string{"x", "y"}}},
		ConfigValidators: []resource.ConfigValidator{custom},
	}}
	got = b.ConfigValidators(ctx)
	if len(got) != 2 {
		t.Fatalf("relational+custom: len = %d, want 2", len(got))
	}
	if _, ok := got[0].(relationalValidator); !ok {
		t.Errorf("validator[0] = %T, want relationalValidator (relational comes first)", got[0])
	}
	if _, ok := got[1].(stubConfigValidator); !ok {
		t.Errorf("validator[1] = %T, want stubConfigValidator (custom comes last)", got[1])
	}

	// Only a custom validator, no Relational: still surfaced.
	b = &Base{desc: desc, hooks: &Hooks{ConfigValidators: []resource.ConfigValidator{custom}}}
	if got := b.ConfigValidators(ctx); len(got) != 1 {
		t.Errorf("custom-only: len = %d, want 1", len(got))
	}
}

// stubConfigValidator is a minimal prebuilt resource.ConfigValidator used to prove
// Hooks.ConfigValidators objects are passed through unchanged.
type stubConfigValidator struct{}

func (stubConfigValidator) Description(context.Context) string         { return "stub" }
func (stubConfigValidator) MarkdownDescription(context.Context) string { return "stub" }
func (stubConfigValidator) ValidateResource(context.Context, resource.ValidateConfigRequest, *resource.ValidateConfigResponse) {
}
