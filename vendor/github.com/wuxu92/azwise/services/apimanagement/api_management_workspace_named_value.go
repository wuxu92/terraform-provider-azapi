package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementWorkspaceNamedValue provides resource knowledge for
// Microsoft.ApiManagement/service/workspaces/namedValues.
//
// Mirrors azurerm_api_management_workspace_named_value. `name` and
// `api_management_workspace_id` are envelope/parent references (ForceNew).
// value (sensitive) and value_from_key_vault (-> keyVault.secretIdentifier)
// are ExactlyOneOf. `tags` here is an ARM list of strings, not the common
// tags map.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_workspace_named_value_resource.go
//     Arguments (lines 55-122); CustomizeDiff (lines 128-143, secret must be
//     true when value_from_key_vault set); Create body (lines 174-185); Timeout 30m.
//   - Microsoft.ApiManagement/service/workspaces/namedValues@2024-05-01
//     namedvalue.NamedValueCreateContractProperties: displayName (required),
//     keyVault{secretIdentifier,identityClientId}, secret, tags, value.
type ApiManagementWorkspaceNamedValue struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementWorkspaceNamedValue)(nil)

// NewApiManagementWorkspaceNamedValue returns knowledge for the workspace named value resource.
func NewApiManagementWorkspaceNamedValue() *ApiManagementWorkspaceNamedValue {
	return &ApiManagementWorkspaceNamedValue{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/workspaces/namedValues",
			ApiVersions:  []string{"2024-05-01"},
			RequiredFields: []string{
				"properties.displayName",
			},
			SensitiveFields: []string{
				"properties.value",
			},
			DefaultValues: []azwise.DefaultValue{
				// secret default false (schema line 81).
				{PropertyPath: "properties.secret", Value: false},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "properties.displayName",
					Regex:        `^[A-Za-z0-9-._]{1,256}$`,
					MaxLength:    256,
					Message:      "display_name must be 1-256 characters of letters, numbers, hyphens, periods, and underscores",
				},
			},
			// ExactlyOneOf: value vs value_from_key_vault (keyVault.secretIdentifier).
			ExactlyOneOf: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.value", "properties.keyVault.secretIdentifier"},
					Message: "exactly one of value or value_from_key_vault must be set",
				},
			},
			// TODO: CustomizeDiff requires properties.secret == true when
			// properties.keyVault is set. Not expressible with current rule types.
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementWorkspaceNamedValue()) }
