package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CosmosDbSqlRoleAssignment provides resource knowledge for
// Microsoft.DocumentDB/databaseAccounts/sqlRoleAssignments
// (azurerm_cosmosdb_sql_role_assignment).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_sql_role_assignment_resource.go
//     :33-38 (timeouts 30/5/30/30), :45-82 (schema: name IsUUID, principal_id IsUUID,
//     scope, role_definition_id ValidateSqlRoleDefinitionID), :117-123 (expand)
//   - go-azure-sdk resource-manager/cosmosdb/2024-08-15/rbacs:
//     model_sqlroleassignmentcreateupdateparameters.go, model_sqlroleassignmentresource.go
//
// Not encoded (deliberate):
//   - name is the resource name (a UUID, validation.IsUUID) carried in the URL, not
//     an ARM body property; route the UUID check as a customizer validator.
//   - principal_id (IsUUID) and role_definition_id (ValidateSqlRoleDefinitionID) are
//     semantic validators, routed to customizer validators, not declarative StringRules.
type CosmosDbSqlRoleAssignment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CosmosDbSqlRoleAssignment)(nil)

func NewCosmosDbSqlRoleAssignment() *CosmosDbSqlRoleAssignment {
	return &CosmosDbSqlRoleAssignment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/databaseAccounts/sqlRoleAssignments",
			ApiVersions:  []string{"2024-08-15"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.principalId"},
				{PropertyPath: "properties.scope"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.principalId",
				"properties.scope",
				"properties.roleDefinitionId",
			},
		},
	}
}

func init() { azwise.Register(NewCosmosDbSqlRoleAssignment()) }
