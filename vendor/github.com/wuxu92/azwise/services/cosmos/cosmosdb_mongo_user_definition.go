package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CosmosDbMongoUserDefinition provides resource knowledge for
// Microsoft.DocumentDB/databaseAccounts/mongodbUserDefinitions
// (azurerm_cosmosdb_mongo_user_definition).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_mongo_user_definition_resource.go
//     :44-76 (schema: cosmos_mongo_database_id, username ForceNew, password Sensitive,
//     inherited_role_names), :82-132 (Create, timeout 30m; databaseName derived from
//     parent id, mechanisms hardcoded SCRAM-SHA-256, roles under properties)
//   - go-azure-sdk resource-manager/cosmosdb/2025-10-15/mongorbacs:
//     model_mongouserdefinitioncreateupdateparameters.go, model_mongouserdefinitionresource.go
//
// Not encoded (deliberate):
//   - cosmos_mongo_database_id is a parent reference (URL segment); databaseName is
//     derived from it and is required in the ARM body.
//   - inherited_role_names -> roles[*] is an array-element path.
//   - Read/Update/Delete timeouts are not declared in the typed SDK resource; only
//     Create uses 30m, so the others fall back to provider defaults.
type CosmosDbMongoUserDefinition struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CosmosDbMongoUserDefinition)(nil)

func NewCosmosDbMongoUserDefinition() *CosmosDbMongoUserDefinition {
	return &CosmosDbMongoUserDefinition{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/databaseAccounts/mongodbUserDefinitions",
			ApiVersions:  []string{"2025-10-15"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.userName"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
			},
			SensitiveFields: []string{"properties.password"},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.mechanisms", Value: "SCRAM-SHA-256"}, // AzureRM hardcodes SCRAM-SHA-256
			},
			RequiredFields: []string{
				"properties.userName",
				"properties.password",
				"properties.databaseName",
			},
		},
	}
}

func init() { azwise.Register(NewCosmosDbMongoUserDefinition()) }
