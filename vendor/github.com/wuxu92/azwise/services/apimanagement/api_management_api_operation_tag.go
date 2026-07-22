package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementApiOperationTag provides resource knowledge for
// Microsoft.ApiManagement/service/apis/operations/tags.
//
// Contributing Terraform resource: azurerm_api_management_api_operation_tag.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_operation_tag_resource.go:22-63
//     (schema: api_operation_id + name Required+ForceNew, display_name Required)
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_operation_tag_resource.go:96-109
//     (create: tagClient.CreateOrUpdate writes display_name onto the sibling tag,
//     then client.TagAssignToOperation associates it — an empty-body PUT)
//   - terraform-provider-azurerm internal/services/apimanagement/validate/api_management.go:12-21
//     (ApiManagementChildName regex used by the tag name segment)
//   - terraform-provider-azurerm vendor/.../apimanagement/2022-08-01/apioperationtag/id_operationtag.go:128-143
//     (resource ID segments: .../apis/{apiId}/operations/{operationId}/tags/{tagId})
//
// Intentionally skipped here:
//   - api_operation_id (validate.ApiOperationID): the parent operation resource ID,
//     an AzAPI ID segment rather than a body property.
//   - display_name (validation.StringIsNotEmpty): AzureRM stores this on the separate
//     Microsoft.ApiManagement/service/tags resource (via tag.TagCreateUpdateParameters),
//     NOT on the operation/tags association, which has no request body. Per the
//     sub-service API separation rule, the display_name rule belongs to the tags
//     resource's knowledge file, not this association.
type ApiManagementApiOperationTag struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementApiOperationTag)(nil)

// NewApiManagementApiOperationTag returns knowledge for the
// Microsoft.ApiManagement/service/apis/operations/tags resource.
func NewApiManagementApiOperationTag() *ApiManagementApiOperationTag {
	return &ApiManagementApiOperationTag{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/apis/operations/tags",
			ApiVersions:  []string{"2022-08-01"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// name (tag id) — validate.ApiManagementChildName.
					Regex:     `^[a-zA-Z0-9]([a-zA-Z0-9-_]{0,78}[a-zA-Z0-9])?$`,
					MaxLength: 80,
					Message:   "may only contain alphanumeric characters, underscores and dashes up to 80 characters in length, beginning and ending with an alphanumeric character",
				},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementApiOperationTag()) }
