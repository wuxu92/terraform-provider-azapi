package generator

import (
	"os"
	"testing"
)

func TestParseStorageAccount(t *testing.T) {
	data, err := os.ReadFile("../../azure/generated/storage/microsoft.storage/2025-01-01/types.json")
	if err != nil {
		t.Skipf("types.json not found: %v", err)
	}

	defs, err := ParseTypesJSON(data)
	if err != nil {
		t.Fatalf("ParseTypesJSON: %v", err)
	}

	if len(defs) == 0 {
		t.Fatal("no resource definitions found")
	}

	// Find storage account
	var sa *ResourceDefinition
	for _, d := range defs {
		if d.Name == "Microsoft.Storage/storageAccounts@2025-01-01" {
			sa = d
			break
		}
	}
	if sa == nil {
		t.Fatal("Microsoft.Storage/storageAccounts@2025-01-01 not found")
	}

	t.Logf("Found %d resource definitions", len(defs))
	t.Logf("Storage account: %s (API version: %s)", sa.Name, sa.APIVersion)

	// Check body is an ObjectType with properties
	if sa.Body == nil {
		t.Fatal("body is nil")
	}
	if sa.Body.Kind != KindObject {
		t.Fatalf("body kind = %d, want KindObject", sa.Body.Kind)
	}

	// Check top-level properties exist
	requiredProps := []string{"id", "name", "type", "apiVersion", "sku", "kind", "location", "properties"}
	for _, prop := range requiredProps {
		if _, ok := sa.Body.Properties[prop]; !ok {
			t.Errorf("missing top-level property: %s", prop)
		}
	}

	// Check 'properties' is an ObjectType with sub-properties
	propsProp := sa.Body.Properties["properties"]
	if propsProp.Type.Kind != KindObject {
		t.Fatalf("properties.type.kind = %d, want KindObject", propsProp.Type.Kind)
	}

	propsType := propsProp.Type
	t.Logf("properties sub-object: %d properties", len(propsType.Properties))

	// Check some known properties exist
	knownProps := []string{"accessTier", "minimumTlsVersion", "supportsHttpsTrafficOnly", "isHnsEnabled", "allowBlobPublicAccess"}
	for _, prop := range knownProps {
		if _, ok := propsType.Properties[prop]; !ok {
			t.Errorf("missing property: properties.%s", prop)
		}
	}

	// Check accessTier is a union (enum)
	accessTier := propsType.Properties["accessTier"]
	if accessTier == nil {
		t.Fatal("accessTier is nil")
	}
	if !accessTier.Type.IsEnum() {
		t.Errorf("accessTier is not an enum (kind=%d, elements=%d)", accessTier.Type.Kind, len(accessTier.Type.Elements))
	} else {
		values := accessTier.Type.EnumValues()
		t.Logf("accessTier enum values: %v", values)
		if len(values) < 3 {
			t.Errorf("accessTier has %d enum values, want >= 3", len(values))
		}
	}

	// Check sku is an ObjectType with 'name' property
	sku := sa.Body.Properties["sku"]
	if sku.Type.Kind != KindObject {
		t.Fatalf("sku kind = %d, want KindObject", sku.Type.Kind)
	}
	if _, ok := sku.Type.Properties["name"]; !ok {
		t.Error("sku missing 'name' property")
	}

	// Check flags
	idProp := sa.Body.Properties["id"]
	if !idProp.Flags.IsReadOnly() {
		t.Error("id should be ReadOnly")
	}
	if !idProp.Flags.IsSystemManaged() {
		t.Error("id should be SystemManaged")
	}
	skuProp := sa.Body.Properties["sku"]
	if !skuProp.Flags.IsRequired() {
		t.Error("sku should be Required")
	}
}

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
