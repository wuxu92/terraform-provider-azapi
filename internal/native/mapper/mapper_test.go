package mapper_test

import (
	"context"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/azure"
	"github.com/Azure/terraform-provider-azapi/internal/native/mapper"
	"github.com/Azure/terraform-provider-azapi/internal/native/services"
	_ "github.com/Azure/terraform-provider-azapi/internal/native/services/all"
	"github.com/Azure/terraform-provider-azapi/internal/native/typegraph"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// loadStorageBody parses the storage account bicep body type graph.
func loadStorageBody(t *testing.T) *typegraph.Type {
	t.Helper()
	return loadBody(t, "Microsoft.Storage/storageAccounts")
}

func TestRoundTripStorageAccount(t *testing.T) {
	ctx := context.Background()
	body := loadStorageBody(t)
	objType := services.Registry["azapi_storage_account"].Schema().Type().(basetypes.ObjectType)

	// A representative ARM GET response (subset of fields across types).
	arm := map[string]interface{}{
		"kind":     "StorageV2",
		"location": "westus2",
		"sku": map[string]interface{}{
			"name": "Standard_LRS",
			"tier": "Standard",
		},
		"tags": map[string]interface{}{"env": "prod", "team": "infra"}, // string-map, round-trips
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
	stateObj, diags := mapper.Flatten(ctx, arm, objType, body, nil)
	if diags.HasError() {
		t.Fatalf("Flatten diags: %v", diags)
	}
	if stateObj.IsNull() {
		t.Fatal("flattened state is null")
	}

	// Re-expand state → ARM JSON; verify the settable fields round-trip.
	out := mapper.Expand(stateObj, body)

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
	tags, ok := out["tags"].(map[string]interface{})
	if !ok || tags["env"] != "prod" || tags["team"] != "infra" {
		t.Errorf("tags round-trip: got %v", out["tags"])
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

func TestFlattenIntoPreservesEquivalentLocationBase(t *testing.T) {
	ctx := context.Background()
	body := loadStorageBody(t)
	objType := services.Registry["azapi_storage_account"].Schema().Type().(basetypes.ObjectType)

	planObj, diags := mapper.Flatten(ctx, map[string]interface{}{
		"kind":     "StorageV2",
		"location": "eastus2",
		"sku":      map[string]interface{}{"name": "Standard_LRS"},
	}, objType, body, nil)
	if diags.HasError() {
		t.Fatalf("Flatten plan diags: %v", diags)
	}

	stateObj, diags := mapper.FlattenInto(ctx, map[string]interface{}{
		"kind":     "StorageV2",
		"location": "East US 2",
		"sku":      map[string]interface{}{"name": "Standard_LRS"},
	}, planObj, body)
	if diags.HasError() {
		t.Fatalf("FlattenInto diags: %v", diags)
	}

	location, ok := stateObj.Attributes()["location"].(types.String)
	if !ok {
		t.Fatalf("location = %T, want types.String", stateObj.Attributes()["location"])
	}
	if got := location.ValueString(); got != "eastus2" {
		t.Fatalf("location = %q, want planned canonical value eastus2", got)
	}
}

func TestFlattenIntoOverwritesDifferentLocation(t *testing.T) {
	ctx := context.Background()
	body := loadStorageBody(t)
	objType := services.Registry["azapi_storage_account"].Schema().Type().(basetypes.ObjectType)

	baseObj, diags := mapper.Flatten(ctx, map[string]interface{}{
		"kind":     "StorageV2",
		"location": "westus2",
		"sku":      map[string]interface{}{"name": "Standard_LRS"},
	}, objType, body, nil)
	if diags.HasError() {
		t.Fatalf("Flatten base diags: %v", diags)
	}

	stateObj, diags := mapper.FlattenInto(ctx, map[string]interface{}{
		"kind":     "StorageV2",
		"location": "East US 2",
		"sku":      map[string]interface{}{"name": "Standard_LRS"},
	}, baseObj, body)
	if diags.HasError() {
		t.Fatalf("FlattenInto diags: %v", diags)
	}

	location := stateObj.Attributes()["location"].(types.String)
	if got := location.ValueString(); got != "East US 2" {
		t.Fatalf("location = %q, want response value East US 2", got)
	}
}

func TestExpandSkipsNullAndUnknown(t *testing.T) {
	ctx := context.Background()
	body := loadStorageBody(t)
	objType := services.Registry["azapi_storage_account"].Schema().Type().(basetypes.ObjectType)

	// Minimal ARM body — only required-ish fields.
	arm := map[string]interface{}{
		"kind": "StorageV2",
		"sku":  map[string]interface{}{"name": "Standard_LRS"},
	}
	stateObj, diags := mapper.Flatten(ctx, arm, objType, body, nil)
	if diags.HasError() {
		t.Fatalf("Flatten diags: %v", diags)
	}
	out := mapper.Expand(stateObj, body)

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

// TestRoundTripObjectMap exercises a map-of-object body field (ARM
// additionalProperties → an object schema), which the schema emits as a
// MapNestedAttribute and the mapper must round-trip key-for-key. Here it is
// identity.userAssignedIdentities, keyed by the identity's ARM resource ID.
func TestRoundTripObjectMap(t *testing.T) {
	ctx := context.Background()
	body := loadStorageBody(t)
	objType := services.Registry["azapi_storage_account"].Schema().Type().(basetypes.ObjectType)

	const uaiID = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai1"
	arm := map[string]interface{}{
		"kind": "StorageV2",
		"identity": map[string]interface{}{
			"type": "UserAssigned",
			"userAssignedIdentities": map[string]interface{}{
				uaiID: map[string]interface{}{
					"clientId":    "11111111-1111-1111-1111-111111111111",
					"principalId": "22222222-2222-2222-2222-222222222222",
				},
			},
		},
	}

	stateObj, diags := mapper.Flatten(ctx, arm, objType, body, nil)
	if diags.HasError() {
		t.Fatalf("Flatten diags: %v", diags)
	}
	out := mapper.Expand(stateObj, body)

	identity, ok := out["identity"].(map[string]interface{})
	if !ok {
		t.Fatalf("identity missing: %v", out["identity"])
	}
	uai, ok := identity["userAssignedIdentities"].(map[string]interface{})
	if !ok || len(uai) != 1 {
		t.Fatalf("userAssignedIdentities round-trip: got %v", identity["userAssignedIdentities"])
	}
	entry, ok := uai[uaiID].(map[string]interface{})
	if !ok {
		t.Fatalf("map key %q missing: got %v", uaiID, uai)
	}
	if entry["clientId"] != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("clientId round-trip: got %v", entry["clientId"])
	}
	if entry["principalId"] != "22222222-2222-2222-2222-222222222222" {
		t.Errorf("principalId round-trip: got %v", entry["principalId"])
	}
}

// userAssignedMap navigates a flattened state object down to its
// identity.user_assigned_identities map, asserting each hop's framework type.
func userAssignedMap(t *testing.T, stateObj types.Object) types.Map {
	t.Helper()
	identity, ok := stateObj.Attributes()["identity"].(types.Object)
	if !ok {
		t.Fatalf("identity = %T, want types.Object", stateObj.Attributes()["identity"])
	}
	uai, ok := identity.Attributes()["user_assigned_identities"].(types.Map)
	if !ok {
		t.Fatalf("user_assigned_identities = %T, want types.Map", identity.Attributes()["user_assigned_identities"])
	}
	return uai
}

// mapKeys returns a map's element keys for readable failure messages.
func mapKeys(m types.Map) []string {
	keys := make([]string, 0, len(m.Elements()))
	for k := range m.Elements() {
		keys = append(keys, k)
	}
	return keys
}

// TestFlattenIntoPreservesMapKeyCasing locks Fix 1: on read-refresh, flattenMapInto
// must keep the prior state's map-key casing for user_assigned_identities when the
// ARM response echoes a case-insensitively-equal key. Some RPs (Web) lowercase the
// resourceGroups segment of the UAI resource ID; Terraform map keys are
// case-sensitive, so without the fix the map churns as a perpetual drop+add.
func TestFlattenIntoPreservesMapKeyCasing(t *testing.T) {
	ctx := context.Background()
	body := loadStorageBody(t)
	objType := services.Registry["azapi_storage_account"].Schema().Type().(basetypes.ObjectType)

	const canonicalID = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai1"
	// Same identity, resource-group segment lowercased — the shape Web RP returns.
	const lowerID = "/subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai1"

	planObj, diags := mapper.Flatten(ctx, map[string]interface{}{
		"kind": "StorageV2",
		"identity": map[string]interface{}{
			"type": "UserAssigned",
			"userAssignedIdentities": map[string]interface{}{
				canonicalID: map[string]interface{}{
					"clientId":    "11111111-1111-1111-1111-111111111111",
					"principalId": "22222222-2222-2222-2222-222222222222",
				},
			},
		},
	}, objType, body, nil)
	if diags.HasError() {
		t.Fatalf("Flatten plan diags: %v", diags)
	}

	stateObj, diags := mapper.FlattenInto(ctx, map[string]interface{}{
		"kind": "StorageV2",
		"identity": map[string]interface{}{
			"type": "UserAssigned",
			"userAssignedIdentities": map[string]interface{}{
				lowerID: map[string]interface{}{
					"clientId":    "11111111-1111-1111-1111-111111111111",
					"principalId": "22222222-2222-2222-2222-222222222222",
				},
			},
		},
	}, planObj, body)
	if diags.HasError() {
		t.Fatalf("FlattenInto diags: %v", diags)
	}

	uai := userAssignedMap(t, stateObj)
	elems := uai.Elements()
	if len(elems) != 1 {
		t.Fatalf("user_assigned_identities has %d keys, want 1: %v", len(elems), mapKeys(uai))
	}
	if _, ok := elems[canonicalID]; !ok {
		t.Fatalf("map key casing not preserved: got %v, want canonical %q", mapKeys(uai), canonicalID)
	}
	if _, ok := elems[lowerID]; ok {
		t.Fatalf("lowercased key %q leaked into state; base casing must win", lowerID)
	}

	// The server-populated values are still carried through under the preserved key.
	entry, ok := elems[canonicalID].(types.Object)
	if !ok {
		t.Fatalf("element = %T, want types.Object", elems[canonicalID])
	}
	if cid := entry.Attributes()["client_id"].(types.String).ValueString(); cid != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("client_id under preserved key = %q, want carried-through value", cid)
	}
}

// TestFlattenIntoNewMapKeyCanonicalizesResourceID locks the complementary half of
// Fix 1: a key present in the ARM response but ABSENT from the prior state has no
// base casing to preserve, so a resource-ID key must land under its CANONICAL Azure
// casing (resourcegroups -> resourceGroups) rather than the response's raw casing.
// This keeps a freshly-returned UAI id stable against the case-canonical id config
// carries, instead of churning as a case-only drop+add.
func TestFlattenIntoNewMapKeyCanonicalizesResourceID(t *testing.T) {
	ctx := context.Background()
	body := loadStorageBody(t)
	objType := services.Registry["azapi_storage_account"].Schema().Type().(basetypes.ObjectType)

	const baseID = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai1"
	// A second identity the base never carried; ARM returns it with a lowercased
	// resourcegroups segment. It must be canonicalized to resourceGroups on the way in.
	const newLowerID = "/subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai2"
	const newCanonicalID = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai2"

	planObj, diags := mapper.Flatten(ctx, map[string]interface{}{
		"kind": "StorageV2",
		"identity": map[string]interface{}{
			"type": "UserAssigned",
			"userAssignedIdentities": map[string]interface{}{
				baseID: map[string]interface{}{
					"clientId":    "11111111-1111-1111-1111-111111111111",
					"principalId": "22222222-2222-2222-2222-222222222222",
				},
			},
		},
	}, objType, body, nil)
	if diags.HasError() {
		t.Fatalf("Flatten plan diags: %v", diags)
	}

	stateObj, diags := mapper.FlattenInto(ctx, map[string]interface{}{
		"kind": "StorageV2",
		"identity": map[string]interface{}{
			"type": "UserAssigned",
			"userAssignedIdentities": map[string]interface{}{
				baseID: map[string]interface{}{
					"clientId":    "11111111-1111-1111-1111-111111111111",
					"principalId": "22222222-2222-2222-2222-222222222222",
				},
				newLowerID: map[string]interface{}{
					"clientId":    "33333333-3333-3333-3333-333333333333",
					"principalId": "44444444-4444-4444-4444-444444444444",
				},
			},
		},
	}, planObj, body)
	if diags.HasError() {
		t.Fatalf("FlattenInto diags: %v", diags)
	}

	uai := userAssignedMap(t, stateObj)
	elems := uai.Elements()
	if len(elems) != 2 {
		t.Fatalf("user_assigned_identities has %d keys, want 2: %v", len(elems), mapKeys(uai))
	}
	if _, ok := elems[newCanonicalID]; !ok {
		t.Fatalf("new key not canonicalized: got %v, want canonical %q", mapKeys(uai), newCanonicalID)
	}
	if _, ok := elems[newLowerID]; ok {
		t.Fatalf("lowercased key %q leaked into state; new resource-ID keys must canonicalize", newLowerID)
	}
	if _, ok := elems[baseID]; !ok {
		t.Fatalf("base key dropped: got %v, want %q retained", mapKeys(uai), baseID)
	}

	// Server-populated values ride through under the canonical key.
	entry, ok := elems[newCanonicalID].(types.Object)
	if !ok {
		t.Fatalf("element = %T, want types.Object", elems[newCanonicalID])
	}
	if cid := entry.Attributes()["client_id"].(types.String).ValueString(); cid != "33333333-3333-3333-3333-333333333333" {
		t.Errorf("client_id under canonical key = %q, want carried-through value", cid)
	}
	if pid := entry.Attributes()["principal_id"].(types.String).ValueString(); pid != "44444444-4444-4444-4444-444444444444" {
		t.Errorf("principal_id under canonical key = %q, want carried-through value", pid)
	}
}

// TestFlattenCanonicalizesImportedResourceIDKey locks the no-base import path
// (base.go ImportState calls mapper.Flatten with NO prior state). A UAI resource-ID
// key that ARM echoes with a lowercased resourcegroups segment must land in state
// under its canonical resourceGroups casing, so imported state matches the id config
// carries. Without the fix the raw lowercased key lands verbatim and drifts.
func TestFlattenCanonicalizesImportedResourceIDKey(t *testing.T) {
	ctx := context.Background()
	body := loadStorageBody(t)
	objType := services.Registry["azapi_storage_account"].Schema().Type().(basetypes.ObjectType)

	const importedLowerID = "/subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai1"
	const importedCanonicalID = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai1"

	stateObj, diags := mapper.Flatten(ctx, map[string]interface{}{
		"kind": "StorageV2",
		"identity": map[string]interface{}{
			"type": "UserAssigned",
			"userAssignedIdentities": map[string]interface{}{
				importedLowerID: map[string]interface{}{
					"clientId":    "11111111-1111-1111-1111-111111111111",
					"principalId": "22222222-2222-2222-2222-222222222222",
				},
			},
		},
	}, objType, body, nil)
	if diags.HasError() {
		t.Fatalf("Flatten diags: %v", diags)
	}

	uai := userAssignedMap(t, stateObj)
	elems := uai.Elements()
	if len(elems) != 1 {
		t.Fatalf("user_assigned_identities has %d keys, want 1: %v", len(elems), mapKeys(uai))
	}
	if _, ok := elems[importedCanonicalID]; !ok {
		t.Fatalf("imported key not canonicalized: got %v, want canonical %q", mapKeys(uai), importedCanonicalID)
	}
	if _, ok := elems[importedLowerID]; ok {
		t.Fatalf("lowercased key %q leaked into imported state; import path must canonicalize", importedLowerID)
	}

	// Server values carried through under the canonical imported key.
	entry, ok := elems[importedCanonicalID].(types.Object)
	if !ok {
		t.Fatalf("element = %T, want types.Object", elems[importedCanonicalID])
	}
	if cid := entry.Attributes()["client_id"].(types.String).ValueString(); cid != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("client_id under canonical imported key = %q, want carried-through value", cid)
	}
}

// TestFlattenKeepsNonResourceIDMapKeyVerbatim guards the other side of
// canonicalMapKey through the same no-base Flatten MapType branch: an open-map key
// that is NOT an ARM resource ID (arm.ParseResourceID fails) must survive byte-for-
// byte. Storage `tags` is a string-valued open map; a mixed-case tag name like
// "CostCenter" must not be case-folded or otherwise rewritten.
func TestFlattenKeepsNonResourceIDMapKeyVerbatim(t *testing.T) {
	ctx := context.Background()
	body := loadStorageBody(t)
	objType := services.Registry["azapi_storage_account"].Schema().Type().(basetypes.ObjectType)

	const tagKey = "CostCenter"
	stateObj, diags := mapper.Flatten(ctx, map[string]interface{}{
		"kind": "StorageV2",
		"tags": map[string]interface{}{tagKey: "1234"},
	}, objType, body, nil)
	if diags.HasError() {
		t.Fatalf("Flatten diags: %v", diags)
	}

	tags, ok := stateObj.Attributes()["tags"].(types.Map)
	if !ok {
		t.Fatalf("tags = %T, want types.Map", stateObj.Attributes()["tags"])
	}
	elems := tags.Elements()
	if len(elems) != 1 {
		t.Fatalf("tags has %d keys, want 1: %v", len(elems), mapKeys(tags))
	}
	if _, ok := elems[tagKey]; !ok {
		t.Fatalf("non-resource-ID key not kept verbatim: got %v, want %q", mapKeys(tags), tagKey)
	}
	if v := elems[tagKey].(types.String).ValueString(); v != "1234" {
		t.Errorf("tag value under verbatim key = %q, want %q", v, "1234")
	}
}

// loadBody parses an arbitrary ARM resource's bicep body type graph for its
// latest stable API version, resolved through the azure schema loader.
func loadBody(t *testing.T, armType string) *typegraph.Type {
	t.Helper()
	version, err := azure.GetLatestStableApiVersion(armType)
	if err != nil {
		t.Skipf("no stable version for %s: %v", armType, err)
	}
	location, err := azure.GetResourceTypeLocation(armType, version)
	if err != nil {
		t.Skipf("no types.json for %s: %v", armType, err)
	}
	data, err := azure.StaticFiles.ReadFile("generated/" + location)
	if err != nil {
		t.Skipf("types.json not found: %v", err)
	}
	defs, err := typegraph.ParseTypesJSON(data)
	if err != nil {
		t.Fatalf("ParseTypesJSON: %v", err)
	}
	typegraph.
		PostProcess(defs)
	tag := armType + "@" + version
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
	objType := services.Registry["azapi_storage_account_blob_service"].Schema().Type().(basetypes.ObjectType)

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
	planObj, diags := mapper.Flatten(ctx, plan, objType, body, nil)
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
	out, diags := mapper.FlattenInto(ctx, resp, planObj, body)
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

func TestFlattenApplyIntoPreservesKnownWebSiteCompleteValues(t *testing.T) {
	ctx := context.Background()
	body := loadBody(t, "Microsoft.Web/sites")
	objType := services.Registry["azapi_web_site"].Schema().Type().(basetypes.ObjectType)
	serverFarmID := "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Web/serverfarms/plan"

	planObj, diags := mapper.Flatten(ctx, map[string]interface{}{
		"kind":     "app",
		"location": "eastus2",
		"properties": map[string]interface{}{
			"serverFarmId":          serverFarmID,
			"clientAffinityEnabled": false,
			"clientCertEnabled":     false,
			"clientCertMode":        "Required",
			"enabled":               true,
			"httpsOnly":             true,
			"publicNetworkAccess":   "Enabled",
			"siteConfig": map[string]interface{}{
				"ftpsState":        "Disabled",
				"http20Enabled":    false,
				"minTlsVersion":    "1.2",
				"scmMinTlsVersion": "1.2",
			},
		},
	}, objType, body, nil)
	if diags.HasError() {
		t.Fatalf("Flatten plan diags: %v", diags)
	}

	stateObj, diags := mapper.FlattenApplyInto(ctx, map[string]interface{}{
		"kind":     "app,linux",
		"location": "East US 2",
		"properties": map[string]interface{}{
			"serverFarmId":          serverFarmID + "-echo",
			"clientAffinityEnabled": true,
			"clientCertEnabled":     true,
			"clientCertMode":        "Optional",
			"enabled":               false,
			"httpsOnly":             false,
			"publicNetworkAccess":   "Disabled",
			"siteConfig": map[string]interface{}{
				"ftpsState":        "AllAllowed",
				"http20Enabled":    true,
				"minTlsVersion":    "1.3",
				"scmMinTlsVersion": "1.3",
			},
		},
	}, planObj, body)
	if diags.HasError() {
		t.Fatalf("FlattenApplyInto diags: %v", diags)
	}

	attrs := stateObj.Attributes()
	if got := attrs["kind"].(types.String).ValueString(); got != "app" {
		t.Fatalf("kind = %q, want planned value app", got)
	}
	props := attrs["properties"].(types.Object)
	propAttrs := props.Attributes()
	if got := propAttrs["server_farm_id"].(types.String).ValueString(); got != serverFarmID {
		t.Fatalf("server_farm_id = %q, want planned value", got)
	}
	if got := propAttrs["client_affinity_enabled"].(types.Bool).ValueBool(); got != false {
		t.Fatalf("client_affinity_enabled = %t, want planned false", got)
	}
	if got := propAttrs["client_cert_mode"].(types.String).ValueString(); got != "Required" {
		t.Fatalf("client_cert_mode = %q, want planned Required", got)
	}
	siteConfig := propAttrs["site_config"].(types.Object)
	if got := siteConfig.Attributes()["ftps_state"].(types.String).ValueString(); got != "Disabled" {
		t.Fatalf("site_config.ftps_state = %q, want planned Disabled", got)
	}
}

func TestFlattenIntoPreservesSensitiveWebSiteConfig(t *testing.T) {
	ctx := context.Background()
	body := loadBody(t, "Microsoft.Web/sites")
	objType := services.Registry["azapi_web_site"].Schema().Type().(basetypes.ObjectType)

	planObj, diags := mapper.Flatten(ctx, map[string]interface{}{
		"kind":     "app",
		"location": "eastus2",
		"properties": map[string]interface{}{
			"serverFarmId": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Web/serverfarms/plan",
		},
	}, objType, body, nil)
	if diags.HasError() {
		t.Fatalf("Flatten plan diags: %v", diags)
	}

	stateObj, diags := mapper.FlattenInto(ctx, map[string]interface{}{
		"kind":     "app",
		"location": "East US 2",
		"properties": map[string]interface{}{
			"serverFarmId": "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Web/serverfarms/plan",
			"siteConfig": map[string]interface{}{
				"ftpsState":     "AllAllowed",
				"http20Enabled": true,
			},
		},
	}, planObj, body)
	if diags.HasError() {
		t.Fatalf("FlattenInto diags: %v", diags)
	}

	props := stateObj.Attributes()["properties"].(types.Object)
	siteConfig := props.Attributes()["site_config"].(types.Object)
	if !siteConfig.IsNull() {
		t.Fatalf("site_config = %s, want preserved null because it is sensitive/write-only", siteConfig.String())
	}
}

// TestBlobServiceCorsSetRoundTrip guards the payload composer for CORS primitive
// arrays. The generated schema models header/method/origin arrays as SetAttribute
// because Azure may reorder them; Expand must still serialize every set element into
// the PUT body. A regression here produced corsRules with only maxAgeInSeconds.
func TestBlobServiceCorsSetRoundTrip(t *testing.T) {
	ctx := context.Background()
	body := loadBody(t, "Microsoft.Storage/storageAccounts/blobServices")
	objType := services.Registry["azapi_storage_account_blob_service"].Schema().Type().(basetypes.ObjectType)

	arm := map[string]interface{}{
		"properties": map[string]interface{}{
			"cors": map[string]interface{}{
				"corsRules": []interface{}{
					map[string]interface{}{
						"allowedHeaders":  []interface{}{"x-ms-*", "content-type"},
						"allowedMethods":  []interface{}{"GET", "PUT"},
						"allowedOrigins":  []interface{}{"https://example.com"},
						"exposedHeaders":  []interface{}{"x-ms-*"},
						"maxAgeInSeconds": float64(3600),
					},
				},
			},
		},
	}
	obj, diags := mapper.Flatten(ctx, arm, objType, body, nil)
	if diags.HasError() {
		t.Fatalf("Flatten diags: %v", diags)
	}

	props := obj.Attributes()["properties"].(types.Object)
	cors := props.Attributes()["cors"].(types.Object)
	rules := cors.Attributes()["cors_rules"].(types.List)
	rule := rules.Elements()[0].(types.Object)
	if _, ok := rule.Attributes()["allowed_headers"].(types.Set); !ok {
		t.Fatalf("allowed_headers = %T, want types.Set", rule.Attributes()["allowed_headers"])
	}

	out := mapper.Expand(obj, body)
	outProps := out["properties"].(map[string]interface{})
	outCors := outProps["cors"].(map[string]interface{})
	outRules := outCors["corsRules"].([]interface{})
	outRule := outRules[0].(map[string]interface{})
	assertStringElements(t, outRule["allowedHeaders"], "x-ms-*", "content-type")
	assertStringElements(t, outRule["allowedMethods"], "GET", "PUT")
	assertStringElements(t, outRule["allowedOrigins"], "https://example.com")
	assertStringElements(t, outRule["exposedHeaders"], "x-ms-*")
	if got := outRule["maxAgeInSeconds"]; got != int64(3600) {
		t.Errorf("maxAgeInSeconds = %v, want 3600", got)
	}
}

func assertStringElements(t *testing.T, got interface{}, want ...string) {
	t.Helper()
	items, ok := got.([]interface{})
	if !ok {
		t.Fatalf("got %T (%v), want []interface{}", got, got)
	}
	seen := map[string]bool{}
	for _, item := range items {
		s, ok := item.(string)
		if !ok {
			t.Fatalf("set item = %T (%v), want string", item, item)
		}
		seen[s] = true
	}
	if len(seen) != len(want) {
		t.Fatalf("elements = %v, want %v", items, want)
	}
	for _, w := range want {
		if !seen[w] {
			t.Fatalf("elements = %v, missing %q", items, w)
		}
	}
}
