package resource

import (
	"context"
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/generated"
	// Populate generated.Registry so New("azapi_storage_account") resolves.
	_ "github.com/Azure/terraform-provider-azapi/internal/native/generated/all"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestStorageAccountSchemaComposition(t *testing.T) {
	ctx := context.Background()
	r := New("azapi_storage_account")

	// Metadata: provider prefix + resource suffix.
	mdResp := &resource.MetadataResponse{}
	r.(resource.Resource).Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "azapi"}, mdResp)
	if mdResp.TypeName != "azapi_storage_account" {
		t.Errorf("TypeName = %q, want azapi_storage_account", mdResp.TypeName)
	}

	// Schema: composed (envelope + body), and framework-valid.
	schemaResp := &resource.SchemaResponse{}
	r.(resource.Resource).Schema(ctx, resource.SchemaRequest{}, schemaResp)
	s := schemaResp.Schema

	// Framework's own schema implementation validation must pass.
	if diags := s.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("schema validation failed: %v", diags)
	}

	// Envelope attributes present. Storage account is resource-group scoped, so
	// its parent reference is named resource_group_id (not the generic parent_id).
	for _, name := range []string{"name", "resource_group_id", "id"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing envelope attribute %q", name)
		}
	}
	if _, ok := s.Attributes["parent_id"]; ok {
		t.Error("storage account should expose resource_group_id, not the generic parent_id")
	}
	// timeouts block present.
	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("missing timeouts block")
	}
	// Body attributes present (from the generated schema).
	for _, name := range []string{"location", "kind", "sku", "properties"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing body attribute %q", name)
		}
	}
}

func TestBlobServiceSchemaComposition(t *testing.T) {
	ctx := context.Background()
	r := New("azapi_storage_account_blob_service")

	// Metadata: provider prefix + resource suffix.
	mdResp := &resource.MetadataResponse{}
	r.(resource.Resource).Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "azapi"}, mdResp)
	if mdResp.TypeName != "azapi_storage_account_blob_service" {
		t.Errorf("TypeName = %q, want azapi_storage_account_blob_service", mdResp.TypeName)
	}

	// Schema: composed (envelope + body), and framework-valid.
	schemaResp := &resource.SchemaResponse{}
	r.(resource.Resource).Schema(ctx, resource.SchemaRequest{}, schemaResp)
	s := schemaResp.Schema
	if diags := s.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("schema validation failed: %v", diags)
	}

	// blobServices is a child resource: its parent reference is named after the
	// parent ARM type (storage_account_id), not the resource-group scope name or
	// the generic parent_id.
	for _, name := range []string{"name", "storage_account_id", "id"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing envelope attribute %q", name)
		}
	}
	for _, absent := range []string{"resource_group_id", "parent_id"} {
		if _, ok := s.Attributes[absent]; ok {
			t.Errorf("blob service should expose storage_account_id, not %q", absent)
		}
	}
	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("missing timeouts block")
	}
	// Body attribute present (from the generated schema).
	if _, ok := s.Attributes["properties"]; !ok {
		t.Error("missing body attribute \"properties\"")
	}
}

func TestResourceGroupSchemaComposition(t *testing.T) {
	ctx := context.Background()
	r := New("azapi_resource_group")

	// Metadata: provider prefix + resource suffix.
	mdResp := &resource.MetadataResponse{}
	r.(resource.Resource).Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "azapi"}, mdResp)
	if mdResp.TypeName != "azapi_resource_group" {
		t.Errorf("TypeName = %q, want azapi_resource_group", mdResp.TypeName)
	}

	// Schema: composed (envelope + body), and framework-valid.
	schemaResp := &resource.SchemaResponse{}
	r.(resource.Resource).Schema(ctx, resource.SchemaRequest{}, schemaResp)
	s := schemaResp.Schema
	if diags := s.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("schema validation failed: %v", diags)
	}

	// resourceGroups is a subscription-scoped top-level resource: its parent
	// reference is named subscription_id, not resource_group_id or parent_id.
	for _, name := range []string{"name", "subscription_id", "id"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing envelope attribute %q", name)
		}
	}
	for _, absent := range []string{"resource_group_id", "parent_id"} {
		if _, ok := s.Attributes[absent]; ok {
			t.Errorf("resource group should expose subscription_id, not %q", absent)
		}
	}
	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("missing timeouts block")
	}
	// Body attributes present (from the generated schema).
	for _, name := range []string{"location", "managed_by", "properties"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing body attribute %q", name)
		}
	}
}

func TestInterfaceAssertions(t *testing.T) {
	r := New("azapi_storage_account")
	if _, ok := r.(resource.ResourceWithConfigure); !ok {
		t.Error("does not implement ResourceWithConfigure")
	}
	if _, ok := r.(resource.ResourceWithModifyPlan); !ok {
		t.Error("does not implement ResourceWithModifyPlan")
	}
	if _, ok := r.(resource.ResourceWithImportState); !ok {
		t.Error("does not implement ResourceWithImportState")
	}
	if _, ok := r.(resource.ResourceWithValidateConfig); !ok {
		t.Error("does not implement ResourceWithValidateConfig")
	}
}

func TestStorageAccountHookRegistered(t *testing.T) {
	if _, ok := hookRegistry["azapi_storage_account"]; !ok {
		t.Error("storage account hooks not registered")
	}
	if hookRegistry["azapi_storage_account"].ModifyPlan == nil {
		t.Error("storage account ModifyPlan hook is nil")
	}
}

func TestLoadStorageBody(t *testing.T) {
	d := generated.Registry["azapi_storage_account"]
	body, err := loadBody(d.ARMType, d.APIVersion)
	if err != nil {
		t.Fatalf("loadBody: %v", err)
	}
	if body == nil || body.Properties["sku"] == nil {
		t.Fatal("storage body missing sku")
	}
	// Cached on second call.
	body2, _ := loadBody(d.ARMType, d.APIVersion)
	if body2 != body {
		t.Error("loadBody not cached")
	}
}
