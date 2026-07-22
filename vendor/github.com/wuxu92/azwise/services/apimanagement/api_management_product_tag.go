package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementProductTag provides resource knowledge for
// Microsoft.ApiManagement/service/products/tags.
//
// Mirrors azurerm_api_management_product_tag. This is a pure association
// resource: TagAssignToProduct takes only the resource ID (no request body), so
// there are no ARM body properties. The tag name and product id are path
// segments (both ForceNew in AzureRM).
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_product_tag_resource.go
//     schema (lines 38-46), Create (lines 50-82), timeouts (30m/5m/-/30m)
//   - go-azure-sdk apimanagement/2022-08-01/producttag id_producttag.go
type ApiManagementProductTag struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementProductTag)(nil)

func NewApiManagementProductTag() *ApiManagementProductTag {
	return &ApiManagementProductTag{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/products/tags",
			ApiVersions:  []string{"2022-08-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"}, // tag name (SchemaApiManagementChildName, ForceNew)
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name (tag name) ── ApiManagementChildName
				{
					Regex:   `^[a-zA-Z0-9]([a-zA-Z0-9-_]{0,78}[a-zA-Z0-9])?$`,
					Message: "may only contain alphanumeric characters, underscores and dashes up to 80 characters in length",
				},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementProductTag()) }
