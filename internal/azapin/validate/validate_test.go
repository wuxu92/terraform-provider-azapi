package validate

import (
	"path/filepath"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/azapin/generated"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestStorageAccountSchemaAgainstBicep(t *testing.T) {
	typesPath := filepath.Join("..", "..", "azure", "generated", "storage", "microsoft.storage", "2025-01-01", "types.json")

	bodies, err := ParseBicepTypes(typesPath)
	if err != nil {
		t.Skipf("types.json not found: %v", err)
	}

	body, ok := bodies["Microsoft.Storage/storageAccounts@2025-01-01"]
	if !ok {
		t.Fatal("storage account body not found in types.json")
	}

	// Get the compiled schema from the generated package
	s := generated.AzapiStorageAccountSchema()

	// Validate
	mismatches := SchemaAgainstBicep(s, body)

	errors := 0
	warnings := 0
	for _, m := range mismatches {
		switch m.Kind {
		case MismatchExtraInSchema, MismatchTypeMismatch:
			errors++
		case MismatchMissingInSchema:
			warnings++
		}
	}

	if errors > 0 || warnings > 0 {
		t.Logf("Validation results:\n%s", FormatMismatches(mismatches))
	}

	if errors > 0 {
		t.Errorf("%d schema errors (extra attributes or type mismatches)", errors)
	}

	if errors == 0 && warnings == 0 {
		t.Logf("Compiled schema validated against bicep types: 0 mismatches")
	}
}

func TestValidateDetectsMismatchesCompiled(t *testing.T) {
	body := &BicepType{
		Kind: BicepKindObject,
		Properties: map[string]*BicepProperty{
			"name":     {Name: "name", Type: &BicepType{Kind: BicepKindString}, Flags: 8}, // SystemManaged
			"location": {Name: "location", Type: &BicepType{Kind: BicepKindString}, Flags: 1},
			"sku": {Name: "sku", Type: &BicepType{Kind: BicepKindObject, Properties: map[string]*BicepProperty{
				"name": {Name: "name", Type: &BicepType{Kind: BicepKindString}, Flags: 1},
				"tier": {Name: "tier", Type: &BicepType{Kind: BicepKindString}},
			}}},
		},
	}

	// Schema missing "tier", has extra "phantom"
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
