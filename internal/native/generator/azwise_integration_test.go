package generator

import (
	"regexp"
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

// TestApplyAzwiseKeyVaultKey verifies that the native generator can parse the
// latest stable ARM management-plane Key Vault key definition and lower the
// resource to the static Azapin shape callers depend on.
func TestApplyAzwiseKeyVaultKey(t *testing.T) {
	defs, ver := latestStableDefs(t, "Microsoft.KeyVault/vaults/keys")
	typegraph.PostProcess(defs)

	tag := "Microsoft.KeyVault/vaults/keys@" + ver
	var key *typegraph.ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			key = d
			break
		}
	}
	if key == nil {
		t.Fatalf("Key Vault key definition %s not found", tag)
	}

	if got := FileName(key); got != "keyvault/key_vault_key_gen.go" {
		t.Errorf("FileName() = %q, want keyvault/key_vault_key_gen.go", got)
	}
	if got := key.Envelope.Parent.Name; got != "key_vault_id" {
		t.Errorf("parent attr = %q, want key_vault_id", got)
	}
	if len(key.Envelope.Parent.Validators) != 1 {
		t.Fatalf("parent validators = %d, want 1", len(key.Envelope.Parent.Validators))
	}
	parentID := "/subscriptions/s/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/vault1"
	if !regexp.MustCompile(key.Envelope.Parent.Validators[0].Pattern).MatchString(parentID) {
		t.Errorf("parent validator should accept Key Vault ID %q", parentID)
	}

	kty := typegraph.Navigate(key.Body, "properties.kty")
	if kty == nil {
		t.Fatal("properties.kty not found")
	}
	if !kty.ForceNew {
		t.Error("expected properties.kty to be ForceNew via azwise")
	}

	keyOps := typegraph.Navigate(key.Body, "properties.keyOps")
	if keyOps == nil {
		t.Fatal("properties.keyOps not found")
	}
	if !keyOps.ForceNew {
		t.Error("expected properties.keyOps to be ForceNew because ARM Keys_CreateIfNotExist cannot update existing keys")
	}
	if keyOps.Type.Kind != typegraph.KindArray || keyOps.Type.ElementType == nil || !keyOps.Type.ElementType.IsEnum() {
		t.Fatalf("properties.keyOps type = %#v, want array of enum values", keyOps.Type)
	}
	allowed := keyOps.Type.ElementType.EnumValues()
	for _, op := range []string{"import", "release"} {
		if !contains(allowed, op) {
			t.Errorf("properties.keyOps enum missing management-plane operation %q: %v", op, allowed)
		}
	}
	for _, op := range []string{"backup", "delete"} {
		if contains(allowed, op) {
			t.Errorf("properties.keyOps enum includes non-ARM key operation %q: %v", op, allowed)
		}
	}
	if tags := typegraph.Navigate(key.Body, "tags"); tags == nil {
		t.Fatal("tags not found")
	} else if !tags.ForceNew {
		t.Error("expected tags to be ForceNew because ARM Keys_CreateIfNotExist cannot update existing keys")
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

// TestApplyAzwiseRoleAssignment verifies that the real latest-stable
// Microsoft.Authorization/roleAssignments graph preserves the ARM body contract for
// role assignments: the two ARM body fields Azure requires are emitted as Required,
// and every AzureRM immutable body field is marked ForceNew so the emitter adds
// RequiresReplace. The user-facing scope_id envelope name is a customizer concern.
func TestApplyAzwiseRoleAssignment(t *testing.T) {
	defs, ver := latestRoleAssignmentDefs(t)
	typegraph.PostProcess(defs)

	tag := "Microsoft.Authorization/roleAssignments@" + ver
	var def *typegraph.ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			def = d
			break
		}
	}
	if def == nil {
		t.Fatal("role assignment definition not found")
	}

	for _, path := range []string{"properties.roleDefinitionId", "properties.principalId"} {
		p := typegraph.Navigate(def.Body, path)
		if p == nil {
			t.Fatalf("%s not found", path)
		}
		if !p.Flags.IsRequired() {
			t.Errorf("%s should be Required", path)
		}
	}

	for _, path := range []string{
		"properties.roleDefinitionId",
		"properties.principalId",
		"properties.principalType",
		"properties.delegatedManagedIdentityResourceId",
		"properties.description",
		"properties.condition",
		"properties.conditionVersion",
	} {
		p := typegraph.Navigate(def.Body, path)
		if p == nil {
			t.Fatalf("%s not found", path)
		}
		if !p.ForceNew {
			t.Errorf("%s should be ForceNew", path)
		}
	}
}

// TestApplyAzwiseRoleDefinition verifies the Microsoft.Authorization/roleDefinitions
// overlay applies the AzureRM CustomRole type default and the not-empty roleName /
// description validators. Customizer-only effects (envelope name UUID validator,
// roleName Required promotion) are out of scope: PostProcess does not run
// customizers, matching TestApplyAzwiseWebServerFarm.
func TestApplyAzwiseRoleDefinition(t *testing.T) {
	defs, ver := latestRoleDefinitionDefs(t)
	typegraph.PostProcess(defs)

	tag := "Microsoft.Authorization/roleDefinitions@" + ver
	var def *typegraph.ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			def = d
			break
		}
	}
	if def == nil {
		t.Fatal("role definition definition not found")
	}

	if p := typegraph.Navigate(def.Body, "properties.type"); p == nil {
		t.Fatal("properties.type not found")
	} else if p.DefaultValue != "CustomRole" {
		t.Errorf("properties.type default = %q, want CustomRole", p.DefaultValue)
	}
	if p := typegraph.Navigate(def.Body, "properties.roleName"); p == nil {
		t.Fatal("properties.roleName not found")
	} else if len(p.Validators) == 0 {
		t.Error("expected properties.roleName not-empty validator from azwise")
	}
	if p := typegraph.Navigate(def.Body, "properties.description"); p == nil {
		t.Fatal("properties.description not found")
	} else if len(p.Validators) == 0 {
		t.Error("expected properties.description not-empty validator from azwise")
	}
}

// TestApplyAzwiseVirtualNetwork verifies the Microsoft.Network/virtualNetworks
// overlay applies AzureRM ForceNew (location, extendedLocation), the flow-timeout
// int range, the encryption/private-endpoint enums, the private-endpoint default,
// and the address_space/ip_address_pool ExactlyOneOf relational constraint.
func TestApplyAzwiseVirtualNetwork(t *testing.T) {
	defs, ver := latestVirtualNetworkDefs(t)
	typegraph.PostProcess(defs)

	tag := "Microsoft.Network/virtualNetworks@" + ver
	var vnet *typegraph.ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			vnet = d
			break
		}
	}
	if vnet == nil {
		t.Fatal("virtual network definition not found")
	}

	for _, path := range []string{"location", "extendedLocation"} {
		if p := typegraph.Navigate(vnet.Body, path); p == nil {
			t.Fatalf("%s not found", path)
		} else if !p.ForceNew {
			t.Errorf("expected %s to be ForceNew via azwise", path)
		}
	}

	if p := typegraph.Navigate(vnet.Body, "properties.flowTimeoutInMinutes"); p == nil {
		t.Fatal("properties.flowTimeoutInMinutes not found")
	} else if len(p.Validators) == 0 {
		t.Error("expected properties.flowTimeoutInMinutes int-range validator from azwise")
	}
	// enforcement is a bicep enum (UnionType); the emitter provides its OneOf, so
	// the azwise AllowedValues rule is intentionally a no-op — assert the enum.
	if p := typegraph.Navigate(vnet.Body, "properties.encryption.enforcement"); p == nil {
		t.Fatal("properties.encryption.enforcement not found")
	} else if !p.Type.IsEnum() {
		t.Error("expected properties.encryption.enforcement to carry bicep enum validation")
	}
	if p := typegraph.Navigate(vnet.Body, "properties.privateEndpointVNetPolicies"); p == nil {
		t.Fatal("properties.privateEndpointVNetPolicies not found")
	} else if p.DefaultValue != "Disabled" {
		t.Errorf("properties.privateEndpointVNetPolicies default = %q, want Disabled", p.DefaultValue)
	}

	// address_space and ip_address_pool are mutually exclusive (ExactlyOneOf),
	// lowered from the azwise overlay onto the resource definition.
	var exactlyOne *typegraph.RelationalConstraintDef
	for i := range vnet.Relational {
		if vnet.Relational[i].Kind == "ExactlyOneOf" {
			exactlyOne = &vnet.Relational[i]
			break
		}
	}
	if exactlyOne == nil {
		t.Fatal("expected an ExactlyOneOf relational constraint from azwise")
	}
	if len(exactlyOne.Paths) != 2 {
		t.Errorf("ExactlyOneOf Paths = %v, want 2 paths", exactlyOne.Paths)
	}

	// Both ExactlyOneOf members must opt out of UseStateForUnknown so that dropping
	// one side from config clears it (server re-owns) rather than pinning stale state.
	for _, path := range []string{
		"properties.addressSpace.addressPrefixes",
		"properties.addressSpace.ipamPoolPrefixAllocations",
	} {
		if p := typegraph.Navigate(vnet.Body, path); p == nil {
			t.Fatalf("%s not found", path)
		} else if !p.SuppressStateReuse {
			t.Errorf("expected %s to have SuppressStateReuse set (ExactlyOneOf member)", path)
		}
	}
}

// TestVirtualNetworkMutexMembersDropStateReuse verifies the emitted schema omits
// UseStateForUnknown on the ExactlyOneOf members (address_prefixes,
// ipam_pool_prefix_allocations) so an omitted side clears, while their parent
// address_space object and unrelated siblings keep the state-reuse modifier.
func TestVirtualNetworkMutexMembersDropStateReuse(t *testing.T) {
	defs, ver := latestVirtualNetworkDefs(t)
	typegraph.PostProcess(defs)

	tag := "Microsoft.Network/virtualNetworks@" + ver
	var vnet *typegraph.ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			vnet = d
			break
		}
	}
	if vnet == nil {
		t.Fatal("virtual network definition not found")
	}

	src, err := EmitSchema(vnet)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}

	// The member list attributes must be emitted without a PlanModifiers block.
	for _, decl := range []string{
		"\"address_prefixes\": schema.ListAttribute{",
		"\"ipam_pool_prefix_allocations\": schema.ListNestedAttribute{",
	} {
		i := strings.Index(src, decl)
		if i < 0 {
			t.Fatalf("declaration %q not found in emitted schema", decl)
		}
		// Inspect the attribute header up to the next nested block / element type.
		window := src[i : i+300]
		if strings.Contains(window, "UseStateForUnknown") {
			t.Errorf("%s must not emit UseStateForUnknown (ExactlyOneOf member)", decl)
		}
	}

	// The parent address_space object keeps its state-reuse modifier.
	i := strings.Index(src, "\"address_space\": schema.SingleNestedAttribute{")
	if i < 0 {
		t.Fatal("address_space declaration not found")
	}
	if !strings.Contains(src[i:i+300], "objectplanmodifier.UseStateForUnknown()") {
		t.Error("address_space object must keep UseStateForUnknown (not a constraint member)")
	}
}
