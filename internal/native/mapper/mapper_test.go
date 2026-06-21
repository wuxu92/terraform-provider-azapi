package mapper

import (
	"context"
	"os"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/generated"
	_ "github.com/Azure/terraform-provider-azapi/internal/native/generated/all"
	"github.com/Azure/terraform-provider-azapi/internal/native/generator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// loadStorageBody parses the storage account bicep body type graph.
func loadStorageBody(t *testing.T) *generator.Type {
	t.Helper()
	data, err := os.ReadFile("../../azure/generated/storage/microsoft.storage/2025-01-01/types.json")
	if err != nil {
		t.Skipf("types.json not found: %v", err)
	}
	defs, err := generator.ParseTypesJSON(data)
	if err != nil {
		t.Fatalf("ParseTypesJSON: %v", err)
	}
	generator.PostProcess(defs)
	for _, d := range defs {
		if d.Name == "Microsoft.Storage/storageAccounts@2025-01-01" {
			return d.Body
		}
	}
	t.Fatal("storage account body not found")
	return nil
}

func TestRoundTripStorageAccount(t *testing.T) {
	ctx := context.Background()
	body := loadStorageBody(t)
	objType := generated.Registry["azapi_storage_account"].Schema().Type().(basetypes.ObjectType)

	// A representative ARM GET response (subset of fields across types).
	arm := map[string]interface{}{
		"kind":     "StorageV2",
		"location": "westus2",
		"sku": map[string]interface{}{
			"name": "Standard_LRS",
			"tier": "Standard",
		},
		"tags": map[string]interface{}{}, // additionalProperties — not modeled, ignored
		"properties": map[string]interface{}{
			"accessTier":               "Hot",
			"minimumTlsVersion":        "TLS1_2",
			"supportsHttpsTrafficOnly": true,
			"isHnsEnabled":             false,
			"provisioningState":        "Succeeded", // computed
			"networkAcls": map[string]interface{}{
				"defaultAction": "Allow",
				"ipRules": []interface{}{
					map[string]interface{}{"value": "1.2.3.4", "action": "Allow"},
				},
			},
		},
	}

	// Flatten ARM → state object.
	envelope := map[string]interface{}{} // no envelope override in this test
	_ = envelope
	stateObj, diags := Flatten(ctx, arm, objType, body, nil)
	if diags.HasError() {
		t.Fatalf("Flatten diags: %v", diags)
	}
	if stateObj.IsNull() {
		t.Fatal("flattened state is null")
	}

	// Re-expand state → ARM JSON; verify the settable fields round-trip.
	out := Expand(stateObj, body)

	if out["kind"] != "StorageV2" {
		t.Errorf("kind round-trip: got %v", out["kind"])
	}
	if out["location"] != "westus2" {
		t.Errorf("location round-trip: got %v", out["location"])
	}
	sku, ok := out["sku"].(map[string]interface{})
	if !ok || sku["name"] != "Standard_LRS" {
		t.Errorf("sku round-trip: got %v", out["sku"])
	}
	props, ok := out["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("properties missing: %v", out["properties"])
	}
	if props["accessTier"] != "Hot" {
		t.Errorf("accessTier round-trip: got %v", props["accessTier"])
	}
	if props["minimumTlsVersion"] != "TLS1_2" {
		t.Errorf("minimumTlsVersion round-trip: got %v", props["minimumTlsVersion"])
	}
	if props["supportsHttpsTrafficOnly"] != true {
		t.Errorf("supportsHttpsTrafficOnly round-trip: got %v", props["supportsHttpsTrafficOnly"])
	}
	// Computed field is read back into state and re-expanded (the resource layer
	// decides whether to send computed fields; the mapper round-trips faithfully).
	if props["provisioningState"] != "Succeeded" {
		t.Errorf("provisioningState round-trip: got %v", props["provisioningState"])
	}
	// Nested list-of-objects.
	nacls, ok := props["networkAcls"].(map[string]interface{})
	if !ok {
		t.Fatalf("networkAcls missing: %v", props["networkAcls"])
	}
	if nacls["defaultAction"] != "Allow" {
		t.Errorf("networkAcls.defaultAction: got %v", nacls["defaultAction"])
	}
	iprules, ok := nacls["ipRules"].([]interface{})
	if !ok || len(iprules) != 1 {
		t.Fatalf("ipRules: got %v", nacls["ipRules"])
	}
	rule := iprules[0].(map[string]interface{})
	if rule["value"] != "1.2.3.4" {
		t.Errorf("ipRules[0].value: got %v", rule["value"])
	}
}

func TestExpandSkipsNullAndUnknown(t *testing.T) {
	ctx := context.Background()
	body := loadStorageBody(t)
	objType := generated.Registry["azapi_storage_account"].Schema().Type().(basetypes.ObjectType)

	// Minimal ARM body — only required-ish fields.
	arm := map[string]interface{}{
		"kind": "StorageV2",
		"sku":  map[string]interface{}{"name": "Standard_LRS"},
	}
	stateObj, diags := Flatten(ctx, arm, objType, body, nil)
	if diags.HasError() {
		t.Fatalf("Flatten diags: %v", diags)
	}
	out := Expand(stateObj, body)

	// Absent fields must NOT appear in the expanded body (they were null).
	if _, ok := out["location"]; ok {
		t.Error("location should be absent (was null)")
	}
	if _, ok := out["properties"]; ok {
		t.Error("properties should be absent (was null)")
	}
	if out["kind"] != "StorageV2" {
		t.Errorf("kind: got %v", out["kind"])
	}
}
