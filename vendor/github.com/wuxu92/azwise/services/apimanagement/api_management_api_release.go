package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementApiRelease provides resource knowledge for
// Microsoft.ApiManagement/service/apis/releases.
//
// Contributing Terraform resource: azurerm_api_management_api_release.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_release_resource.go:23-64
//     (schema: name + api_id Required+ForceNew, notes Optional)
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_release_resource.go:96-101
//     (create: ApiReleaseContractProperties.ApiId + .Notes)
//   - terraform-provider-azurerm internal/services/apimanagement/validate/api_management.go:12-21
//     (ApiManagementChildName regex used by the release name)
//   - terraform-provider-azurerm vendor/.../apimanagement/2022-08-01/apirelease/id_release.go:122-135
//     (resource ID segments: .../apis/{apiId}/releases/{releaseId})
//
// Notes:
//   - api_id is Required+ForceNew and is also serialised into the body as
//     properties.apiId (the API this release targets), so it is both a ForceNew body
//     property and a required field.
type ApiManagementApiRelease struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementApiRelease)(nil)

// NewApiManagementApiRelease returns knowledge for the
// Microsoft.ApiManagement/service/apis/releases resource.
func NewApiManagementApiRelease() *ApiManagementApiRelease {
	return &ApiManagementApiRelease{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/apis/releases",
			ApiVersions:  []string{"2022-08-01"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.apiId"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// name — validate.ApiManagementChildName.
					Regex:     `^[a-zA-Z0-9]([a-zA-Z0-9-_]{0,78}[a-zA-Z0-9])?$`,
					MaxLength: 80,
					Message:   "may only contain alphanumeric characters, underscores and dashes up to 80 characters in length, beginning and ending with an alphanumeric character",
				},
				{
					PropertyPath: "properties.notes",
					MinLength:    1,
					Message:      "must not be empty",
				},
			},
			RequiredFields: []string{
				"properties.apiId",
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementApiRelease()) }
