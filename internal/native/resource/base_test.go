package resource_test

import (
	"context"
	"testing"

	"reflect"

	// Populate services.Registry and hookRegistry so New("azapi_storage_account") resolves.
	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
	_ "github.com/Azure/terraform-provider-azapi/internal/native/services/all"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	r := nativeresource.New("azapi_service_plan")

	mdResp := &resource.MetadataResponse{}
	r.(resource.Resource).Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "azapi"}, mdResp)
	if mdResp.TypeName != "azapi_service_plan" {
		t.Errorf("TypeName = %q, want azapi_service_plan", mdResp.TypeName)
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

func TestBlobServiceSingletonReset(t *testing.T) {
	// blobServices/default always exists and has no ARM create/delete: its generated-
	// service hook marks it a Singleton default so a create updates the always-present
	// default in place and a destroy resets it to DefaultBody instead of issuing a
	// 405 DELETE.
	blob := nativeresource.New("azapi_storage_account_blob_service")
	singleton := hookField(t, blob, "Singleton")
	if singleton.IsNil() {
		t.Fatal("blob service Singleton hook should be set")
	}
	body := singleton.Elem().FieldByName("DefaultBody")
	if body.IsNil() || body.Len() == 0 {
		t.Fatal("blob service Singleton.DefaultBody must carry the reset body")
	}
	if !body.MapIndex(reflect.ValueOf("properties")).IsValid() {
		t.Error("reset body must contain a properties block")
	}

	// A regular resource is not a Singleton default.
	if !hookField(t, nativeresource.New("azapi_storage_account"), "Singleton").IsNil() {
		t.Error("storage account should not be a Singleton default")
	}

	// Delete now takes the reset path (a PUT) instead of the old no-op state removal,
	// so a nil-provider Delete reaches the provider check and errors.
	resp := &resource.DeleteResponse{}
	blob.Delete(context.Background(), resource.DeleteRequest{}, resp)
	if !resp.Diagnostics.HasError() {
		t.Error("blob service Delete with nil provider should error (reset needs the provider)")
	}
}

func TestWebSiteConfigurationHookRegistered(t *testing.T) {
	webSite := nativeresource.New("azapi_web_site")
	if hookField(t, webSite, "AfterCreate").IsNil() || hookField(t, webSite, "AfterUpdate").IsNil() || hookField(t, webSite, "AfterRead").IsNil() {
		t.Fatal("web site configuration read hooks must be registered for create, update, and read")
	}
}

func TestKeyVaultAccessPoliciesHookRegistered(t *testing.T) {
	keyVault := nativeresource.New("azapi_key_vault")
	if hookField(t, keyVault, "BeforeCreate").IsNil() || hookField(t, keyVault, "BeforeUpdate").IsNil() {
		t.Fatal("key vault ensureAccessPolicies hook must be registered for both create and update")
	}
}

func TestUserAssignedIdentitySchemaComposition(t *testing.T) {
	ctx := context.Background()
	r := nativeresource.New("azapi_user_assigned_identity")

	// Metadata: provider prefix + resource suffix.
	mdResp := &resource.MetadataResponse{}
	r.(resource.Resource).Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "azapi"}, mdResp)
	if mdResp.TypeName != "azapi_user_assigned_identity" {
		t.Errorf("TypeName = %q, want azapi_user_assigned_identity", mdResp.TypeName)
	}

	// Schema: composed (envelope + body), and framework-valid.
	schemaResp := &resource.SchemaResponse{}
	r.(resource.Resource).Schema(ctx, resource.SchemaRequest{}, schemaResp)
	s := schemaResp.Schema
	if diags := s.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("schema validation failed: %v", diags)
	}

	// userAssignedIdentities is resource-group-scoped: its parent reference is
	// named resource_group_id, not subscription_id or parent_id.
	for _, name := range []string{"name", "resource_group_id", "id"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing envelope attribute %q", name)
		}
	}
	for _, absent := range []string{"subscription_id", "parent_id"} {
		if _, ok := s.Attributes[absent]; ok {
			t.Errorf("user assigned identity should expose resource_group_id, not %q", absent)
		}
	}
	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("missing timeouts block")
	}
	// Body attributes present (from the generated schema).
	for _, name := range []string{"location", "properties"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing body attribute %q", name)
		}
	}
}

func TestKeyVaultSchemaComposition(t *testing.T) {
	ctx := context.Background()
	r := nativeresource.New("azapi_key_vault")

	// Metadata: provider prefix + resource suffix.
	mdResp := &resource.MetadataResponse{}
	r.(resource.Resource).Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "azapi"}, mdResp)
	if mdResp.TypeName != "azapi_key_vault" {
		t.Errorf("TypeName = %q, want azapi_key_vault", mdResp.TypeName)
	}

	// Schema: composed (envelope + body), and framework-valid.
	schemaResp := &resource.SchemaResponse{}
	r.(resource.Resource).Schema(ctx, resource.SchemaRequest{}, schemaResp)
	s := schemaResp.Schema
	if diags := s.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("schema validation failed: %v", diags)
	}

	// key vault is resource-group-scoped: its parent reference is
	// named resource_group_id, not subscription_id or parent_id.
	for _, name := range []string{"name", "resource_group_id", "id"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing envelope attribute %q", name)
		}
	}
	for _, absent := range []string{"subscription_id", "parent_id"} {
		if _, ok := s.Attributes[absent]; ok {
			t.Errorf("key vault should expose resource_group_id, not %q", absent)
		}
	}
	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("missing timeouts block")
	}
	// Body attributes present (from the generated schema).
	for _, name := range []string{"location", "properties", "tags"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing body attribute %q", name)
		}
	}
}

func TestRoleDefinitionSchemaComposition(t *testing.T) {
	ctx := context.Background()
	r := nativeresource.New("azapi_role_definition")

	// Metadata: provider prefix + resource suffix.
	mdResp := &resource.MetadataResponse{}
	r.(resource.Resource).Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "azapi"}, mdResp)
	if mdResp.TypeName != "azapi_role_definition" {
		t.Errorf("TypeName = %q, want azapi_role_definition", mdResp.TypeName)
	}

	// Schema: composed (envelope + body), and framework-valid.
	schemaResp := &resource.SchemaResponse{}
	r.(resource.Resource).Schema(ctx, resource.SchemaRequest{}, schemaResp)
	s := schemaResp.Schema
	if diags := s.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("schema validation failed: %v", diags)
	}

	// role definition is an extension/multi-scope resource: its parent reference is
	// an arbitrary Azure scope named scope_id, not resource_group_id or subscription_id.
	for _, name := range []string{"name", "scope_id", "id"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing envelope attribute %q", name)
		}
	}
	// Keyed by scope_id, and role definitions have no location.
	for _, absent := range []string{"resource_group_id", "subscription_id", "parent_id", "location"} {
		if _, ok := s.Attributes[absent]; ok {
			t.Errorf("role definition should be keyed by scope_id with no location, not expose %q", absent)
		}
	}
	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("missing timeouts block")
	}
	// The name envelope attribute is Required (RequiresReplace UUID).
	if name, ok := s.Attributes["name"]; ok && !name.IsRequired() {
		t.Error(`envelope "name" attribute must be Required`)
	}

	// Body: properties nested attribute present; a later cast panics if absent.
	props, ok := s.Attributes["properties"].(schema.SingleNestedAttribute)
	if !ok {
		t.Fatalf("properties attribute is %T, want schema.SingleNestedAttribute", s.Attributes["properties"])
	}
	// Customizer promoted the bicep-optional role_name to Required.
	if roleName, ok := props.Attributes["role_name"]; !ok {
		t.Error("missing properties.role_name attribute")
	} else if !roleName.IsRequired() {
		t.Error("properties.role_name must be Required (customizer promotion)")
	}
	// azwise overlay wired the CustomRole default onto type.
	if typeAttr, ok := props.Attributes["type"].(schema.StringAttribute); !ok {
		t.Fatalf("properties.type is %T, want schema.StringAttribute", props.Attributes["type"])
	} else if typeAttr.Default == nil {
		t.Error("properties.type must have a Default (CustomRole)")
	}
}

func TestRoleDefinitionAssignableScopesHookRegistered(t *testing.T) {
	roleDefinition := nativeresource.New("azapi_role_definition")
	if hookField(t, roleDefinition, "BeforeCreate").IsNil() || hookField(t, roleDefinition, "BeforeUpdate").IsNil() {
		t.Fatal("role definition ensureAssignableScopes hook must be registered for both create and update")
	}
}

func TestVirtualNetworkSchemaComposition(t *testing.T) {
	ctx := context.Background()
	r := nativeresource.New("azapi_virtual_network")

	// Metadata: provider prefix + resource suffix.
	mdResp := &resource.MetadataResponse{}
	r.(resource.Resource).Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "azapi"}, mdResp)
	if mdResp.TypeName != "azapi_virtual_network" {
		t.Errorf("TypeName = %q, want azapi_virtual_network", mdResp.TypeName)
	}

	// Schema: composed (envelope + body), and framework-valid.
	schemaResp := &resource.SchemaResponse{}
	r.(resource.Resource).Schema(ctx, resource.SchemaRequest{}, schemaResp)
	s := schemaResp.Schema
	if diags := s.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("schema validation failed: %v", diags)
	}

	// virtualNetworks is resource-group scoped: parent reference resource_group_id.
	for _, name := range []string{"name", "resource_group_id", "id"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing envelope attribute %q", name)
		}
	}
	if _, ok := s.Attributes["parent_id"]; ok {
		t.Error("virtual network should expose resource_group_id, not the generic parent_id")
	}
	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("missing timeouts block")
	}
	// Body attributes present (from the generated schema).
	for _, name := range []string{"location", "properties"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing body attribute %q", name)
		}
	}
}

func TestKustoClusterSchemaComposition(t *testing.T) {
	ctx := context.Background()
	r := nativeresource.New("azapi_kusto_cluster")

	mdResp := &resource.MetadataResponse{}
	r.(resource.Resource).Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "azapi"}, mdResp)
	if mdResp.TypeName != "azapi_kusto_cluster" {
		t.Errorf("TypeName = %q, want azapi_kusto_cluster", mdResp.TypeName)
	}

	schemaResp := &resource.SchemaResponse{}
	r.(resource.Resource).Schema(ctx, resource.SchemaRequest{}, schemaResp)
	s := schemaResp.Schema
	if diags := s.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("schema validation failed: %v", diags)
	}

	// clusters is resource-group scoped: parent reference resource_group_id.
	for _, name := range []string{"name", "resource_group_id", "id"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing envelope attribute %q", name)
		}
	}
	if _, ok := s.Attributes["parent_id"]; ok {
		t.Error("kusto cluster should expose resource_group_id, not the generic parent_id")
	}
	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("missing timeouts block")
	}
	// A plain (non-discriminated) body: sku + properties present, required sku.
	for _, name := range []string{"location", "sku", "properties"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing body attribute %q", name)
		}
	}
}

// TestKustoClusterDatabaseSchemaComposition covers a resource whose body is a
// discriminated ROOT (kind: ReadWrite / ReadOnlyFollowing): the two variant blocks
// sit at the envelope root beside the cluster_id parent reference, and the
// discriminator `kind` is synthesized (never a schema attribute).
func TestKustoClusterDatabaseSchemaComposition(t *testing.T) {
	ctx := context.Background()
	r := nativeresource.New("azapi_kusto_database")

	mdResp := &resource.MetadataResponse{}
	r.(resource.Resource).Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "azapi"}, mdResp)
	if mdResp.TypeName != "azapi_kusto_database" {
		t.Errorf("TypeName = %q, want azapi_kusto_database", mdResp.TypeName)
	}

	schemaResp := &resource.SchemaResponse{}
	r.(resource.Resource).Schema(ctx, resource.SchemaRequest{}, schemaResp)
	s := schemaResp.Schema
	if diags := s.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("schema validation failed: %v", diags)
	}

	// databases is a child of a cluster: parent reference cluster_id.
	for _, name := range []string{"name", "cluster_id", "id"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing envelope attribute %q", name)
		}
	}
	for _, absent := range []string{"resource_group_id", "parent_id"} {
		if _, ok := s.Attributes[absent]; ok {
			t.Errorf("kusto database should expose cluster_id, not %q", absent)
		}
	}
	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("missing timeouts block")
	}
	// Discriminated root: one variant block per discriminator value at the top level.
	for _, name := range []string{"read_write", "read_only_following"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing discriminated variant block %q", name)
		}
	}
	// The discriminator itself is synthesized by the mapper, never surfaced.
	if _, ok := s.Attributes["kind"]; ok {
		t.Error("discriminator `kind` must not surface as a schema attribute")
	}
}

// TestDocumentDBDatabaseAccountSchemaComposition covers a resource with a nested
// discriminated property: properties.backup_policy carries per-variant blocks
// (periodic / continuous) with the discriminator `type` synthesized, not surfaced.
func TestDocumentDBDatabaseAccountSchemaComposition(t *testing.T) {
	ctx := context.Background()
	r := nativeresource.New("azapi_cosmosdb_account")

	mdResp := &resource.MetadataResponse{}
	r.(resource.Resource).Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "azapi"}, mdResp)
	if mdResp.TypeName != "azapi_cosmosdb_account" {
		t.Errorf("TypeName = %q, want azapi_cosmosdb_account", mdResp.TypeName)
	}

	schemaResp := &resource.SchemaResponse{}
	r.(resource.Resource).Schema(ctx, resource.SchemaRequest{}, schemaResp)
	s := schemaResp.Schema
	if diags := s.ValidateImplementation(ctx); diags.HasError() {
		t.Fatalf("schema validation failed: %v", diags)
	}

	// databaseAccounts is resource-group scoped: parent reference resource_group_id.
	for _, name := range []string{"name", "resource_group_id", "id"} {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing envelope attribute %q", name)
		}
	}
	if _, ok := s.Attributes["parent_id"]; ok {
		t.Error("cosmos account should expose resource_group_id, not the generic parent_id")
	}
	if _, ok := s.Blocks["timeouts"]; !ok {
		t.Error("missing timeouts block")
	}
	// The nested discriminated property lives under properties.backup_policy with
	// periodic/continuous variant blocks; assert the top-level body attr composes.
	if _, ok := s.Attributes["properties"]; !ok {
		t.Error("missing body attribute \"properties\"")
	}
}
