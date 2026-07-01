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

// fakeBlobServiceDef builds a minimal def carrying the server-populated read-only
// blob service properties the customizer targets.
func fakeBlobServiceDef() *generator.ResourceDefinition {
	body := &generator.Type{Kind: generator.KindObject, Properties: map[string]*generator.Property{
		"properties": {Name: "properties", Type: &generator.Type{Kind: generator.KindObject, Properties: map[string]*generator.Property{
			"cors": {Name: "cors", Type: &generator.Type{Kind: generator.KindObject, Properties: map[string]*generator.Property{
				"corsRules": {Name: "corsRules", Type: &generator.Type{Kind: generator.KindArray, ElementType: &generator.Type{Kind: generator.KindObject, Properties: map[string]*generator.Property{
					"allowedHeaders": {Name: "allowedHeaders", Type: &generator.Type{Kind: generator.KindArray, ElementType: &generator.Type{Kind: generator.KindString}}, Flags: generator.FlagRequired},
					"allowedMethods": {Name: "allowedMethods", Type: &generator.Type{Kind: generator.KindArray, ElementType: &generator.Type{Kind: generator.KindString}}, Flags: generator.FlagRequired},
					"allowedOrigins": {Name: "allowedOrigins", Type: &generator.Type{Kind: generator.KindArray, ElementType: &generator.Type{Kind: generator.KindString}}, Flags: generator.FlagRequired},
					"exposedHeaders": {Name: "exposedHeaders", Type: &generator.Type{Kind: generator.KindArray, ElementType: &generator.Type{Kind: generator.KindString}}, Flags: generator.FlagRequired},
				}}}},
			}}},
			"lastAccessTimeTrackingPolicy": {Name: "lastAccessTimeTrackingPolicy", Type: &generator.Type{Kind: generator.KindObject, Properties: map[string]*generator.Property{
				"name":                      {Name: "name", Type: &generator.Type{Kind: generator.KindString}},
				"blobType":                  {Name: "blobType", Type: &generator.Type{Kind: generator.KindArray, ElementType: &generator.Type{Kind: generator.KindString}}},
				"trackingGranularityInDays": {Name: "trackingGranularityInDays", Type: &generator.Type{Kind: generator.KindInt}},
			}}},
			"restorePolicy": {Name: "restorePolicy", Type: &generator.Type{Kind: generator.KindObject, Properties: map[string]*generator.Property{
				"minRestoreTime":  {Name: "minRestoreTime", Type: &generator.Type{Kind: generator.KindString}},
				"lastEnabledTime": {Name: "lastEnabledTime", Type: &generator.Type{Kind: generator.KindString}},
			}}},
		}}},
	}}
	return &generator.ResourceDefinition{
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
	generator.PostProcess([]*generator.ResourceDefinition{def})
	Apply([]*generator.ResourceDefinition{def})

	for _, path := range []string{
		"properties.lastAccessTimeTrackingPolicy.name",
		"properties.lastAccessTimeTrackingPolicy.blobType",
		"properties.lastAccessTimeTrackingPolicy.trackingGranularityInDays",
		"properties.restorePolicy.minRestoreTime",
	} {
		if prop := generator.FindProperty(def, path); !prop.NonNullStateForUnknown {
			t.Errorf("customizer did not set NonNullStateForUnknown on %s", path)
		}
	}
	if lastEnabled := generator.FindProperty(def, "properties.restorePolicy.lastEnabledTime"); lastEnabled.NonNullStateForUnknown {
		t.Error("customizer must not flag restorePolicy.lastEnabledTime without evidence it transitions null -> non-null")
	}
	for _, path := range []string{
		"properties.cors.corsRules.allowedHeaders",
		"properties.cors.corsRules.allowedMethods",
		"properties.cors.corsRules.allowedOrigins",
		"properties.cors.corsRules.exposedHeaders",
	} {
		if prop := generator.FindProperty(def, path); !prop.UseSet {
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

func TestWebServerFarmCustomizerAddsNameIDAndSkuRules(t *testing.T) {
	def := &generator.ResourceDefinition{
		Name:           "Microsoft.Web/serverfarms@2025-03-01",
		WritableScopes: naming.ScopeResourceGroup,
		Body: &generator.Type{Kind: generator.KindObject, Properties: map[string]*generator.Property{
			"sku": {Name: "sku", Type: &generator.Type{Kind: generator.KindObject, Properties: map[string]*generator.Property{
				"name": {Name: "name", Type: &generator.Type{Kind: generator.KindString}},
			}}},
			"properties": {Name: "properties", Type: &generator.Type{Kind: generator.KindObject, Properties: map[string]*generator.Property{
				"hostingEnvironmentProfile": {Name: "hostingEnvironmentProfile", Type: &generator.Type{Kind: generator.KindObject, Properties: map[string]*generator.Property{
					"id": {Name: "id", Type: &generator.Type{Kind: generator.KindString}},
				}}},
			}}},
		}},
	}
	generator.PostProcess([]*generator.ResourceDefinition{def})
	customizeWebServerFarm(def)

	if got := len(def.Envelope.Name.Validators); got != 2 {
		t.Fatalf("name validators = %d, want 2", got)
	}
	if p := generator.FindProperty(def, "sku"); !p.Flags.IsRequired() {
		t.Fatal("sku should be required by customizer")
	}
	if p := generator.FindProperty(def, "sku.name"); !p.Flags.IsRequired() {
		t.Fatal("sku.name should be required by customizer")
	}
	p := generator.FindProperty(def, "properties.hostingEnvironmentProfile.id")
	hasSharedID := false
	for _, v := range p.Validators {
		if v.Kind == generator.ValidatorShared && v.Call == "AzureResourceID()" {
			hasSharedID = true
		}
	}
	if !hasSharedID {
		t.Fatalf("ASE ID validators = %#v, want AzureResourceID shared validator", p.Validators)
	}
}

func TestWebSiteCustomizerAddsValidatorsAndClearsSiteConfigDefaults(t *testing.T) {
	restrictionElem := &generator.Type{Kind: generator.KindObject, Properties: map[string]*generator.Property{
		"priority": {Name: "priority", Type: &generator.Type{Kind: generator.KindInt}},
	}}
	def := &generator.ResourceDefinition{
		Name:           "Microsoft.Web/sites@2025-03-01",
		WritableScopes: naming.ScopeResourceGroup,
		Body: &generator.Type{Kind: generator.KindObject, Properties: map[string]*generator.Property{
			"properties": {Name: "properties", Type: &generator.Type{Kind: generator.KindObject, Properties: map[string]*generator.Property{
				"serverFarmId":           {Name: "serverFarmId", Type: &generator.Type{Kind: generator.KindString}},
				"virtualNetworkSubnetId": {Name: "virtualNetworkSubnetId", Type: &generator.Type{Kind: generator.KindString}},
				"siteConfig": {Name: "siteConfig", Type: &generator.Type{Kind: generator.KindObject, Properties: map[string]*generator.Property{
					"ftpsState": {Name: "ftpsState", Type: &generator.Type{Kind: generator.KindString}, DefaultValue: "Disabled"},
					"nested": {Name: "nested", Type: &generator.Type{Kind: generator.KindObject, Properties: map[string]*generator.Property{
						"enabled": {Name: "enabled", Type: &generator.Type{Kind: generator.KindBool}, DefaultValue: "false"},
					}}},
					"ipSecurityRestrictions":    {Name: "ipSecurityRestrictions", Type: &generator.Type{Kind: generator.KindArray, ElementType: restrictionElem}},
					"scmIpSecurityRestrictions": {Name: "scmIpSecurityRestrictions", Type: &generator.Type{Kind: generator.KindArray, ElementType: restrictionElem}},
				}}},
			}}},
		}},
	}
	generator.PostProcess([]*generator.ResourceDefinition{def})
	customizeWebSite(def)

	if got := len(def.Envelope.Name.Validators); got != 2 {
		t.Fatalf("name validators = %d, want 2", got)
	}
	for _, path := range []string{"properties.serverFarmId", "properties.virtualNetworkSubnetId"} {
		p := generator.FindProperty(def, path)
		if len(p.Validators) != 1 || p.Validators[0].Kind != generator.ValidatorShared || p.Validators[0].Call != "AzureResourceID()" {
			t.Fatalf("%s validators = %#v, want AzureResourceID shared validator", path, p.Validators)
		}
	}
	for _, path := range []string{"properties.siteConfig.ftpsState", "properties.siteConfig.nested.enabled"} {
		if got := generator.FindProperty(def, path).DefaultValue; got != "" {
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
		priority := generator.FindProperty(def, path)
		if len(priority.Validators) != 1 || priority.Validators[0].Kind != generator.ValidatorIntRange {
			t.Fatalf("%s validators = %#v, want one int-range validator", path, priority.Validators)
		}
		v := priority.Validators[0]
		if v.Min == nil || *v.Min != 1 || v.Max == nil || *v.Max != 2147483646 {
			t.Fatalf("%s range = [%v, %v], want [1, 2147483646]", path, v.Min, v.Max)
		}
	}
}
