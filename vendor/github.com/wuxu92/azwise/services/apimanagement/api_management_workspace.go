package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementWorkspace provides resource knowledge for
// Microsoft.ApiManagement/service/workspaces.
//
// Mirrors azurerm_api_management_workspace. `name` and `api_management_id` are
// envelope/parent references (ForceNew). The only ARM body fields are
// properties.displayName (required) and properties.description.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_workspace_resource.go
//     Arguments (lines 45-71); Create body (lines 106-114); Timeouts 30m/5m/30m/30m.
//   - Microsoft.ApiManagement/service/workspaces@2024-05-01 workspace.WorkspaceContractProperties:
//     displayName (required), description (optional).
type ApiManagementWorkspace struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementWorkspace)(nil)

// NewApiManagementWorkspace returns knowledge for the API Management workspace resource.
func NewApiManagementWorkspace() *ApiManagementWorkspace {
	return &ApiManagementWorkspace{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/workspaces",
			ApiVersions:  []string{"2024-05-01"},
			RequiredFields: []string{
				"properties.displayName",
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "properties.displayName", MinLength: 1, Message: "display_name must not be empty"},
				{PropertyPath: "properties.description", MinLength: 1, Message: "description must not be empty"},
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

func init() { azwise.Register(NewApiManagementWorkspace()) }
