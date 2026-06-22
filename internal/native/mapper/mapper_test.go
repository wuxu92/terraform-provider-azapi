package mapper

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/generated"
	_ "github.com/Azure/terraform-provider-azapi/internal/native/generated/all"
	"github.com/Azure/terraform-provider-azapi/internal/native/generator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// loadStorageBody parses the storage account bicep body type graph.
func loadStorageBody(t *testing.T) *generator.Type {
	t.Helper()
	indexPath := filepath.Join("..", "..", "azure", "generated", "index.json")
	idx, err := generator.LoadIndex(indexPath)
	if err != nil {
		t.Skipf("no index.json: %v", err)
	}
	tag, typesPath, err := idx.ResolveLatestStable("Microsoft.Storage/storageAccounts")
	if err != nil {
		t.Skipf("no stable storage version: %v", err)
	}
	data, err := os.ReadFile(typesPath)
	if err != nil {
		t.Skipf("types.json not found: %v", err)
	}
	defs, err := generator.ParseTypesJSON(data)
	if err != nil {
		t.Fatalf("ParseTypesJSON: %v", err)
	}
	generator.PostProcess(defs)
	for _, d := range defs {
		if d.Name == tag {
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

// loadBody parses an arbitrary ARM resource's bicep body type graph.
func loadBody(t *testing.T, armType string) *generator.Type {
	t.Helper()
	indexPath := filepath.Join("..", "..", "azure", "generated", "index.json")
	idx, err := generator.LoadIndex(indexPath)
	if err != nil {
		t.Skipf("no index.json: %v", err)
	}
	tag, typesPath, err := idx.ResolveLatestStable(armType)
	if err != nil {
		t.Skipf("no stable version for %s: %v", armType, err)
	}
	data, err := os.ReadFile(typesPath)
	if err != nil {
		t.Skipf("types.json not found: %v", err)
	}
	defs, err := generator.ParseTypesJSON(data)
	if err != nil {
		t.Fatalf("ParseTypesJSON: %v", err)
	}
	generator.PostProcess(defs)
	for _, d := range defs {
		if d.Name == tag {
			return d.Body
		}
	}
	t.Fatalf("body not found for %s", armType)
	return nil
}

// TestFlattenIntoPreservesNestedOmitted locks the fix for "inconsistent result
// after apply": a nested optional the user set but ARM drops from its GET response
// must keep the planned value (not collapse to null), at every depth. blobServices
// omits automaticSnapshotPolicyEnabled (deprecated) and, for containerDelete-
// RetentionPolicy, allowPermanentDelete when false — while echoing siblings.
func TestFlattenIntoPreservesNestedOmitted(t *testing.T) {
	ctx := context.Background()
	body := loadBody(t, "Microsoft.Storage/storageAccounts/blobServices")
	objType := generated.Registry["azapi_storage_account_blob_service"].Schema().Type().(basetypes.ObjectType)

	// Plan: user explicitly set both flags to false (plus an echoed sibling).
	plan := map[string]interface{}{
		"properties": map[string]interface{}{
			"automaticSnapshotPolicyEnabled": false,
			"containerDeleteRetentionPolicy": map[string]interface{}{
				"enabled":              true,
				"days":                 float64(7),
				"allowPermanentDelete": false,
			},
		},
	}
	planObj, diags := Flatten(ctx, plan, objType, body, nil)
	if diags.HasError() {
		t.Fatalf("Flatten plan diags: %v", diags)
	}

	// ARM GET response: drops the deprecated top-level flag and the nested
	// allowPermanentDelete, echoes the parent block, and changes days 7 -> 5.
	resp := map[string]interface{}{
		"properties": map[string]interface{}{
			"containerDeleteRetentionPolicy": map[string]interface{}{
				"enabled": true,
				"days":    float64(5),
			},
		},
	}
	out, diags := FlattenInto(ctx, resp, planObj, body)
	if diags.HasError() {
		t.Fatalf("FlattenInto diags: %v", diags)
	}

	props := out.Attributes()["properties"].(types.Object)

	// Top-level nested optional ARM omitted: planned false is preserved.
	snap := props.Attributes()["automatic_snapshot_policy_enabled"].(types.Bool)
	if snap.IsNull() || snap.ValueBool() != false {
		t.Errorf("automatic_snapshot_policy_enabled: got %v, want false (preserved)", snap)
	}

	cdrp := props.Attributes()["container_delete_retention_policy"].(types.Object)
	// Doubly-nested optional ARM omitted: planned false is preserved.
	apd := cdrp.Attributes()["allow_permanent_delete"].(types.Bool)
	if apd.IsNull() || apd.ValueBool() != false {
		t.Errorf("container_delete_retention_policy.allow_permanent_delete: got %v, want false (preserved)", apd)
	}
	// A field ARM echoed with a new value is still overwritten.
	days := cdrp.Attributes()["days"].(types.Int64)
	if days.IsNull() || days.ValueInt64() != 5 {
		t.Errorf("container_delete_retention_policy.days: got %v, want 5 (overwritten)", days)
	}
}
