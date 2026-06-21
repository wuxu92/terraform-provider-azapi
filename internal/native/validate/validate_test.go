package validate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/generated"
	_ "github.com/Azure/terraform-provider-azapi/internal/native/generated/all"
	"github.com/Azure/terraform-provider-azapi/internal/native/generator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestStorageAccountSchemaAgainstBicep(t *testing.T) {
	typesPath := filepath.Join("..", "..", "azure", "generated", "storage", "microsoft.storage", "2025-01-01", "types.json")

	data, err := os.ReadFile(typesPath)
	if err != nil {
		t.Skipf("types.json not found: %v", err)
	}

	defs, err := generator.ParseTypesJSON(data)
	if err != nil {
		t.Fatalf("ParseTypesJSON: %v", err)
	}

	generator.PostProcess(defs)

	var body *generator.Type
	for _, d := range defs {
		if d.Name == "Microsoft.Storage/storageAccounts@2025-01-01" {
			body = d.Body
			break
		}
	}
	if body == nil {
		t.Fatal("storage account body not found")
	}

	// Get the compiled schema
	s := generated.Registry["azapi_storage_account"].Schema()

	// Verify the azapin tag is present
	tag, ok := ExtractResourceTag(s.Description)
	if !ok {
		t.Error("no [azapin:...] tag in Description")
	} else {
		t.Logf("Tag: %s", tag)
	}

	// Validate; exclude the synthesized envelope attributes (name / parent / id).
	d := generated.Registry["azapi_storage_account"]
	mismatches := SchemaAgainstBicep(s, body, "name", d.ParentAttr, "id")

	errors := 0
	for _, m := range mismatches {
		if m.Kind != MismatchMissingInSchema {
			errors++
		}
	}

	if errors > 0 {
		t.Errorf("%d schema errors:\n%s", errors, FormatMismatches(mismatches))
	} else if len(mismatches) > 0 {
		t.Logf("Warnings:\n%s", FormatMismatches(mismatches))
	} else {
		t.Log("Compiled schema validated: 0 mismatches")
	}
}

func TestValidateDetectsMismatchesCompiled(t *testing.T) {
	body := &generator.Type{
		Kind: generator.KindObject,
		Properties: map[string]*generator.Property{
			"name":     {Name: "name", Type: &generator.Type{Kind: generator.KindString}, Flags: generator.FlagSystemManaged},
			"location": {Name: "location", Type: &generator.Type{Kind: generator.KindString}, Flags: generator.FlagRequired},
			"sku": {Name: "sku", Type: &generator.Type{Kind: generator.KindObject, Properties: map[string]*generator.Property{
				"name": {Name: "name", Type: &generator.Type{Kind: generator.KindString}, Flags: generator.FlagRequired},
				"tier": {Name: "tier", Type: &generator.Type{Kind: generator.KindString}},
			}}},
		},
	}

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"location": schema.StringAttribute{Required: true},
			"sku": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{Required: true},
				},
			},
			"phantom": schema.StringAttribute{Optional: true},
		},
	}

	mismatches := SchemaAgainstBicep(s, body)

	var extraCount, missingCount int
	for _, m := range mismatches {
		switch m.Kind {
		case MismatchExtraInSchema:
			extraCount++
		case MismatchMissingInSchema:
			missingCount++
		}
	}

	if extraCount != 1 {
		t.Errorf("expected 1 extra, got %d", extraCount)
	}
	if missingCount != 1 {
		t.Errorf("expected 1 missing, got %d", missingCount)
	}
	t.Logf("Synthetic test:\n%s", FormatMismatches(mismatches))
}

func TestExtractResourceTag(t *testing.T) {
	tests := []struct {
		desc string
		want string
		ok   bool
	}{
		{"Manages a resource. [azapin:Microsoft.Storage/storageAccounts@2025-01-01]", "Microsoft.Storage/storageAccounts@2025-01-01", true},
		{"No tag here", "", false},
		{"[azapin:Microsoft.KeyVault/vaults@2023-07-01]", "Microsoft.KeyVault/vaults@2023-07-01", true},
	}
	for _, tt := range tests {
		got, ok := ExtractResourceTag(tt.desc)
		if ok != tt.ok || got != tt.want {
			t.Errorf("ExtractResourceTag(%q) = (%q, %v), want (%q, %v)", tt.desc, got, ok, tt.want, tt.ok)
		}
	}
}
