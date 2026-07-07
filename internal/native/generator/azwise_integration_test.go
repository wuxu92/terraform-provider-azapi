package generator

import (
	"strings"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
)

// These are generation integration tests: they run the azwise overlay on real
// bicep fixtures (parse + PostProcess + ApplyAzwise) and assert on the emitted
// schema. They live in package generator because they exercise EmitSchema; the
// pure model-navigation tests stay in package typegraph.

// TestApplyAzwiseStorageAccount verifies the azwise overlay flags real storage
// account properties as ForceNew and applies verified defaults.
func TestApplyAzwiseStorageAccount(t *testing.T) {
	defs, ver := latestStorageDefs(t)
	tag := "Microsoft.Storage/storageAccounts@" + ver
	var sa *typegraph.ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			sa = d
			break
		}
	}
	if sa == nil {
		t.Fatal("storage account not found")
	}
	typegraph.

		// Apply azwise overlay directly (PostProcess also calls it).
		ApplyAzwise(sa)

	// is_hns_enabled (properties.isHnsEnabled) is a known ForceNew bool.
	if p := typegraph.Navigate(sa.Body, "properties.isHnsEnabled"); p == nil {
		t.Error("properties.isHnsEnabled not found")
	} else if !p.ForceNew {
		t.Error("expected properties.isHnsEnabled to be ForceNew via azwise")
	}

	// minimumTlsVersion has a verified azwise default (TLS1_2).
	if p := typegraph.Navigate(sa.Body, "properties.minimumTlsVersion"); p == nil {
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
	typegraph.
		// Full post-processing runs description mining AND the azwise overlay — the
		// combination that previously double-stacked the int-range validator.
		PostProcess(defs)

	var blob *typegraph.ResourceDefinition
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
	p := typegraph.Navigate(blob.Body, "properties.deleteRetentionPolicy.days")
	if p == nil {
		t.Fatal("properties.deleteRetentionPolicy.days not found")
	}
	n := 0
	for _, v := range p.Validators {
		if v.Kind == typegraph.ValidatorIntRange {
			n++
		}
	}
	if n != 1 {
		t.Errorf("expected exactly 1 int-range validator on deleteRetentionPolicy.days, got %d", n)
	}

	// defaultServiceVersion is a free string in bicep; azwise adds the OneOf enum.
	dv := typegraph.Navigate(blob.Body, "properties.defaultServiceVersion")
	if dv == nil {
		t.Fatal("properties.defaultServiceVersion not found")
	}
	hasOneOf := false
	for _, v := range dv.Validators {
		if v.Kind == typegraph.ValidatorStringOneOf {
			hasOneOf = true
		}
	}
	if !hasOneOf {
		t.Error("expected a OneOf validator on defaultServiceVersion from azwise")
	}
}

// TestApplyAzwiseResourceGroup verifies the azwise overlay flags location as
// ForceNew and bakes the managed_by non-empty rule (StringIsNotEmpty -> length)
// into the resource group body.
func TestApplyAzwiseResourceGroup(t *testing.T) {
	defs, ver := latestResourceGroupDefs(t)
	typegraph.PostProcess(defs)

	tag := "Microsoft.Resources/resourceGroups@" + ver
	var rg *typegraph.ResourceDefinition
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
	if p := typegraph.Navigate(rg.Body, "location"); p == nil {
		t.Fatal("location not found")
	} else if !p.ForceNew {
		t.Error("expected location to be ForceNew via azwise")
	}

	// managed_by carries a length validator from azwise (StringIsNotEmpty).
	mb := typegraph.Navigate(rg.Body, "managedBy")
	if mb == nil {
		t.Fatal("managedBy not found")
	}
	hasLen := false
	for _, v := range mb.Validators {
		if v.Kind == typegraph.ValidatorStringLength {
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
	typegraph.PostProcess(defs)

	tag := "Microsoft.Web/serverfarms@" + ver
	var farm *typegraph.ResourceDefinition
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
		if p := typegraph.Navigate(farm.Body, path); p == nil {
			t.Fatalf("%s not found", path)
		} else if !p.ForceNew {
			t.Errorf("expected %s to be ForceNew via azwise", path)
		}
	}
	if p := typegraph.Navigate(farm.Body, "properties.perSiteScaling"); p == nil {
		t.Fatal("properties.perSiteScaling not found")
	} else if p.DefaultValue != "false" {
		t.Errorf("perSiteScaling default = %q, want false", p.DefaultValue)
	}
	if p := typegraph.Navigate(farm.Body, "sku.name"); p == nil {
		t.Fatal("sku.name not found")
	} else if len(p.Validators) == 0 {
		t.Error("expected sku.name validator from azwise")
	}
	if p := typegraph.Navigate(farm.Body, "sku.capacity"); p == nil {
		t.Fatal("sku.capacity not found")
	} else if len(p.Validators) == 0 {
		t.Error("expected sku.capacity int validator from azwise")
	}
}

// TestApplyAzwiseWebSite verifies the Microsoft.Web/sites overlay applies common
// App Service AzureRM defaults and validators to the shared site ARM type.
func TestApplyAzwiseWebSite(t *testing.T) {
	defs, ver := latestWebSiteDefs(t)
	typegraph.PostProcess(defs)

	tag := "Microsoft.Web/sites@" + ver
	var site *typegraph.ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			site = d
			break
		}
	}
	if site == nil {
		t.Fatal("web site definition not found")
	}

	if p := typegraph.Navigate(site.Body, "location"); p == nil {
		t.Fatal("location not found")
	} else if !p.ForceNew {
		t.Error("expected location to be ForceNew via azwise")
	}
	if p := typegraph.Navigate(site.Body, "properties.serverFarmId"); p == nil {
		t.Fatal("properties.serverFarmId not found")
	} else if p.Type.Kind != typegraph.KindString || p.ForceComputed {
		t.Error("expected properties.serverFarmId to remain a settable string ARM body field")
	}

	if p := typegraph.Navigate(site.Body, "properties.httpsOnly"); p == nil {
		t.Fatal("properties.httpsOnly not found")
	} else if p.DefaultValue != "false" {
		t.Errorf("properties.httpsOnly default = %q, want false", p.DefaultValue)
	}
	if p := typegraph.Navigate(site.Body, "properties.clientAffinityPartitioningEnabled"); p == nil {
		t.Fatal("properties.clientAffinityPartitioningEnabled not found")
	} else if p.DefaultValue != "false" {
		t.Errorf("properties.clientAffinityPartitioningEnabled default = %q, want false", p.DefaultValue)
	}
	if p := typegraph.Navigate(site.Body, "properties.siteConfig.minTlsVersion"); p == nil {
		t.Fatal("properties.siteConfig.minTlsVersion not found")
	} else if p.DefaultValue != "1.2" {
		t.Errorf("minTlsVersion default = %q, want 1.2", p.DefaultValue)
	}

	mode := typegraph.Navigate(site.Body, "properties.clientCertMode")
	if mode == nil {
		t.Fatal("properties.clientCertMode not found")
	}
	if !mode.Type.IsEnum() {
		t.Error("expected clientCertMode to carry bicep enum validation")
	}
}

func TestEmitWebSiteSensitiveOnlyOnAzwiseLeaf(t *testing.T) {
	defs, ver := latestWebSiteDefs(t)
	typegraph.PostProcess(defs)

	tag := "Microsoft.Web/sites@" + ver
	var site *typegraph.ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			site = d
			break
		}
	}
	if site == nil {
		t.Fatal("web site definition not found")
	}

	if p := typegraph.Navigate(site.Body, "properties.siteConfig"); p == nil {
		t.Fatal("properties.siteConfig not found")
	} else if p.Sensitive {
		t.Fatal("siteConfig must not inherit sensitivity from nested connection string")
	}
	if p := typegraph.Navigate(site.Body, "properties.siteConfig.connectionStrings[*].connectionString"); p == nil {
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

// TestApplyAzwiseUserAssignedIdentity verifies the azwise overlay flags location
// as ForceNew and that the read-only identity coordinates (clientId, principalId,
// tenantId) resolve as Computed-only so StripComputedFields drops them from the
// PUT body.
func TestApplyAzwiseUserAssignedIdentity(t *testing.T) {
	defs, ver := latestUserAssignedIdentityDefs(t)
	typegraph.PostProcess(defs)

	tag := "Microsoft.ManagedIdentity/userAssignedIdentities@" + ver
	var uai *typegraph.ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			uai = d
			break
		}
	}
	if uai == nil {
		t.Fatal("user assigned identity definition not found")
	}

	// location is ForceNew via azwise (commonschema.Location).
	if p := typegraph.Navigate(uai.Body, "location"); p == nil {
		t.Fatal("location not found")
	} else if !p.ForceNew {
		t.Error("expected location to be ForceNew via azwise")
	}

	// The read-only identity coordinates are Computed-only (baked by the azwise
	// ComputedFields overlay), so StripComputedFields drops them from the PUT body.
	for _, path := range []string{"properties.clientId", "properties.principalId", "properties.tenantId"} {
		p := typegraph.Navigate(uai.Body, path)
		if p == nil {
			t.Fatalf("%s not found", path)
		}
		if !typegraph.EffectiveComputed(p) {
			t.Errorf("expected %s to be Computed-only via azwise", path)
		}
	}
}
