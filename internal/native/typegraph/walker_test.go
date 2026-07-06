package typegraph

import (
	"os"
	"testing"
)

func TestParseKeyVault(t *testing.T) {
	data, err := os.ReadFile("../../azure/generated/keyvault/microsoft.keyvault/2023-07-01/types.json")
	if err != nil {
		t.Skipf("types.json not found: %v", err)
	}

	defs, err := ParseTypesJSON(data)
	if err != nil {
		t.Fatalf("ParseTypesJSON: %v", err)
	}

	// Find vault
	var vault *ResourceDefinition
	for _, d := range defs {
		if d.APIVersion == "2023-07-01" && !containsSubResource(d.Name) {
			vault = d
			t.Logf("Found: %s", d.Name)
			break
		}
	}
	if vault == nil {
		// Try any vault definition
		for _, d := range defs {
			t.Logf("Available: %s", d.Name)
		}
		t.Skip("no vault definition found")
	}

	if vault.Body.Kind != KindObject {
		t.Fatalf("body kind = %d, want KindObject", vault.Body.Kind)
	}
	t.Logf("Vault properties: %d top-level", len(vault.Body.Properties))
}

func containsSubResource(name string) bool {
	// Count slashes in the resource type part (before @)
	parts := name
	if at := len(name) - 1; at >= 0 {
		for i, c := range name {
			if c == '@' {
				parts = name[:i]
				break
			}
		}
	}
	slashes := 0
	for _, c := range parts {
		if c == '/' {
			slashes++
		}
	}
	return slashes > 1
}
