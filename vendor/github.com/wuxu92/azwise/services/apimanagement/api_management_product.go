package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementProduct provides resource knowledge for
// Microsoft.ApiManagement/service/products.
//
// Mirrors azurerm_api_management_product.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_product_resource.go
//     schema (lines 41-84), CreateUpdate body (lines 118-147), timeouts (30m/5m/30m/30m)
//   - go-azure-sdk apimanagement/2022-08-01/product ProductContractProperties + ProductState constants
//   - schemaz.SchemaApiManagementChildName -> validate.ApiManagementChildName regex
type ApiManagementProduct struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementProduct)(nil)

func NewApiManagementProduct() *ApiManagementProduct {
	return &ApiManagementProduct{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/products",
			ApiVersions:  []string{"2022-08-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"}, // product_id (SchemaApiManagementChildName, ForceNew)
			},
			RequiredFields: []string{
				"properties.displayName",
			},
			// subscription_required defaults to true in AzureRM schema.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.subscriptionRequired", Value: true},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name (product_id) ── ApiManagementChildName
				{
					Regex:   `^[a-zA-Z0-9]([a-zA-Z0-9-_]{0,78}[a-zA-Z0-9])?$`,
					Message: "may only contain alphanumeric characters, underscores and dashes up to 80 characters in length",
				},
				// ── display_name → properties.displayName ── StringIsNotEmpty
				{PropertyPath: "properties.displayName", MinLength: 1, Message: "display_name must not be empty"},
				// ── published → properties.state ── ProductState enum
				{
					PropertyPath:  "properties.state",
					AllowedValues: []string{"published", "notPublished"},
					Message:       "state must be one of: published, notPublished",
				},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementProduct()) }
