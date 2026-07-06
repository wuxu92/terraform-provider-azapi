package generator

import (
	"strings"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
)

// Generation integration tests exercising the parse -> emit pipeline on real and
// hand-built type graphs. They live in package generator because they call
// EmitSchema; pure parse/navigation tests stay in package typegraph.

func TestParseStorageAccount(t *testing.T) {
	defs, ver := latestStorageDefs(t)
	tag := "Microsoft.Storage/storageAccounts@" + ver

	if len(defs) == 0 {
		t.Fatal("no resource definitions found")
	}

	// Find storage account
	var sa *typegraph.ResourceDefinition
	for _, d := range defs {
		if d.Name == tag {
			sa = d
			break
		}
	}
	if sa == nil {
		t.Fatalf("%s not found", tag)
	}

	t.Logf("Found %d resource definitions", len(defs))
	t.Logf("Storage account: %s (API version: %s)", sa.Name, sa.APIVersion)

	// Check body is an ObjectType with properties
	if sa.Body == nil {
		t.Fatal("body is nil")
	}
	if sa.Body.Kind != typegraph.KindObject {
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
	if propsProp.Type.Kind != typegraph.KindObject {
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
	if sku.Type.Kind != typegraph.KindObject {
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

// TestParseSelfReferentialType ensures the walker breaks reference cycles so
// that EmitSchema terminates instead of stack-overflowing. The ErrorDetail
// pattern (an object with a `details: ErrorDetail[]` self-reference) is the
// canonical recursive ARM type.
func TestParseSelfReferentialType(t *testing.T) {
	// Hand-built types.json:
	//   #0 string
	//   #1 ArrayType(itemType -> #2 ErrorDetail)   -- forward ref
	//   #2 ErrorDetail { code: string, details: #1 }  -- back-edge via #1
	//   #3 props { error: #2 }
	//   #4 ResourceType body -> #3
	raw := `[
		{"$type":"StringType"},
		{"$type":"ArrayType","itemType":{"$ref":"#/2"}},
		{"$type":"ObjectType","name":"ErrorDetail","properties":{
			"code":{"type":{"$ref":"#/0"},"flags":0},
			"details":{"type":{"$ref":"#/1"},"flags":0}
		}},
		{"$type":"ObjectType","name":"props","properties":{
			"error":{"type":{"$ref":"#/2"},"flags":0}
		}},
		{"$type":"ResourceType","name":"Microsoft.Test/things@2024-01-01","body":{"$ref":"#/3"}}
	]`

	defs, err := typegraph.ParseTypesJSON([]byte(raw))
	if err != nil {
		t.Fatalf("ParseTypesJSON: %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("expected 1 resource def, got %d", len(defs))
	}
	typegraph.PostProcess(defs)

	// EmitSchema must terminate (no stack overflow). A bounded call is the test.
	src, err := EmitSchema(defs[0])
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}
	if len(src) == 0 {
		t.Fatal("empty schema source")
	}
	// The cycle is broken: `error` is a nested object containing `details`,
	// whose recursive element degraded to a non-object (so a flat list), and
	// emission terminated. Both fields must be present.
	if !strings.Contains(src, `"error": schema.SingleNestedAttribute`) {
		t.Error("expected `error` to be a SingleNestedAttribute")
	}
	if !strings.Contains(src, `"details"`) {
		t.Error("expected `details` field to be present")
	}
}

// TestParseDiscriminatedObjectType ensures a DiscriminatedObjectType (whose
// `elements` is a JSON object, not an array) does not fail the whole document's
// unmarshal. Previously this errored: "cannot unmarshal object into Go struct
// field rawEntry.elements of type []generator.rawRef".
func TestParseDiscriminatedObjectType(t *testing.T) {
	raw := `[
		{"$type":"StringType"},
		{"$type":"ObjectType","name":"VariantA","properties":{"a":{"type":{"$ref":"#/0"},"flags":0}}},
		{"$type":"DiscriminatedObjectType","name":"Poly","discriminator":"kind","baseProperties":{
			"kind":{"type":{"$ref":"#/0"},"flags":1}
		},"elements":{"A":{"$ref":"#/1"}}},
		{"$type":"ObjectType","name":"props","properties":{
			"variant":{"type":{"$ref":"#/2"},"flags":0}
		}},
		{"$type":"ResourceType","name":"Microsoft.Test/poly@2024-01-01","body":{"$ref":"#/3"}}
	]`

	defs, err := typegraph.ParseTypesJSON([]byte(raw))
	if err != nil {
		t.Fatalf("ParseTypesJSON failed on DiscriminatedObjectType: %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("expected 1 resource def, got %d", len(defs))
	}
	// The discriminated `variant` property degrades to KindAny → DynamicAttribute.
	src, err := EmitSchema(defs[0])
	if err != nil {
		t.Fatalf("EmitSchema: %v", err)
	}
	if !strings.Contains(src, `"variant": schema.DynamicAttribute`) {
		t.Error("expected discriminated property to emit a DynamicAttribute")
	}
}
