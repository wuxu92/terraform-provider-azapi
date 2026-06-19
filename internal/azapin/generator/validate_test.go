package generator

import (
	"os"
	"testing"
)

func TestValidateStorageAccountSchema(t *testing.T) {
	data, err := os.ReadFile("../../azure/generated/storage/microsoft.storage/2025-01-01/types.json")
	if err != nil {
		t.Skipf("types.json not found: %v", err)
	}

	defs, err := ParseTypesJSON(data)
	if err != nil {
		t.Fatalf("ParseTypesJSON: %v", err)
	}

	PostProcess(defs)

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

	// Generate the schema source
	source, err := EmitSchema(sa)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}

	// Validate: emitted source covers all bicep properties and vice versa. The
	// synthesized envelope attributes (name / parent reference / id) are excluded
	// since they are not part of the bicep body graph.
	mismatches := ValidateEmittedSchema(source, sa.Body, EnvelopeAttrNames(sa)...)

	if len(mismatches) > 0 {
		t.Logf("Mismatches:\n%s", FormatMismatches(mismatches))

		errors := 0
		for _, m := range mismatches {
			if m.Kind != MismatchMissingInSchema {
				errors++
			}
		}
		if errors > 0 {
			t.Errorf("%d schema errors (extra or type mismatch)", errors)
		}
	} else {
		// Count validated paths
		paths := make(map[string]*Property)
		CollectExpectedPaths(sa.Body, "", paths)
		t.Logf("Schema-bicep validation passed: %d properties validated, 0 mismatches", len(paths))
	}
}

func TestValidateDetectsMismatches(t *testing.T) {
	// Build a bicep type with known properties
	bicepBody := &Type{
		Kind: KindObject,
		Properties: map[string]*Property{
			"name":     {Name: "name", Type: &Type{Kind: KindString}, Flags: FlagSystemManaged},
			"location": {Name: "location", Type: &Type{Kind: KindString}, Flags: FlagRequired},
			"sku": {Name: "sku", Type: &Type{Kind: KindObject, Properties: map[string]*Property{
				"name": {Name: "name", Type: &Type{Kind: KindString}, Flags: FlagRequired},
				"tier": {Name: "tier", Type: &Type{Kind: KindString}},
			}}},
		},
	}

	// Generate a schema source that has an extra attribute and misses "tier"
	fakeSource := `
		"location": schema.StringAttribute{
			Required: true,
		},
		"sku": schema.SingleNestedAttribute{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{
					Required: true,
				},
			},
		},
		"phantom": schema.StringAttribute{
			Optional: true,
		},
	`

	mismatches := ValidateEmittedSchema(fakeSource, bicepBody)

	var extraCount, missingCount int
	for _, m := range mismatches {
		switch m.Kind {
		case MismatchExtraInSchema:
			extraCount++
			if m.Path != "phantom" {
				t.Errorf("unexpected extra: %s", m.Path)
			}
		case MismatchMissingInSchema:
			missingCount++
			if m.Path != "sku.tier" {
				t.Errorf("unexpected missing: %s", m.Path)
			}
		}
	}

	if extraCount != 1 {
		t.Errorf("expected 1 extra in schema, got %d", extraCount)
	}
	if missingCount != 1 {
		t.Errorf("expected 1 missing in schema, got %d", missingCount)
	}

	t.Logf("Mismatches:\n%s", FormatMismatches(mismatches))
}
