package resource_test

import (
	"context"
	"testing"

	"reflect"

	// Populate services.Registry and hookRegistry so New("azapi_storage_account") resolves.
	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
	_ "github.com/Azure/terraform-provider-azapi/internal/native/services/all"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func hookField(t *testing.T, r resource.Resource, name string) reflect.Value {
	t.Helper()
	hooks := reflect.ValueOf(r).Elem().FieldByName("hooks")
	if !hooks.IsValid() || hooks.IsNil() {
		t.Fatalf("%T has no hooks", r)
	}
	field := hooks.Elem().FieldByName(name)
	if !field.IsValid() {
		t.Fatalf("hook field %q not found", name)
	}
	return field
}

func TestStorageAccountSchemaComposition(t *testing.T) {
	ctx := context.Background()
	r := nativeresource.New("azapi_storage_account")

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
	r := nativeresource.New("azapi_storage_account_blob_service")

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
	r := nativeresource.New("azapi_resource_group")

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

func TestWebServerFarmSchemaComposition(t *testing.T) {
	ctx := context.Background()
	r := nativeresource.New("azapi_web_server_farm")

	mdResp := &resource.MetadataResponse{}
	r.(resource.Resource).Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "azapi"}, mdResp)
	if mdResp.TypeName != "azapi_web_server_farm" {
		t.Errorf("TypeName = %q, want azapi_web_server_farm", mdResp.TypeName)
	}

	schemaResp := &resource.SchemaResponse{}
	r.(resource.Resource).Schema(ctx, resource.SchemaRequest{}, schemaResp)
	s := schemaResp.Schema
	if diags := s.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("schema validation failed: %v", diags)
	}

	for _, name := range []string{"name", "resource_group_id", "id"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing envelope attribute %q", name)
		}
	}
	if _, ok := s.Attributes["parent_id"]; ok {
		t.Error("web server farm should expose resource_group_id, not the generic parent_id")
	}
	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("missing timeouts block")
	}
	for _, name := range []string{"location", "kind", "sku", "properties", "identity", "tags"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing body attribute %q", name)
		}
	}
}

func TestWebSiteSchemaComposition(t *testing.T) {
	ctx := context.Background()
	r := nativeresource.New("azapi_web_site")

	mdResp := &resource.MetadataResponse{}
	r.(resource.Resource).Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "azapi"}, mdResp)
	if mdResp.TypeName != "azapi_web_site" {
		t.Errorf("TypeName = %q, want azapi_web_site", mdResp.TypeName)
	}

	schemaResp := &resource.SchemaResponse{}
	r.(resource.Resource).Schema(ctx, resource.SchemaRequest{}, schemaResp)
	s := schemaResp.Schema
	if diags := s.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("schema validation failed: %v", diags)
	}

	for _, name := range []string{"name", "resource_group_id", "id"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing envelope attribute %q", name)
		}
	}
	if _, ok := s.Attributes["parent_id"]; ok {
		t.Error("web site should expose resource_group_id, not the generic parent_id")
	}
	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("missing timeouts block")
	}
	for _, name := range []string{"location", "kind", "properties", "identity", "tags"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing body attribute %q", name)
		}
	}
}

func TestInterfaceAssertions(t *testing.T) {
	r := nativeresource.New("azapi_storage_account")
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
	if hookField(t, nativeresource.New("azapi_storage_account"), "ModifyPlan").IsNil() {
		t.Error("storage account ModifyPlan hook is nil")
	}
}

func TestBlobServiceSkipARMDelete(t *testing.T) {
	// blobServices/default is a singleton with no ARM delete operation, so its
	// generated-service hook registers a state-only delete (SkipARMDelete) to avoid a 405.
	if !hookField(t, nativeresource.New("azapi_storage_account_blob_service"), "SkipARMDelete").Bool() {
		t.Error("blob service SkipARMDelete should be true")
	}

	ctx := context.Background()

	// With SkipARMDelete, Delete returns before touching the provider (nil here) and
	// records no error — the framework then removes the resource from state.
	skip := nativeresource.New("azapi_storage_account_blob_service")
	resp := &resource.DeleteResponse{}
	skip.Delete(ctx, resource.DeleteRequest{}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("SkipARMDelete Delete should not error, got: %v", resp.Diagnostics)
	}

	// Without the flag the same nil-provider Delete reaches the provider check and
	// errors — confirming the early return is what bypasses the ARM delete path.
	plain := nativeresource.New("azapi_storage_account")
	resp2 := &resource.DeleteResponse{}
	plain.Delete(ctx, resource.DeleteRequest{}, resp2)
	if !resp2.Diagnostics.HasError() {
		t.Error("Delete without SkipARMDelete and nil provider should error")
	}
}

func TestWebSiteConfigurationHookRegistered(t *testing.T) {
	webSite := nativeresource.New("azapi_web_site")
	if hookField(t, webSite, "AfterCreate").IsNil() || hookField(t, webSite, "AfterUpdate").IsNil() || hookField(t, webSite, "AfterRead").IsNil() {
		t.Fatal("web site configuration read hooks must be registered for create, update, and read")
	}
}
