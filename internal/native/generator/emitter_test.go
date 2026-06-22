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
