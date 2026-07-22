package authorization

import (
	"time"

	"github.com/wuxu92/azwise"
)

// RoleAssignment provides resource knowledge for
// Microsoft.Authorization/roleAssignments.
//
// Mirrors azurerm_role_assignment where it maps to the ARM management-plane model.
// The ARM resource name is the role assignment GUID (azurerm's optional name), and
// the parent_id envelope is the assignment scope (azurerm's scope). AzureRM's
// role_definition_name convenience and skip_service_principal_aad_check flag are
// provider-only surfaces and are intentionally not represented in the native ARM
// schema.
//
// Sources:
//   - terraform-provider-azurerm internal/services/authorization/role_assignment_resource.go
//     schema + Create/Read/Delete (timeouts 30m/5m/30m; roleAssignment fields ForceNew)
//   - terraform-provider-azurerm internal/services/authorization/marketplace_role_assignment_resource.go
//     same Microsoft.Authorization/roleAssignments type (marketplace scope); ForceNew/Required
//     set is a subset, adds role_definition_id StringIsNotEmpty (marketplace L81)
//   - Microsoft.Authorization/roleAssignments@2022-04-01 bicep type graph
//     RoleAssignmentProperties: roleDefinitionId/principalId required; created*/updated*/scope
//     are read-only in the ARM schema.
type RoleAssignment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*RoleAssignment)(nil)

// NewRoleAssignment returns knowledge for the roleAssignments resource.
func NewRoleAssignment() *RoleAssignment {
	return &RoleAssignment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Authorization/roleAssignments",
			ApiVersions:  []string{"2022-04-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.roleDefinitionId"},
				{PropertyPath: "properties.principalId"},
				{PropertyPath: "properties.principalType"},
				{PropertyPath: "properties.delegatedManagedIdentityResourceId"},
				{PropertyPath: "properties.description"},
				{PropertyPath: "properties.condition"},
				{PropertyPath: "properties.conditionVersion"},
			},
			RequiredFields: []string{
				"properties.roleDefinitionId",
				"properties.principalId",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "properties.roleDefinitionId", MinLength: 1, Message: "role definition id must not be empty"},
				{PropertyPath: "properties.condition", MinLength: 1, Message: "condition must not be empty"},
				{PropertyPath: "properties.conditionVersion", AllowedValues: []string{"1.0", "2.0"}, Message: "condition version must be 1.0 or 2.0"},
				{PropertyPath: "properties.description", MinLength: 1, Message: "description must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewRoleAssignment()) }
