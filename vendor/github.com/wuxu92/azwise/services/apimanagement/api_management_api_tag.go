package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementApiTag provides resource knowledge for
// Microsoft.ApiManagement/service/apis/tags.
//
// Contributing Terraform resource: azurerm_api_management_api_tag.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_tag_resource.go:21-53
//     (schema: api_id + name, both Required+ForceNew; Create/Read/Delete only, no Update)
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_tag_resource.go:91-93
//     (create: client.TagAssignToApi — a pure tag-to-API association PUT with no request body)
//   - terraform-provider-azurerm vendor/.../apimanagement/2022-08-01/apitag/id_apitag.go:121-135
//     (resource ID segments: .../apis/{apiId}/tags/{tagId})
//
// Intentionally skipped here:
//   - api_id (api.ValidateApiID): the parent API resource ID, an AzAPI ID segment
//     rather than a body property of the tag association.
//   - name: the tag id (name segment). AzureRM applies no ValidateFunc, so no rule.
//   - This association carries NO ARM request body; the tag's own display_name lives
//     on the sibling Microsoft.ApiManagement/service/tags resource, not here.
type ApiManagementApiTag struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementApiTag)(nil)

// NewApiManagementApiTag returns knowledge for the
// Microsoft.ApiManagement/service/apis/tags resource.
func NewApiManagementApiTag() *ApiManagementApiTag {
	return &ApiManagementApiTag{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/apis/tags",
			ApiVersions:  []string{"2022-08-01"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementApiTag()) }
