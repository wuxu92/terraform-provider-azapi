package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CosmosDbSqlRoleDefinition provides resource knowledge for
// Microsoft.DocumentDB/databaseAccounts/sqlRoleDefinitions
// (azurerm_cosmosdb_sql_role_definition).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_sql_role_definition_resource.go
//     :33-38 (timeouts 30/5/30/30), :45-105 (schema: role_definition_id IsUUID,
//     type RoleDefinitionType enum default CustomRole, assignable_scopes,
//     name->roleName, permissions), :140-147 (expand under properties)
//   - go-azure-sdk resource-manager/cosmosdb/2024-08-15/rbacs:
//     model_sqlroledefinitioncreateupdateparameters.go, model_sqlroledefinitionresource.go,
//     model_permission.go, constants.go PossibleValuesForRoleDefinitionType
//
// Not encoded (deliberate):
//   - role_definition_id is the resource name (a UUID, validation.IsUUID) carried in
//     the URL, not an ARM body property; route the UUID check as a customizer validator.
//   - permissions[*].dataActions and assignableScopes[*] are array-element paths;
//     azwise cannot lower a rule through an array element, so they are skipped.
type CosmosDbSqlRoleDefinition struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CosmosDbSqlRoleDefinition)(nil)

func NewCosmosDbSqlRoleDefinition() *CosmosDbSqlRoleDefinition {
	return &CosmosDbSqlRoleDefinition{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/databaseAccounts/sqlRoleDefinitions",
			ApiVersions:  []string{"2024-08-15"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.type"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "properties.type", AllowedValues: []string{"BuiltInRole", "CustomRole"}},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.type", Value: "CustomRole"},
			},
			RequiredFields: []string{
				"properties.roleName",
				"properties.assignableScopes",
				"properties.permissions",
			},
		},
	}
}

func init() { azwise.Register(NewCosmosDbSqlRoleDefinition()) }
