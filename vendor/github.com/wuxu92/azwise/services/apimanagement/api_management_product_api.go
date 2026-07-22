package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementProductApi provides resource knowledge for
// Microsoft.ApiManagement/service/products/apis.
//
// Mirrors azurerm_api_management_product_api. This is a pure association
// resource: CreateOrUpdate takes only the resource ID (no request body), so
// there are no ARM body properties. The api_name and product_id are path
// segments (both ForceNew in AzureRM).
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_product_api_resource.go
//     schema (lines 37-45), Create (lines 49-77), timeouts (30m/5m/-/30m)
//   - go-azure-sdk apimanagement/2022-08-01/productapi id_productapi.go
type ApiManagementProductApi struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementProductApi)(nil)

func NewApiManagementProductApi() *ApiManagementProductApi {
	return &ApiManagementProductApi{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/products/apis",
			ApiVersions:  []string{"2022-08-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"}, // api_name (path segment, ForceNew)
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name (api_name) ── ApiManagementApiName
				{
					Regex:     `^[^*#&+]{1,256}$`,
					MaxLength: 256,
					Message:   "may only be up to 256 characters in length and not include the characters `*`, `#`, `&` or `+`",
				},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementProductApi()) }
