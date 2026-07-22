package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementWorkspacePolicyFragment provides resource knowledge for
// Microsoft.ApiManagement/service/workspaces/policyFragments.
//
// Mirrors azurerm_api_management_workspace_policy_fragment. `name` and
// `api_management_workspace_id` are envelope/parent references (ForceNew).
// xml_content -> properties.value (required); xml_format -> properties.format
// (enum, default "xml"); description -> properties.description.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_workspace_policy_fragment_resource.go
//     Arguments (lines 45-70); Create body (lines 107-116); Timeouts 30m/5m/-/30m.
//   - Microsoft.ApiManagement/service/workspaces/policyFragments@2024-05-01
//     policyfragment.PolicyFragmentContractProperties: value (required),
//     format (enum rawxml|xml), description.
type ApiManagementWorkspacePolicyFragment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementWorkspacePolicyFragment)(nil)

// NewApiManagementWorkspacePolicyFragment returns knowledge for the workspace policy fragment resource.
func NewApiManagementWorkspacePolicyFragment() *ApiManagementWorkspacePolicyFragment {
	return &ApiManagementWorkspacePolicyFragment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/workspaces/policyFragments",
			ApiVersions:  []string{"2024-05-01"},
			RequiredFields: []string{
				"properties.value",
			},
			DefaultValues: []azwise.DefaultValue{
				// xml_format default "xml" (schema line 66).
				{PropertyPath: "properties.format", Value: "xml"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.format",
					AllowedValues: []string{"rawxml", "xml"},
					Message:       "xml_format must be one of rawxml, xml",
				},
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

func init() { azwise.Register(NewApiManagementWorkspacePolicyFragment()) }
