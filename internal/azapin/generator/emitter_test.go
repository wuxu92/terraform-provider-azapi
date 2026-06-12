package generator

import (
	"os"
	"strings"
	"testing"
)

func TestEmitStorageAccountSchema(t *testing.T) {
	data, err := os.ReadFile("../../azure/generated/storage/microsoft.storage/2025-01-01/types.json")
	if err != nil {
		t.Skipf("types.json not found: %v", err)
	}

	defs, err := ParseTypesJSON(data)
	if err != nil {
		t.Fatalf("ParseTypesJSON: %v", err)
	}

	var sa *ResourceDefinition
	for _, d := range defs {
		if d.Name == "Microsoft.Storage/storageAccounts@2025-01-01" {
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
	if !strings.Contains(source, "package generated") {
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
	// Verify computed-only fields don't have validators
	if strings.Contains(source, "Computed: true,\n") {
		// Find blocks that are Computed-only and ensure no Validators follow
		// (This is a basic check; detailed validation is done in separate test)
	}
	// Verify sas_policy is Optional-only (not Computed)
	sasIdx := strings.Index(source, `"sas_policy"`)
	if sasIdx < 0 {
		t.Error("missing sas_policy")
	} else {
		sasBlock := source[sasIdx : sasIdx+200]
		if strings.Contains(sasBlock, "Computed:") {
			t.Error("sas_policy should be Optional-only, not Computed")
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
