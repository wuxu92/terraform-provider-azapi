package resource

import (
	"github.com/Azure/terraform-provider-azapi/internal/native/services"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// testResourceGroupSchema is a lean, test-only stand-in for the generated
// azapi_resource_group schema (services/resources). The runtime-engine tests drive
// Base/DataSource against it WITHOUT importing a concrete generated service package:
// every generated service package imports this (resource) package for its hooks, so
// a test here importing services/resources would form an import cycle (foreseen in
// base_hook_contract_test.resourceGroupDescriptor). Only attribute names and types
// matter — decode/flatten/compose read the object type, not validators or plan
// modifiers — so those are intentionally omitted. The ARM type + API version match
// the real resource so loadBody resolves the same embedded body graph.
func testResourceGroupSchema() schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name":            schema.StringAttribute{Required: true},
			"subscription_id": schema.StringAttribute{Required: true},
			"location":        schema.StringAttribute{Required: true},
			"managed_by":      schema.StringAttribute{Optional: true, Computed: true},
			"properties": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"provisioning_state": schema.StringAttribute{Computed: true},
				},
			},
			"system_data": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"created_at":            schema.StringAttribute{Optional: true, Computed: true},
					"created_by":            schema.StringAttribute{Optional: true, Computed: true},
					"created_by_type":       schema.StringAttribute{Optional: true, Computed: true},
					"last_modified_at":      schema.StringAttribute{Optional: true, Computed: true},
					"last_modified_by":      schema.StringAttribute{Optional: true, Computed: true},
					"last_modified_by_type": schema.StringAttribute{Optional: true, Computed: true},
				},
			},
			"tags": schema.MapAttribute{Optional: true, Computed: true, ElementType: types.StringType},
			"id":   schema.StringAttribute{Computed: true},
		},
	}
}

// Register azapi_resource_group into services.Registry so registry-driven tests
// (NewDataSource / New) resolve it without importing services/resources.
func init() {
	services.Register(services.Descriptor{
		Name:       "azapi_resource_group",
		ARMType:    "Microsoft.Resources/resourceGroups",
		APIVersion: "2025-04-01",
		Schema:     testResourceGroupSchema,
		ParentAttr: "subscription_id",
	})
}
