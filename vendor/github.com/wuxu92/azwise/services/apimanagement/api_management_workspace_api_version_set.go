package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementWorkspaceApiVersionSet provides resource knowledge for
// Microsoft.ApiManagement/service/workspaces/apiVersionSets.
//
// Mirrors azurerm_api_management_workspace_api_version_set. `name` and
// `api_management_workspace_id` are envelope/parent references (ForceNew).
// version_header_name and version_query_name are ConflictsWith.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_workspace_api_version_set_resource.go
//     Arguments (lines 52-98); Create body (lines 133-150); CustomizeDiff
//     (lines 289-330); Timeout 30m.
//   - Microsoft.ApiManagement/service/workspaces/apiVersionSets@2024-05-01
//     apiversionset.ApiVersionSetContractProperties: displayName (required),
//     versioningScheme (enum Header|Query|Segment, required), description,
//     versionHeaderName, versionQueryName.
type ApiManagementWorkspaceApiVersionSet struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementWorkspaceApiVersionSet)(nil)

// NewApiManagementWorkspaceApiVersionSet returns knowledge for the workspace API version set resource.
func NewApiManagementWorkspaceApiVersionSet() *ApiManagementWorkspaceApiVersionSet {
	return &ApiManagementWorkspaceApiVersionSet{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/workspaces/apiVersionSets",
			ApiVersions:  []string{"2024-05-01"},
			RequiredFields: []string{
				"properties.displayName",
				"properties.versioningScheme",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.versioningScheme",
					AllowedValues: []string{"Header", "Query", "Segment"},
					Message:       "versioning_scheme must be one of Header, Query, Segment",
				},
				{PropertyPath: "properties.displayName", MinLength: 1, Message: "display_name must not be empty"},
			},
			// ConflictsWith: version_header_name vs version_query_name.
			ConflictsWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.versionHeaderName", "properties.versionQueryName"},
					Message: "version_header_name and version_query_name cannot be set together",
				},
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

func init() { azwise.Register(NewApiManagementWorkspaceApiVersionSet()) }
