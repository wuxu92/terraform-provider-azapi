package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementTag provides resource knowledge for
// Microsoft.ApiManagement/service/tags.
//
// Mirrors azurerm_api_management_tag. `api_management_id` (parent) and `name`
// are ForceNew envelope references. The only ARM body field is
// properties.displayName (required; AzureRM defaults it to the tag name when
// omitted).
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_tag_resource.go
//     schema (lines 41-62); Create body (lines 99-103); Timeouts 30m/5m/30m/30m.
//   - Microsoft.ApiManagement/service/tags@2022-08-01 tag.TagContractProperties:
//     displayName (required).
type ApiManagementTag struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementTag)(nil)

// NewApiManagementTag returns knowledge for the API Management tag resource.
func NewApiManagementTag() *ApiManagementTag {
	return &ApiManagementTag{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/tags",
			ApiVersions:  []string{"2022-08-01"},
			RequiredFields: []string{
				"properties.displayName",
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "properties.displayName", MinLength: 1, Message: "display_name must not be empty"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementTag()) }
