package generator

import (
	"strings"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
)

func TestEmitStorageAccountSchema(t *testing.T) {
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

	source, err := EmitSchema(sa)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}

	// Basic sanity checks on generated source
	if !strings.Contains(source, "package storage") {
		t.Error("missing package declaration")
	}
	if !strings.Contains(source, "AzapiStorageAccountSchema()") {
		t.Error("missing schema function")
	}
	if !strings.Contains(source, `"sku"`) {
		t.Error("missing sku attribute")
	}
	if !strings.Contains(source, `"location"`) {
		t.Error("missing location attribute")
	}
	if !strings.Contains(source, "nativeschema.UseStateForEquivalentLocation()") {
		t.Error("missing location semantic equality plan modifier")
	}
	if !strings.Contains(source, `"properties"`) {
		t.Error("missing properties attribute")
	}
	if !strings.Contains(source, `"access_tier"`) {
		t.Error("missing access_tier attribute in properties")
	}
	if !strings.Contains(source, `"minimum_tls_version"`) {
		t.Error("missing minimum_tls_version attribute")
	}
	if !strings.Contains(source, "stringvalidator.OneOf") {
		t.Error("missing enum validator")
	}
	if !strings.Contains(source, `"Hot"`) {
		t.Error("missing Hot enum value for accessTier")
	}

	// Registration: the descriptor is exposed as a named package var (StorageAccount)
	// and registered at init — not an inline Register(Descriptor{...}). It is the single
	// source of truth a same-package config builder reads as StorageAccount.Name.
	if !strings.Contains(source, "var StorageAccount = services.Descriptor{") {
		t.Error("missing exposed descriptor var StorageAccount")
	}
	if !strings.Contains(source, "func init() { services.Register(StorageAccount) }") {
		t.Error("missing init registration of the StorageAccount descriptor var")
	}
	if strings.Contains(source, "services.Register(services.Descriptor{") {
		t.Error("descriptor must be a named package var, not an inline Register(Descriptor{...})")
	}
	// Verify computed-only fields don't have validators
	// (Detailed validation done separately; basic check here)

	// Verify fully-computed blocks (like ipv6_endpoints inside primaryEndpoints)
	// are Computed-only, not Optional+Computed
	ipv6Idx := strings.Index(source, `"ipv6_endpoints"`)
	if ipv6Idx >= 0 {
		ipv6Block := source[ipv6Idx : ipv6Idx+200]
		if strings.Contains(ipv6Block, "Optional:") {
			t.Error("ipv6_endpoints should be Computed-only (all children are ReadOnly)")
		}
	}

	// Verify sas_policy is Optional+Computed (safe default — server may return it in GET)
	sasIdx := strings.Index(source, `"sas_policy"`)
	if sasIdx < 0 {
		t.Error("missing sas_policy")
	} else {
		sasBlock := source[sasIdx : sasIdx+200]
		if !strings.Contains(sasBlock, "Optional:") {
			t.Error("sas_policy should be Optional")
		}
		if !strings.Contains(sasBlock, "Computed:") {
			t.Error("sas_policy should be Optional+Computed (safe default)")
		}
	}

	// Log first 100 lines for manual review
	lines := strings.Split(source, "\n")
	maxLines := 5
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	t.Logf("Generated schema (%d lines total):\n%s\n...", len(lines), strings.Join(lines[:maxLines], "\n"))
}

func TestEmitSchemaOrdersCommonTopLevelAttributes(t *testing.T) {
	defs, ver := latestStorageDefs(t)
	typegraph.PostProcess(defs)
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

	source, err := EmitSchema(sa)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}

	assertSourceOrder(t, source,
		`"name": schema.StringAttribute{`,
		`"resource_group_id": schema.StringAttribute{`,
		`"location": schema.StringAttribute{`,
		`"sku": schema.SingleNestedAttribute{`,
		`"kind": schema.StringAttribute{`,
		`"properties": schema.SingleNestedAttribute{`,
		`"identity": nativeschema.ManagedServiceIdentity`,
		`"tags": schema.MapAttribute{`,
	)
}

func assertSourceOrder(t *testing.T, source string, needles ...string) {
	t.Helper()
	last := -1
	for _, needle := range needles {
		idx := strings.Index(source, needle)
		if idx < 0 {
			t.Fatalf("missing %q", needle)
		}
		if idx <= last {
			t.Fatalf("%q appears out of order", needle)
		}
		last = idx
	}
}

func TestEmitPlanModifiersHonorsNonNullStateForUnknown(t *testing.T) {
	for _, tc := range []struct {
		name    string
		nonNull bool
		want    string
		notWant string
	}{
		{"default", false, "stringplanmodifier.UseStateForUnknown()", "UseNonNullStateForUnknown"},
		{"opted-in", true, "stringplanmodifier.UseNonNullStateForUnknown()", "UseStateForUnknown()"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prop := &typegraph.Property{
				Name:                   "name",
				Type:                   &typegraph.Type{Kind: typegraph.KindString},
				NonNullStateForUnknown: tc.nonNull,
			}
			var b strings.Builder
			emitPlanModifiers(&b, prop, true, "")
			got := b.String()
			if !strings.Contains(got, tc.want) {
				t.Errorf("emitPlanModifiers(nonNull=%v) = %q, want substring %q", tc.nonNull, got, tc.want)
			}
			if strings.Contains(got, tc.notWant) {
				t.Errorf("emitPlanModifiers(nonNull=%v) = %q, must not contain %q", tc.nonNull, got, tc.notWant)
			}
		})
	}
}

func TestListSizeValidatorItems(t *testing.T) {
	atLeast := func(n int64) typegraph.DescriptionValidator { return typegraph.ListSizeAtLeastValidator(n) }
	atMost := func(n int64) typegraph.DescriptionValidator { return typegraph.ListSizeAtMostValidator(n) }

	for _, tc := range []struct {
		name       string
		validators []typegraph.DescriptionValidator
		want       []string
	}{
		{"none", nil, nil},
		{"min-only", []typegraph.DescriptionValidator{atLeast(1)}, []string{"listvalidator.SizeAtLeast(1)"}},
		{"max-only", []typegraph.DescriptionValidator{atMost(2)}, []string{"listvalidator.SizeAtMost(2)"}},
		{"both-collapse", []typegraph.DescriptionValidator{atLeast(1), atMost(2)}, []string{"listvalidator.SizeBetween(1, 2)"}},
		{"both-reverse-order", []typegraph.DescriptionValidator{atMost(2), atLeast(1)}, []string{"listvalidator.SizeBetween(1, 2)"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := listSizeValidatorItems(tc.validators)
			if len(got) != len(tc.want) {
				t.Fatalf("listSizeValidatorItems = %v, want %v", got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("item %d = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// TestEmitDiscriminatedRootBody confirms a resource whose body IS a discriminated
// type (e.g. Microsoft.Resources/deploymentScripts, discriminated by kind) emits
// its shared base props plus one variant block per discriminator value directly
// at the top level, wrapped by the operational envelope, and gets an ExactlyOneOf
// over the variant blocks.
func TestEmitDiscriminatedRootBody(t *testing.T) {
	strT := &typegraph.Type{Kind: typegraph.KindString}
	body := &typegraph.Type{
		Kind: typegraph.KindDiscriminated, Name: "DeploymentScript", Discriminator: "kind",
		Properties: map[string]*typegraph.Property{"location": {Name: "location", Type: strT}},
		Variants: map[string]*typegraph.Type{
			"AzureCLI":        {Kind: typegraph.KindObject, Name: "AzureCLI", Properties: map[string]*typegraph.Property{"scriptContent": {Name: "scriptContent", Type: strT}}},
			"AzurePowerShell": {Kind: typegraph.KindObject, Name: "AzurePowerShell", Properties: map[string]*typegraph.Property{"azPowerShellVersion": {Name: "azPowerShellVersion", Type: strT}}},
		},
	}
	def := &typegraph.ResourceDefinition{Name: "Microsoft.Resources/deploymentScripts@2023-08-01", Body: body}
	typegraph.PostProcess([]*typegraph.ResourceDefinition{def})

	source, err := EmitSchema(def)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}
	for _, needle := range []string{
		`"name": schema.StringAttribute{`,
		`"location": schema.StringAttribute{`,
		`"azure_cli": schema.SingleNestedAttribute{`,
		`"azure_power_shell": schema.SingleNestedAttribute{`,
		`"script_content": schema.StringAttribute{`,
		`services.ExactlyOneOf`,
	} {
		if !strings.Contains(source, needle) {
			t.Errorf("emitted source missing %q:\n%s", needle, source)
		}
	}
	// The discriminator itself is synthesized by the mapper, never a schema attribute.
	if strings.Contains(source, `"kind": schema.`) {
		t.Errorf("discriminator `kind` must not surface as a schema attribute:\n%s", source)
	}
}

// TestEmitStringLiteralAsValidatedString locks the fix for a dropped required
// constant: a standalone bicep StringLiteralType property (e.g. Cosmos DB's
// databaseAccountOfferType = "Standard") must emit a StringAttribute with a OneOf
// validator on the literal, NOT a DynamicAttribute. The mapper routes
// KindStringLiteral to a types.String; emitting a Dynamic would make expand's type
// assertion fail and silently omit the required field from the ARM payload.
func TestEmitStringLiteralAsValidatedString(t *testing.T) {
	body := &typegraph.Type{
		Kind: typegraph.KindObject, Name: "Account",
		Properties: map[string]*typegraph.Property{
			"properties": {Name: "properties", Type: &typegraph.Type{
				Kind: typegraph.KindObject, Name: "AccountProperties",
				Properties: map[string]*typegraph.Property{
					"databaseAccountOfferType": {
						Name:  "databaseAccountOfferType",
						Type:  &typegraph.Type{Kind: typegraph.KindStringLiteral, Name: "Standard"},
						Flags: typegraph.FlagRequired,
					},
				},
			}},
		},
	}
	def := &typegraph.ResourceDefinition{Name: "Microsoft.DocumentDB/databaseAccounts@2025-04-15", Body: body}
	typegraph.PostProcess([]*typegraph.ResourceDefinition{def})

	source, err := EmitSchema(def)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}
	if strings.Contains(source, `"database_account_offer_type": schema.DynamicAttribute{`) {
		t.Errorf("standalone string literal must not emit a DynamicAttribute:\n%s", source)
	}
	for _, needle := range []string{
		`"database_account_offer_type": schema.StringAttribute{`,
		`stringvalidator.OneOf(`,
		`"Standard",`,
	} {
		if !strings.Contains(source, needle) {
			t.Errorf("emitted source missing %q:\n%s", needle, source)
		}
	}
}
