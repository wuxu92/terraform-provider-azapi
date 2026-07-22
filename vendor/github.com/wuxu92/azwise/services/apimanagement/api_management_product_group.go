package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementProductGroup provides resource knowledge for
// Microsoft.ApiManagement/service/products/groups.
//
// Mirrors azurerm_api_management_product_group. This is a pure association
// resource: CreateOrUpdate takes only the resource ID (no request body), so
// there are no ARM body properties. The group_name and product_id are path
// segments (both ForceNew in AzureRM).
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_product_group_resource.go
//     schema (lines 37-45), Create (lines 49-77), timeouts (30m/5m/-/30m)
//   - go-azure-sdk apimanagement/2022-08-01/productgroup id_productgroup.go
type ApiManagementProductGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementProductGroup)(nil)

func NewApiManagementProductGroup() *ApiManagementProductGroup {
	return &ApiManagementProductGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/products/groups",
			ApiVersions:  []string{"2022-08-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"}, // group_name (SchemaApiManagementChildName, ForceNew)
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name (group_name) ── ApiManagementChildName
				{
					Regex:   `^[a-zA-Z0-9]([a-zA-Z0-9-_]{0,78}[a-zA-Z0-9])?$`,
					Message: "may only contain alphanumeric characters, underscores and dashes up to 80 characters in length",
				},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementProductGroup()) }
