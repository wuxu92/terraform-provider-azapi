package generator

import (
	"strings"
	"testing"
)

func TestNavigate(t *testing.T) {
	body := &Type{
		Kind: KindObject,
		Properties: map[string]*Property{
			"sku": {Name: "sku", Type: &Type{Kind: KindObject, Properties: map[string]*Property{
				"name": {Name: "name", Type: &Type{Kind: KindString}},
			}}},
			"properties": {Name: "properties", Type: &Type{Kind: KindObject, Properties: map[string]*Property{
				"networkAcls": {Name: "networkAcls", Type: &Type{Kind: KindObject, Properties: map[string]*Property{
					"defaultAction": {Name: "defaultAction", Type: &Type{Kind: KindString}},
				}}},
			}}},
		},
	}

	cases := []struct {
		path string
		want string // expected resolved property ARM name, "" = nil
	}{
		{"sku.name", "name"},
		{"properties.networkAcls.defaultAction", "defaultAction"},
		{"sku", "sku"},
		{"sku.missing", ""},
		{"properties.networkAcls.missing", ""},
		{"missing.path", ""},
	}
	for _, c := range cases {
		got := navigate(body, c.path)
		if c.want == "" {
			if got != nil {
				t.Errorf("navigate(%q) = %q, want nil", c.path, got.Name)
			}
			continue
		}
		if got == nil || got.Name != c.want {
			t.Errorf("navigate(%q) = %v, want %q", c.path, got, c.want)
		}
	}
}

func TestFormatAzwiseDefault(t *testing.T) {
	cases := []struct {
		v    interface{}
		typ  *Type
		want string
		ok   bool
	}{
		{true, &Type{Kind: KindBool}, "true", true},
		{false, &Type{Kind: KindBool}, "false", true},
		{float64(5), &Type{Kind: KindInt}, "5", true},
		{"TLS1_2", &Type{Kind: KindString}, "TLS1_2", true},
		{"Hot", &Type{Kind: KindUnion, Elements: []*Type{{Kind: KindStringLiteral, Name: "Hot"}}}, "Hot", true},
		{"x", &Type{Kind: KindBool}, "", false}, // wrong type
	}
	for _, c := range cases {
		got, ok := formatAzwiseDefault(c.v, c.typ)
		if ok != c.ok || got != c.want {
			t.Errorf("formatAzwiseDefault(%v, %d) = (%q, %v), want (%q, %v)", c.v, c.typ.Kind, got, ok, c.want, c.ok)
		}
	}
}

// TestApplyAzwiseStorageAccount verifies the azwise overlay flags real storage
// account properties as ForceNew and applies verified defaults.
func TestApplyAzwiseStorageAccount(t *testing.T) {
	defs, ver := latestStorageDefs(t)
	tag := "Microsoft.Storage/storageAccounts@" + ver
	var sa *ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			sa = d
			break
		}
	}
	if sa == nil {
		t.Fatal("storage account not found")
	}

	// Apply azwise overlay directly (PostProcess also calls it).
	ApplyAzwise(sa)

	// is_hns_enabled (properties.isHnsEnabled) is a known ForceNew bool.
	if p := navigate(sa.Body, "properties.isHnsEnabled"); p == nil {
		t.Error("properties.isHnsEnabled not found")
	} else if !p.ForceNew {
		t.Error("expected properties.isHnsEnabled to be ForceNew via azwise")
	}

	// minimumTlsVersion has a verified azwise default (TLS1_2).
	if p := navigate(sa.Body, "properties.minimumTlsVersion"); p == nil {
		t.Error("properties.minimumTlsVersion not found")
	} else if p.DefaultValue == "" {
		t.Error("expected properties.minimumTlsVersion to have an azwise default")
	}

	// Emitted schema should contain RequiresReplace and the plan modifier import.
	src, err := EmitSchema(sa)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}
	if !strings.Contains(src, "RequiresReplace()") {
		t.Error("expected emitted schema to contain RequiresReplace()")
	}
	if !strings.Contains(src, "boolplanmodifier") {
		t.Error("expected emitted schema to import boolplanmodifier")
	}
}

// TestApplyAzwiseBlobService verifies the azwise overlay bakes the blobServices
// validation rules into the generated body and — critically — that an azwise
// int-range rule replaces (does not stack onto) the description-mined one.
func TestApplyAzwiseBlobService(t *testing.T) {
	defs, _ := latestStorageDefs(t)
	// Full post-processing runs description mining AND the azwise overlay — the
	// combination that previously double-stacked the int-range validator.
	PostProcess(defs)

	var blob *ResourceDefinition
	for _, d := range defs {
		if strings.Contains(d.Name, "blobServices") {
			blob = d
			break
		}
	}
	if blob == nil {
		t.Fatal("blobServices definition not found")
	}

	// delete_retention_policy.days carries an int range from BOTH the ARM
	// description and the azwise IntRule; the overlay must leave exactly one.
	p := navigate(blob.Body, "properties.deleteRetentionPolicy.days")
	if p == nil {
		t.Fatal("properties.deleteRetentionPolicy.days not found")
	}
	n := 0
	for _, v := range p.Validators {
		if v.Kind == ValidatorIntRange {
			n++
		}
	}
	if n != 1 {
		t.Errorf("expected exactly 1 int-range validator on deleteRetentionPolicy.days, got %d", n)
	}

	// defaultServiceVersion is a free string in bicep; azwise adds the OneOf enum.
	dv := navigate(blob.Body, "properties.defaultServiceVersion")
	if dv == nil {
		t.Fatal("properties.defaultServiceVersion not found")
	}
	hasOneOf := false
	for _, v := range dv.Validators {
		if v.Kind == ValidatorStringOneOf {
			hasOneOf = true
		}
	}
	if !hasOneOf {
		t.Error("expected a OneOf validator on defaultServiceVersion from azwise")
	}
}

// TestDropValidators verifies the helper removes only same-kind validators so an
// authoritative azwise rule can replace a heuristic one without stacking.
func TestDropValidators(t *testing.T) {
	prop := &Property{Validators: []DescriptionValidator{
		{Kind: ValidatorIntRange},
		{Kind: ValidatorRegex},
		{Kind: ValidatorIntRange},
	}}
	dropValidators(prop, ValidatorIntRange)
	if len(prop.Validators) != 1 || prop.Validators[0].Kind != ValidatorRegex {
		t.Fatalf("dropValidators left %d validators, want 1 Regex: %+v", len(prop.Validators), prop.Validators)
	}
}

// TestApplyAzwiseResourceGroup verifies the azwise overlay flags location as
// ForceNew and bakes the managed_by non-empty rule (StringIsNotEmpty -> length)
// into the resource group body.
func TestApplyAzwiseResourceGroup(t *testing.T) {
	defs, ver := latestResourceGroupDefs(t)
	PostProcess(defs)

	tag := "Microsoft.Resources/resourceGroups@" + ver
	var rg *ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			rg = d
			break
		}
	}
	if rg == nil {
		t.Fatal("resource group definition not found")
	}

	// location is ForceNew via azwise (commonschema.Location).
	if p := navigate(rg.Body, "location"); p == nil {
		t.Fatal("location not found")
	} else if !p.ForceNew {
		t.Error("expected location to be ForceNew via azwise")
	}

	// managed_by carries a length validator from azwise (StringIsNotEmpty).
	mb := navigate(rg.Body, "managedBy")
	if mb == nil {
		t.Fatal("managedBy not found")
	}
	hasLen := false
	for _, v := range mb.Validators {
		if v.Kind == ValidatorStringLength {
			hasLen = true
		}
	}
	if !hasLen {
		t.Error("expected a length validator on managedBy from azwise")
	}
}

// TestApplyAzwiseWebServerFarm verifies the Microsoft.Web/serverfarms overlay applies
// App Service Plan ForceNew/default/validation knowledge from AzureRM.
func TestApplyAzwiseWebServerFarm(t *testing.T) {
	defs, ver := latestWebServerFarmDefs(t)
	PostProcess(defs)

	tag := "Microsoft.Web/serverfarms@" + ver
	var farm *ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			farm = d
			break
		}
	}
	if farm == nil {
		t.Fatal("web server farm definition not found")
	}

	for _, path := range []string{"location", "properties.reserved", "properties.hyperV"} {
		if p := navigate(farm.Body, path); p == nil {
			t.Fatalf("%s not found", path)
		} else if !p.ForceNew {
			t.Errorf("expected %s to be ForceNew via azwise", path)
		}
	}
	if p := navigate(farm.Body, "properties.perSiteScaling"); p == nil {
		t.Fatal("properties.perSiteScaling not found")
	} else if p.DefaultValue != "false" {
		t.Errorf("perSiteScaling default = %q, want false", p.DefaultValue)
	}
	if p := navigate(farm.Body, "sku.name"); p == nil {
		t.Fatal("sku.name not found")
	} else if len(p.Validators) == 0 {
		t.Error("expected sku.name validator from azwise")
	}
	if p := navigate(farm.Body, "sku.capacity"); p == nil {
		t.Fatal("sku.capacity not found")
	} else if len(p.Validators) == 0 {
		t.Error("expected sku.capacity int validator from azwise")
	}
}

// TestApplyAzwiseWebSite verifies the Microsoft.Web/sites overlay applies common
// App Service AzureRM defaults and validators to the shared site ARM type.
func TestApplyAzwiseWebSite(t *testing.T) {
	defs, ver := latestWebSiteDefs(t)
	PostProcess(defs)

	tag := "Microsoft.Web/sites@" + ver
	var site *ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			site = d
			break
		}
	}
	if site == nil {
		t.Fatal("web site definition not found")
	}

	if p := navigate(site.Body, "location"); p == nil {
		t.Fatal("location not found")
	} else if !p.ForceNew {
		t.Error("expected location to be ForceNew via azwise")
	}
	if p := navigate(site.Body, "properties.serverFarmId"); p == nil {
		t.Fatal("properties.serverFarmId not found")
	} else if p.Type.Kind != KindString || p.ForceComputed {
		t.Error("expected properties.serverFarmId to remain a settable string ARM body field")
	}

	if p := navigate(site.Body, "properties.httpsOnly"); p == nil {
		t.Fatal("properties.httpsOnly not found")
	} else if p.DefaultValue != "false" {
		t.Errorf("properties.httpsOnly default = %q, want false", p.DefaultValue)
	}
	if p := navigate(site.Body, "properties.clientAffinityPartitioningEnabled"); p == nil {
		t.Fatal("properties.clientAffinityPartitioningEnabled not found")
	} else if p.DefaultValue != "false" {
		t.Errorf("properties.clientAffinityPartitioningEnabled default = %q, want false", p.DefaultValue)
	}
	if p := navigate(site.Body, "properties.siteConfig.minTlsVersion"); p == nil {
		t.Fatal("properties.siteConfig.minTlsVersion not found")
	} else if p.DefaultValue != "1.2" {
		t.Errorf("minTlsVersion default = %q, want 1.2", p.DefaultValue)
	}

	mode := navigate(site.Body, "properties.clientCertMode")
	if mode == nil {
		t.Fatal("properties.clientCertMode not found")
	}
	if !mode.Type.IsEnum() {
		t.Error("expected clientCertMode to carry bicep enum validation")
	}
}

func TestEmitWebSiteSensitiveOnlyOnAzwiseLeaf(t *testing.T) {
	defs, ver := latestWebSiteDefs(t)
	PostProcess(defs)

	tag := "Microsoft.Web/sites@" + ver
	var site *ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			site = d
			break
		}
	}
	if site == nil {
		t.Fatal("web site definition not found")
	}

	if p := navigate(site.Body, "properties.siteConfig"); p == nil {
		t.Fatal("properties.siteConfig not found")
	} else if p.Sensitive {
		t.Fatal("siteConfig must not inherit sensitivity from nested connection string")
	}
	if p := navigate(site.Body, "properties.siteConfig.connectionStrings[*].connectionString"); p == nil {
		t.Fatal("connection string leaf not found")
	} else if !p.Sensitive {
		t.Fatal("connection string leaf should be sensitive via azwise")
	}

	src, err := EmitSchema(site)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}
	siteConfigIdx := strings.Index(src, `"site_config": schema.SingleNestedAttribute{`)
	if siteConfigIdx < 0 {
		t.Fatal("site_config schema block not emitted")
	}
	appSettingsIdx := strings.Index(src[siteConfigIdx:], `"app_settings": schema.ListNestedAttribute{`)
	if appSettingsIdx < 0 {
		t.Fatal("app_settings schema block not emitted")
	}
	if strings.Contains(src[siteConfigIdx:siteConfigIdx+appSettingsIdx], "Sensitive:   true") {
		t.Fatal("site_config parent block must not be emitted as Sensitive")
	}
	connectionStringIdx := strings.Index(src, `"connection_string": schema.StringAttribute{`)
	if connectionStringIdx < 0 {
		t.Fatal("connection_string schema block not emitted")
	}
	if !strings.Contains(src[connectionStringIdx:connectionStringIdx+250], "Sensitive:   true") {
		t.Fatal("connection_string leaf should be emitted as Sensitive")
	}
}
