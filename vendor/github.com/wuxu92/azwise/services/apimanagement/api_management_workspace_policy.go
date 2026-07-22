package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementWorkspacePolicy provides resource knowledge for
// Microsoft.ApiManagement/service/workspaces/policies (singleton child named
// "policy" under a workspace).
//
// Mirrors azurerm_api_management_workspace_policy. `api_management_workspace_id`
// is the parent reference (ForceNew). AzureRM's xml_content / xml_link are
// ExactlyOneOf convenience surfaces that both expand into ARM properties.value
// with a differing properties.format (rawxml vs rawxml-link).
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_workspace_policy_resource.go
//     Arguments (lines 44-63); Create body (lines 97-111); Timeout 30m.
//   - Microsoft.ApiManagement/service/workspaces/policies@2024-05-01
//     workspacepolicy.PolicyContractProperties: value (required),
//     format (enum rawxml|rawxml-link|xml|xml-link).
type ApiManagementWorkspacePolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementWorkspacePolicy)(nil)

// NewApiManagementWorkspacePolicy returns knowledge for the workspace policy resource.
func NewApiManagementWorkspacePolicy() *ApiManagementWorkspacePolicy {
	return &ApiManagementWorkspacePolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/workspaces/policies",
			ApiVersions:  []string{"2024-05-01"},
			RequiredFields: []string{
				"properties.value",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.format",
					AllowedValues: []string{"rawxml", "rawxml-link", "xml", "xml-link"},
					Message:       "format must be one of rawxml, rawxml-link, xml, xml-link",
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

func init() { azwise.Register(NewApiManagementWorkspacePolicy()) }
