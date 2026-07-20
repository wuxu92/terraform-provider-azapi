package resource

import (
	"context"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/services"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// upgradeStateOverride wraps Base to add ResourceWithUpgradeState, one of the
// framework extension interfaces Base deliberately does not implement (a single
// shared Base cannot advertise it for every resource — Go interface satisfaction is
// static). It stands in for the identity / upgrade / move-state interfaces.
type upgradeStateOverride struct{ *Base }

func (upgradeStateOverride) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{}
}

// TestRegisterOverrideEnablesExtraInterface proves the code-driven escape hatch:
// a resource can wrap Base in a type that satisfies a framework interface Base does
// not, and New returns that wrapper while still promoting Base's default methods.
func TestRegisterOverrideEnablesExtraInterface(t *testing.T) {
	const name = "azapi_test_override"
	services.Register(services.Descriptor{Name: name, ARMType: "Microsoft.Test/things", APIVersion: "2024-01-01", Schema: testResourceGroupSchema})
	defer delete(services.Registry, name)

	// Baseline: a plain resource does NOT satisfy ResourceWithUpgradeState.
	if _, ok := New(name).(resource.ResourceWithUpgradeState); ok {
		t.Fatal("plain Base unexpectedly implements ResourceWithUpgradeState")
	}

	RegisterOverride(name, func(b *Base) resource.Resource { return upgradeStateOverride{Base: b} })
	defer delete(overrideRegistry, name)

	r := New(name)
	if _, ok := r.(resource.ResourceWithUpgradeState); !ok {
		t.Fatalf("override not applied: %T does not implement ResourceWithUpgradeState", r)
	}
	// Base methods remain promoted through the wrapper (core Resource still satisfied).
	if _, ok := r.(resource.ResourceWithConfigValidators); !ok {
		t.Errorf("wrapper %T lost promoted Base method set (ResourceWithConfigValidators)", r)
	}
}
