package azwise

import "time"

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
//   - Microsoft.Authorization/roleAssignments@2022-04-01 bicep type graph
//     RoleAssignmentProperties: roleDefinitionId/principalId required; created*/updated*/scope
//     are read-only in the ARM schema.
type RoleAssignment struct {
	BaseKnowledge
}

var _ ResourceKnowledge = (*RoleAssignment)(nil)

// NewRoleAssignment returns knowledge for the roleAssignments resource.
func NewRoleAssignment() *RoleAssignment {
	return &RoleAssignment{
		BaseKnowledge: BaseKnowledge{
			ResourceType: "Microsoft.Authorization/roleAssignments",
			ApiVersions:  []string{"2022-04-01"},
			ForceNew: []ForceNewRule{
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
			TimeoutsConfig: &Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []StringRule{
				{PropertyPath: "properties.condition", MinLength: 1, Message: "condition must not be empty"},
				{PropertyPath: "properties.conditionVersion", AllowedValues: []string{"1.0", "2.0"}, Message: "condition version must be 1.0 or 2.0"},
				{PropertyPath: "properties.description", MinLength: 1, Message: "description must not be empty"},
			},
		},
	}
}
