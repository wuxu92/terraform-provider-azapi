package generator

import (
	"strings"
	"testing"
)

func TestEmitStorageAccountSchema(t *testing.T) {
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
	maxLines := 80
	if len(lines) < maxLines {
		maxLines = len(lines)
	}
	t.Logf("Generated schema (%d lines total):\n%s\n...", len(lines), strings.Join(lines[:maxLines], "\n"))
}

func TestEmitSchemaOrdersCommonTopLevelAttributes(t *testing.T) {
	defs, ver := latestStorageDefs(t)
	PostProcess(defs)
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
			prop := &Property{
				Name:                   "name",
				Type:                   &Type{Kind: KindString},
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
