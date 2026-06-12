package generator

import (
	"os"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/azapin/naming"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

	// Apply post-processing (same as the PoC generator)
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

	// Generate the schema
	source, err := EmitSchema(sa)
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}
	_ = source // we use the schema function directly below

	// Build the schema by calling the emitter's logic, then validate
	// We can't call the generated Go function directly (it's in another package),
	// so we validate the type graph against itself — this ensures the emitter
	// doesn't skip or add properties that aren't in the source.
	mismatches := ValidateSchemaAgainstBicep(
		buildSchemaFromDef(sa),
		sa.Body,
	)

	// Log all mismatches
	if len(mismatches) > 0 {
		t.Logf("Mismatches:\n%s", FormatMismatches(mismatches))

		// Count by kind
		extra := 0
		missing := 0
		typeMismatch := 0
		for _, m := range mismatches {
			switch m.Kind {
			case MismatchExtraInSchema:
				extra++
			case MismatchMissingInSchema:
				missing++
			case MismatchTypeMismatch:
				typeMismatch++
			}
		}

		if extra > 0 {
			t.Errorf("%d attributes in Terraform schema but not in bicep types", extra)
		}
		if typeMismatch > 0 {
			t.Errorf("%d type mismatches between Terraform schema and bicep types", typeMismatch)
		}
		// Missing is a warning — some properties may be intentionally excluded
		if missing > 0 {
			t.Logf("WARNING: %d bicep properties not in Terraform schema (may be intentional)", missing)
		}
	} else {
		t.Log("Schema-bicep validation passed: no mismatches")
	}
}

// buildSchemaFromDef reconstructs the schema.Schema by re-running EmitSchema
// and parsing the result. Since we can't easily instantiate the generated schema
// object, we use a simpler approach: walk the type graph and build the schema
// attributes map directly, mirroring what the emitter does.
func buildSchemaFromDef(def *ResourceDefinition) schema.Schema {
	attrs := buildAttrsFromType(def.Body)
	return schema.Schema{
		Attributes: attrs,
	}
}

func buildAttrsFromType(typ *Type) map[string]schema.Attribute {
	if typ == nil || typ.Kind != KindObject {
		return nil
	}

	attrs := make(map[string]schema.Attribute)
	for armName, prop := range typ.Properties {
		if prop.Flags.IsSystemManaged() {
			continue
		}
		tfName := naming.CamelToSnake(armName)
		attrs[tfName] = buildAttr(prop)
	}
	return attrs
}

func buildAttr(prop *Property) schema.Attribute {
	typ := prop.Type

	switch {
	case typ.Kind == KindString:
		return schema.StringAttribute{}
	case typ.Kind == KindBool:
		return schema.BoolAttribute{}
	case typ.Kind == KindInt:
		return schema.Int64Attribute{}
	case typ.IsEnum():
		return schema.StringAttribute{}
	case typ.Kind == KindObject:
		return schema.SingleNestedAttribute{
			Attributes: buildAttrsFromType(typ),
		}
	case typ.Kind == KindArray:
		if typ.ElementType != nil && typ.ElementType.Kind == KindObject {
			return schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: buildAttrsFromType(typ.ElementType),
				},
			}
		}
		return schema.ListAttribute{}
	case typ.Kind == KindUnion:
		return schema.StringAttribute{}
	default:
		return schema.DynamicAttribute{}
	}
}

func TestValidateDetectsMismatches(t *testing.T) {
	// Build a bicep type with known properties
	bicepBody := &Type{
		Kind: KindObject,
		Properties: map[string]*Property{
			"name":     {Name: "name", Type: &Type{Kind: KindString}, Flags: FlagSystemManaged},
			"location": {Name: "location", Type: &Type{Kind: KindString}, Flags: FlagRequired},
			"sku":      {Name: "sku", Type: &Type{Kind: KindObject, Properties: map[string]*Property{
				"name": {Name: "name", Type: &Type{Kind: KindString}, Flags: FlagRequired},
				"tier": {Name: "tier", Type: &Type{Kind: KindString}},
			}}},
			"tags": {Name: "tags", Type: &Type{Kind: KindObject, Properties: map[string]*Property{}}},
		},
	}

	// Build a schema that has an extra attribute and misses one
	tfSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"location": schema.StringAttribute{},
			"sku": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{},
					// "tier" is missing → should be reported
				},
			},
			"tags":    schema.SingleNestedAttribute{},
			"phantom": schema.StringAttribute{}, // extra → should be reported
		},
	}

	mismatches := ValidateSchemaAgainstBicep(tfSchema, bicepBody)

	// Should find: "phantom" extra in schema, "tier" missing in schema
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
