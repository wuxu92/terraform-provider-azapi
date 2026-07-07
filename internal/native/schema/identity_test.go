package schema

import (
	"fmt"
	"strings"
	"testing"

	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestManagedServiceIdentitySystemPrincipalModifier(t *testing.T) {
	identity := ManagedServiceIdentity(false, "identity")

	for _, name := range []string{"principal_id", "tenant_id"} {
		attr := identity.Attributes[name].(fwschema.StringAttribute)
		if !attr.Computed || attr.Optional || attr.Required {
			t.Fatalf("%s flags = optional:%t computed:%t required:%t, want computed-only", name, attr.Optional, attr.Computed, attr.Required)
		}
		if len(attr.PlanModifiers) != 1 || !strings.Contains(fmt.Sprintf("%T", attr.PlanModifiers[0]), "systemPrincipalPlanModifier") {
			t.Fatalf("%s plan modifier = %#v, want systemPrincipalPlanModifier", name, attr.PlanModifiers)
		}
	}
}

func TestManagedServiceIdentityTypeFlags(t *testing.T) {
	optional := ManagedServiceIdentity(false, "")
	optionalType := optional.Attributes["type"].(fwschema.StringAttribute)
	if !optionalType.Optional || !optionalType.Computed || optionalType.Required {
		t.Fatalf("optional type flags = optional:%t computed:%t required:%t", optionalType.Optional, optionalType.Computed, optionalType.Required)
	}

	required := ManagedServiceIdentity(true, "")
	requiredType := required.Attributes["type"].(fwschema.StringAttribute)
	if !requiredType.Required || requiredType.Optional || requiredType.Computed {
		t.Fatalf("required type flags = optional:%t computed:%t required:%t", requiredType.Optional, requiredType.Computed, requiredType.Required)
	}
}

func TestManagedServiceIdentityDescriptionDefaults(t *testing.T) {
	defaulted := ManagedServiceIdentity(false, "")
	if defaulted.Description != defaultManagedServiceIdentityDescription {
		t.Fatalf("default description = %q, want %q", defaulted.Description, defaultManagedServiceIdentityDescription)
	}

	custom := ManagedServiceIdentity(false, "Custom identity.")
	if custom.Description != "Custom identity." {
		t.Fatalf("custom description = %q", custom.Description)
	}
}
