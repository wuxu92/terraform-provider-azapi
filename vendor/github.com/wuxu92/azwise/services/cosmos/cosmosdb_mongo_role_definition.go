package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CosmosDbMongoRoleDefinition provides resource knowledge for
// Microsoft.DocumentDB/databaseAccounts/mongodbRoleDefinitions
// (azurerm_cosmosdb_mongo_role_definition).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_mongo_role_definition_resource.go
//     :55-118 (schema: cosmos_mongo_database_id, role_name ForceNew, inherited_role_names,
//     privilege), :124-175 (Create, timeout 30m; databaseName derived from parent id,
//     type hardcoded CustomRole, roles/privileges under properties)
//   - go-azure-sdk resource-manager/cosmosdb/2025-10-15/mongorbacs:
//     model_mongoroledefinitioncreateupdateparameters.go, model_mongoroledefinitionresource.go,
//     model_privilege.go, model_role.go
//
// Not encoded (deliberate):
//   - cosmos_mongo_database_id is a parent reference (URL segment); databaseName is
//     derived from it and is required in the ARM body.
//   - privilege[*].actions / privilege[*].resource and inherited_role_names -> roles[*]
//     are array-element paths; azwise cannot lower a rule through an array element.
//   - Read/Update/Delete timeouts are not declared in the typed SDK resource; only
//     Create uses 30m, so the others fall back to provider defaults.
type CosmosDbMongoRoleDefinition struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CosmosDbMongoRoleDefinition)(nil)

func NewCosmosDbMongoRoleDefinition() *CosmosDbMongoRoleDefinition {
	return &CosmosDbMongoRoleDefinition{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/databaseAccounts/mongodbRoleDefinitions",
			ApiVersions:  []string{"2025-10-15"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.roleName"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.type", Value: "CustomRole"}, // AzureRM hardcodes CustomRole
			},
			RequiredFields: []string{
				"properties.roleName",
				"properties.databaseName",
			},
		},
	}
}

func init() { azwise.Register(NewCosmosDbMongoRoleDefinition()) }
