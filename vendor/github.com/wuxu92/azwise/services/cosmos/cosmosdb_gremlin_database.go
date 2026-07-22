package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CosmosDbGremlinDatabase provides resource knowledge for
// Microsoft.DocumentDB/databaseAccounts/gremlinDatabases (azurerm_cosmosdb_gremlin_database).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_gremlin_database_resource.go
//     :40-45 (timeouts 30/5/30/30), :47-72 (schema: name CosmosEntityName,
//     throughput CosmosThroughput, autoscale_settings), :96-113 (expand)
//   - AzureRM internal/services/cosmos/validate/cosmos.go CosmosEntityName/Throughput/MaxThroughput
//   - go-azure-sdk resource-manager/cosmosdb/2024-08-15/cosmosdb:
//     model_gremlindatabasecreateupdateproperties.go, model_createupdateoptions.go
//
// Not encoded (deliberate): name/resource_group_name/account_name are envelope-owned;
// throughput/maxThroughput modulo (100/1000) constraints are not expressible as IntRule.
type CosmosDbGremlinDatabase struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CosmosDbGremlinDatabase)(nil)

func NewCosmosDbGremlinDatabase() *CosmosDbGremlinDatabase {
	return &CosmosDbGremlinDatabase{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/databaseAccounts/gremlinDatabases",
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

func init() { azwise.Register(NewCosmosDbGremlinDatabase()) }
