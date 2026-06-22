package customizers

import (
	"regexp"
	"strings"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/generator"
	"github.com/Azure/terraform-provider-azapi/internal/native/naming"
)

// fakeWidgetDef builds a minimal resource-group-scoped definition for exercising
// the customizer pipeline without depending on real bicep data.
func fakeWidgetDef() *generator.ResourceDefinition {
	body := &generator.Type{Kind: generator.KindObject, Properties: map[string]*generator.Property{
		"properties": {Name: "properties", Type: &generator.Type{Kind: generator.KindObject, Properties: map[string]*generator.Property{
			"minimumTlsVersion": {Name: "minimumTlsVersion", Type: &generator.Type{Kind: generator.KindString}},
			"accessTier":        {Name: "accessTier", Type: &generator.Type{Kind: generator.KindString}},
		}}},
	}}
	return &generator.ResourceDefinition{
		Name:           "Microsoft.Fake/widgets@2024-01-01",
		APIVersion:     "2024-01-01",
		Body:           body,
		WritableScopes: naming.ScopeResourceGroup,
	}
}

func TestCustomizerBakesEnvelopeAndPropertyChanges(t *testing.T) {
	const armType = "Microsoft.Fake/widgets"
	Register(armType, func(d *generator.ResourceDefinition) {
		d.Envelope.Name.Validators = []generator.DescriptionValidator{
			generator.LengthValidator(3, 24),
			generator.RegexValidator(`^[a-z0-9]+$`, "name must be lowercase alphanumeric"),
		}
		generator.FindProperty(d, "properties.minimumTlsVersion").DefaultValue = "TLS1_2"
		accessTier := generator.FindProperty(d, "properties.accessTier")
		accessTier.Validators = append(accessTier.Validators, generator.CustomValidator("StorageAccountIPRule()"))
	})
	t.Cleanup(func() { delete(registry, armType) })

	def := fakeWidgetDef()
	generator.PostProcess([]*generator.ResourceDefinition{def})
	Apply([]*generator.ResourceDefinition{def})

	// Envelope defaults (from PostProcess): a resource-group-scoped top-level type
	// references its parent as resource_group_id (not the generic parent_id).
	if def.Envelope.Parent.Name != "resource_group_id" {
		t.Errorf("parent attr = %q, want resource_group_id", def.Envelope.Parent.Name)
	}
	if len(def.Envelope.Parent.Validators) != 1 {
		t.Errorf("parent validators = %d, want 1 (RG scope pattern)", len(def.Envelope.Parent.Validators))
	}

	// Customizer effects baked into the graph (consumed at emit time, never at runtime).
	if len(def.Envelope.Name.Validators) != 2 {
		t.Errorf("name validators = %d, want 2 (from customizer)", len(def.Envelope.Name.Validators))
	}
	if p := generator.FindProperty(def, "properties.minimumTlsVersion"); p.DefaultValue != "TLS1_2" {
		t.Errorf("minimumTlsVersion default not applied by customizer: %+v", p)
	}

	// Emitted source bakes the envelope + validators + ParentAttr.
	src, err := generator.EmitSchema(def)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}
	for _, want := range []string{
		`"resource_group_id": schema.StringAttribute{`,
		`stringvalidator.LengthBetween(3, 24)`,
		"regexp.MustCompile(`^[a-z0-9]+$`)",
		`nativeschema.StaticString("TLS1_2")`,
		`StorageAccountIPRule()`,
	} {
		if !strings.Contains(src, want) {
			t.Errorf("emitted source missing %q", want)
		}
	}
	// ParentAttr is a gofmt-aligned struct field, so match across the padding.
	if !regexp.MustCompile(`ParentAttr:\s+"resource_group_id",`).MatchString(src) {
		t.Error("emitted descriptor missing ParentAttr: \"resource_group_id\"")
	}
}

func TestRegisterRejectsDuplicate(t *testing.T) {
	const armType = "Microsoft.Fake/dupes"
	Register(armType, func(*generator.ResourceDefinition) {})
	t.Cleanup(func() { delete(registry, armType) })

	defer func() {
		if recover() == nil {
			t.Error("duplicate Register did not panic")
		}
	}()
	Register(armType, func(*generator.ResourceDefinition) {})
}
