package customizers

import (
	"reflect"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/naming"
	"github.com/Azure/terraform-provider-azapi/internal/native/schema/validators"
	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
)

// refsValidator reports whether vs contains a func-referenced validator whose
// constructor is fn (compared by func pointer, the identity typegraph.Validator
// stores).
func refsValidator(vs []typegraph.DescriptionValidator, fn any) bool {
	want := reflect.ValueOf(fn).Pointer()
	for _, v := range vs {
		if v.Kind == typegraph.ValidatorFunc && v.Func != nil && reflect.ValueOf(v.Func).Pointer() == want {
			return true
		}
	}
	return false
}

func TestRegisterRejectsDuplicate(t *testing.T) {
	const armType = "Microsoft.Fake/dupes"
	Register(armType, func(*typegraph.ResourceDefinition) {})
	t.Cleanup(func() { delete(registry, armType) })

	defer func() {
		if recover() == nil {
			t.Error("duplicate Register did not panic")
		}
	}()
	Register(armType, func(*typegraph.ResourceDefinition) {})
}

func TestWebServerFarmCustomizerAddsNameIDAndSkuRules(t *testing.T) {
	def := &typegraph.ResourceDefinition{
		Name:           "Microsoft.Web/serverfarms@2025-03-01",
		WritableScopes: naming.ScopeResourceGroup,
		Body: &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
			"sku": {Name: "sku", Type: &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
				"name": {Name: "name", Type: &typegraph.Type{Kind: typegraph.KindString}},
			}}},
			"properties": {Name: "properties", Type: &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
				"hostingEnvironmentProfile": {Name: "hostingEnvironmentProfile", Type: &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
					"id": {Name: "id", Type: &typegraph.Type{Kind: typegraph.KindString}},
				}}},
			}}},
		}},
	}
	typegraph.
		PostProcess([]*typegraph.ResourceDefinition{def})
	customizeWebServerFarm(def)

	if got := len(def.Envelope.Name.Validators); got != 2 {
		t.Fatalf("name validators = %d, want 2", got)
	}
	if p := typegraph.FindProperty(def, "sku"); !p.Flags.IsRequired() {
		t.Fatal("sku should be required by customizer")
	}
	if p := typegraph.FindProperty(def, "sku.name"); !p.Flags.IsRequired() {
		t.Fatal("sku.name should be required by customizer")
	}
	p := typegraph.FindProperty(def, "properties.hostingEnvironmentProfile.id")
	if !refsValidator(p.Validators, validators.AzureResourceID) {
		t.Fatalf("ASE ID validators = %#v, want AzureResourceID validator", p.Validators)
	}
}

func TestWebSiteCustomizerAddsValidatorsAndClearsSiteConfigDefaults(t *testing.T) {
	restrictionElem := &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
		"priority": {Name: "priority", Type: &typegraph.Type{Kind: typegraph.KindInt}},
	}}
	def := &typegraph.ResourceDefinition{
		Name:           "Microsoft.Web/sites@2025-03-01",
		WritableScopes: naming.ScopeResourceGroup,
		Body: &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
			"properties": {Name: "properties", Type: &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
				"serverFarmId":           {Name: "serverFarmId", Type: &typegraph.Type{Kind: typegraph.KindString}},
				"virtualNetworkSubnetId": {Name: "virtualNetworkSubnetId", Type: &typegraph.Type{Kind: typegraph.KindString}},
				"siteConfig": {Name: "siteConfig", Type: &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
					"ftpsState": {Name: "ftpsState", Type: &typegraph.Type{Kind: typegraph.KindString}, DefaultValue: "Disabled"},
					"nested": {Name: "nested", Type: &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
						"enabled": {Name: "enabled", Type: &typegraph.Type{Kind: typegraph.KindBool}, DefaultValue: "false"},
					}}},
					"ipSecurityRestrictions":    {Name: "ipSecurityRestrictions", Type: &typegraph.Type{Kind: typegraph.KindArray, ElementType: restrictionElem}},
					"scmIpSecurityRestrictions": {Name: "scmIpSecurityRestrictions", Type: &typegraph.Type{Kind: typegraph.KindArray, ElementType: restrictionElem}},
				}}},
			}}},
		}},
	}
	typegraph.
		PostProcess([]*typegraph.ResourceDefinition{def})
	customizeWebSite(def)

	if got := len(def.Envelope.Name.Validators); got != 2 {
		t.Fatalf("name validators = %d, want 2", got)
	}
	for _, path := range []string{"properties.serverFarmId", "properties.virtualNetworkSubnetId"} {
		p := typegraph.FindProperty(def, path)
		if len(p.Validators) != 1 || !refsValidator(p.Validators, validators.AzureResourceID) {
			t.Fatalf("%s validators = %#v, want single AzureResourceID validator", path, p.Validators)
		}
	}
	for _, path := range []string{"properties.siteConfig.ftpsState", "properties.siteConfig.nested.enabled"} {
		if got := typegraph.FindProperty(def, path).DefaultValue; got != "" {
			t.Fatalf("%s default = %q, want cleared for write-only siteConfig", path, got)
		}
	}

	// Both restriction arrays share one bicep element type; the cap must land on each
	// independently (exactly one validator per path proves IsolateArrayElement un-shared
	// them rather than double-appending to a common element).
	for _, path := range []string{
		"properties.siteConfig.ipSecurityRestrictions[*].priority",
		"properties.siteConfig.scmIpSecurityRestrictions[*].priority",
	} {
		priority := typegraph.FindProperty(def, path)
		if len(priority.Validators) != 1 || priority.Validators[0].Kind != typegraph.ValidatorIntRange {
			t.Fatalf("%s validators = %#v, want one int-range validator", path, priority.Validators)
		}
		v := priority.Validators[0]
		if v.Min == nil || *v.Min != 1 || v.Max == nil || *v.Max != 2147483646 {
			t.Fatalf("%s range = [%v, %v], want [1, 2147483646]", path, v.Min, v.Max)
		}
	}
}
