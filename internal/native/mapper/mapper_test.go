package mapper_test

import (
	"context"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/azure"
	"github.com/Azure/terraform-provider-azapi/internal/native/generated"
	_ "github.com/Azure/terraform-provider-azapi/internal/native/generated/all"
	"github.com/Azure/terraform-provider-azapi/internal/native/generator"
	"github.com/Azure/terraform-provider-azapi/internal/native/mapper"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// loadStorageBody parses the storage account bicep body type graph.
func loadStorageBody(t *testing.T) *generator.Type {
	t.Helper()
	return loadBody(t, "Microsoft.Storage/storageAccounts")
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
	objType := generated.Registry["azapi_storage_account"].Schema().Type().(basetypes.ObjectType)

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
	objType := generated.Registry["azapi_storage_account"].Schema().Type().(basetypes.ObjectType)

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
	objType := generated.Registry["azapi_storage_account"].Schema().Type().(basetypes.ObjectType)

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
	objType := generated.Registry["azapi_storage_account"].Schema().Type().(basetypes.ObjectType)

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

// loadBody parses an arbitrary ARM resource's bicep body type graph for its
// latest stable API version, resolved through the azure schema loader.
func loadBody(t *testing.T, armType string) *generator.Type {
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
	defs, err := generator.ParseTypesJSON(data)
	if err != nil {
		t.Fatalf("ParseTypesJSON: %v", err)
	}
	generator.PostProcess(defs)
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
	objType := generated.Registry["azapi_web_site"].Schema().Type().(basetypes.ObjectType)
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
	objType := generated.Registry["azapi_web_site"].Schema().Type().(basetypes.ObjectType)

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
	objType := generated.Registry["azapi_storage_account_blob_service"].Schema().Type().(basetypes.ObjectType)

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
