package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CosmosDbMongoDatabase provides resource knowledge for
// Microsoft.DocumentDB/databaseAccounts/mongodbDatabases (azurerm_cosmosdb_mongo_database).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_mongo_database_resource.go
//     :41-46 (timeouts 30/5/30/30), :48-73 (schema: name CosmosEntityName,
//     throughput CosmosThroughput, autoscale_settings), :95-112 (expand:
//     throughput/autoscaleSettings under properties.options)
//   - AzureRM internal/services/cosmos/validate/cosmos.go CosmosEntityName/Throughput/MaxThroughput
//   - go-azure-sdk resource-manager/cosmosdb/2024-08-15/cosmosdb:
//     model_mongodbdatabasecreateupdateproperties.go, model_createupdateoptions.go,
//     model_autoscalesettings.go
//
// Not encoded (deliberate): name/resource_group_name/account_name are envelope-owned;
// throughput/maxThroughput modulo (100/1000) constraints are not expressible as IntRule.
type CosmosDbMongoDatabase struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CosmosDbMongoDatabase)(nil)

func NewCosmosDbMongoDatabase() *CosmosDbMongoDatabase {
	return &CosmosDbMongoDatabase{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/databaseAccounts/mongodbDatabases",
			ApiVersions:  []string{"2024-08-15"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{MinLength: 1, MaxLength: 255, Message: "Cosmos DB entity name must be between 1 and 255 characters"},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.options.throughput", MinValue: azwise.Ptr(int64(400))},
				{PropertyPath: "properties.options.autoscaleSettings.maxThroughput", MinValue: azwise.Ptr(int64(1000))},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.options.throughput"}, // optional+computed; server decides
			},
		},
	}
}

func init() { azwise.Register(NewCosmosDbMongoDatabase()) }
