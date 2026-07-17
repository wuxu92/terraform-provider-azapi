package customizers_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/generator"
	"github.com/Azure/terraform-provider-azapi/internal/native/generator/customizers"
	"github.com/Azure/terraform-provider-azapi/internal/native/naming"
	storagevalidators "github.com/Azure/terraform-provider-azapi/internal/native/services/storage/validators"
	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
)

// fakeWidgetDef builds a minimal resource-group-scoped definition for exercising
// the customizer pipeline without depending on real bicep data.
func fakeWidgetDef() *typegraph.ResourceDefinition {
	body := &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
		"properties": {Name: "properties", Type: &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
			"minimumTlsVersion": {Name: "minimumTlsVersion", Type: &typegraph.Type{Kind: typegraph.KindString}},
			"accessTier":        {Name: "accessTier", Type: &typegraph.Type{Kind: typegraph.KindString}},
		}}},
	}}
	return &typegraph.ResourceDefinition{
		Name:           "Microsoft.Fake/widgets@2024-01-01",
		APIVersion:     "2024-01-01",
		Body:           body,
		WritableScopes: naming.ScopeResourceGroup,
	}
}

func TestCustomizerBakesEnvelopeAndPropertyChanges(t *testing.T) {
	const armType = "Microsoft.Fake/widgets"
	customizers.Register(armType, func(d *typegraph.ResourceDefinition) {
		d.Envelope.Name.Validators = []typegraph.DescriptionValidator{typegraph.
			LengthValidator(3, 24), typegraph.
			RegexValidator(`^[a-z0-9]+$`, "name must be lowercase alphanumeric"),
		}
		typegraph.
			FindProperty(d, "properties.minimumTlsVersion").DefaultValue = "TLS1_2"
		accessTier := typegraph.FindProperty(d, "properties.accessTier")
		accessTier.Validators = append(accessTier.Validators, typegraph.Validator(storagevalidators.StorageAccountIPRule))
	})
	t.Cleanup(func() { customizers.Unregister(armType) })

	def := fakeWidgetDef()
	typegraph.
		PostProcess([]*typegraph.ResourceDefinition{def})
	customizers.Apply([]*typegraph.ResourceDefinition{def})

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
	if p := typegraph.FindProperty(def, "properties.minimumTlsVersion"); p.DefaultValue != "TLS1_2" {
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
		`stringdefault.StaticString("TLS1_2")`,
		`storagevalidators.StorageAccountIPRule()`,
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

// fakeBlobServiceDef builds a minimal def carrying the server-populated read-only
// blob service properties the customizer targets.
func fakeBlobServiceDef() *typegraph.ResourceDefinition {
	body := &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
		"properties": {Name: "properties", Type: &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
			"cors": {Name: "cors", Type: &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
				"corsRules": {Name: "corsRules", Type: &typegraph.Type{Kind: typegraph.KindArray, ElementType: &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
					"allowedHeaders": {Name: "allowedHeaders", Type: &typegraph.Type{Kind: typegraph.KindArray, ElementType: &typegraph.Type{Kind: typegraph.KindString}}, Flags: typegraph.FlagRequired},
					"allowedMethods": {Name: "allowedMethods", Type: &typegraph.Type{Kind: typegraph.KindArray, ElementType: &typegraph.Type{Kind: typegraph.KindString}}, Flags: typegraph.FlagRequired},
					"allowedOrigins": {Name: "allowedOrigins", Type: &typegraph.Type{Kind: typegraph.KindArray, ElementType: &typegraph.Type{Kind: typegraph.KindString}}, Flags: typegraph.FlagRequired},
					"exposedHeaders": {Name: "exposedHeaders", Type: &typegraph.Type{Kind: typegraph.KindArray, ElementType: &typegraph.Type{Kind: typegraph.KindString}}, Flags: typegraph.FlagRequired},
				}}}},
			}}},
			"lastAccessTimeTrackingPolicy": {Name: "lastAccessTimeTrackingPolicy", Type: &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
				"name":                      {Name: "name", Type: &typegraph.Type{Kind: typegraph.KindString}},
				"blobType":                  {Name: "blobType", Type: &typegraph.Type{Kind: typegraph.KindArray, ElementType: &typegraph.Type{Kind: typegraph.KindString}}},
				"trackingGranularityInDays": {Name: "trackingGranularityInDays", Type: &typegraph.Type{Kind: typegraph.KindInt}},
			}}},
			"restorePolicy": {Name: "restorePolicy", Type: &typegraph.Type{Kind: typegraph.KindObject, Properties: map[string]*typegraph.Property{
				"minRestoreTime":  {Name: "minRestoreTime", Type: &typegraph.Type{Kind: typegraph.KindString}},
				"lastEnabledTime": {Name: "lastEnabledTime", Type: &typegraph.Type{Kind: typegraph.KindString}},
			}}},
		}}},
	}}
	return &typegraph.ResourceDefinition{
		Name:           "Microsoft.Storage/storageAccounts/blobServices@2025-06-01",
		APIVersion:     "2025-06-01",
		Body:           body,
		WritableScopes: naming.ScopeResourceGroup,
	}
}

// TestBlobServiceCustomizerMarksServerPopulatedFields guards the fixes for apply-time
// inconsistent-result drift: enabling last-access tracking / restore makes Azure
// populate read-only children that were null in prior state, so those children must
// use UseNonNullStateForUnknown while unrelated siblings keep the default state reuse.
func TestBlobServiceCustomizerMarksServerPopulatedFields(t *testing.T) {
	def := fakeBlobServiceDef()
	typegraph.
		PostProcess([]*typegraph.ResourceDefinition{def})
	customizers.Apply([]*typegraph.ResourceDefinition{def})

	for _, path := range []string{
		"properties.lastAccessTimeTrackingPolicy.name",
		"properties.lastAccessTimeTrackingPolicy.blobType",
		"properties.lastAccessTimeTrackingPolicy.trackingGranularityInDays",
		"properties.restorePolicy.minRestoreTime",
	} {
		if prop := typegraph.FindProperty(def, path); !prop.NonNullStateForUnknown {
			t.Errorf("customizer did not set NonNullStateForUnknown on %s", path)
		}
	}
	if lastEnabled := typegraph.FindProperty(def, "properties.restorePolicy.lastEnabledTime"); lastEnabled.NonNullStateForUnknown {
		t.Error("customizer must not flag restorePolicy.lastEnabledTime without evidence it transitions null -> non-null")
	}
	for _, path := range []string{
		"properties.cors.corsRules.allowedHeaders",
		"properties.cors.corsRules.allowedMethods",
		"properties.cors.corsRules.allowedOrigins",
		"properties.cors.corsRules.exposedHeaders",
	} {
		if prop := typegraph.FindProperty(def, path); !prop.UseSet {
			t.Errorf("customizer did not mark %s as a set", path)
		}
	}

	src, err := generator.EmitSchema(def)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}
	if strings.Count(src, "schema.SetAttribute") != 4 {
		t.Errorf("expected four CORS primitive arrays to emit SetAttribute, got %d", strings.Count(src, "schema.SetAttribute"))
	}
	if strings.Count(src, "stringplanmodifier.UseNonNullStateForUnknown()") != 2 {
		t.Errorf("expected two string UseNonNullStateForUnknown modifiers (name + min_restore_time), got %d", strings.Count(src, "stringplanmodifier.UseNonNullStateForUnknown()"))
	}
	if !strings.Contains(src, "listplanmodifier.UseNonNullStateForUnknown()") {
		t.Error("emitted source missing UseNonNullStateForUnknown for last_access_time_tracking_policy.blob_type")
	}
	if !strings.Contains(src, "int64planmodifier.UseNonNullStateForUnknown()") {
		t.Error("emitted source missing UseNonNullStateForUnknown for last_access_time_tracking_policy.tracking_granularity_in_days")
	}
	if strings.Count(src, "UseNonNullStateForUnknown") != 4 {
		t.Errorf("expected four UseNonNullStateForUnknown modifiers, got %d", strings.Count(src, "UseNonNullStateForUnknown"))
	}
	if !strings.Contains(src, "stringplanmodifier.UseStateForUnknown()") {
		t.Error("restore_policy.last_enabled_time must keep the default UseStateForUnknown")
	}
}
