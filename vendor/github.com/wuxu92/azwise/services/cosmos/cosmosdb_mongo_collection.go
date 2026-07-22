package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CosmosDbMongoCollection provides resource knowledge for
// Microsoft.DocumentDB/databaseAccounts/mongodbDatabases/collections
// (azurerm_cosmosdb_mongo_collection).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_mongo_collection_resource.go
//     :44-49 (timeouts 30/5/30/30), :51-146 (schema: name CosmosEntityName,
//     shard_key ForceNew, default_ttl_seconds, analytical_storage_ttl IntAtLeast(-1),
//     throughput, index, system_indexes computed), :168-206 (expand under
//     properties.resource / properties.options)
//   - AzureRM internal/services/cosmos/validate/cosmos.go CosmosEntityName/Throughput/MaxThroughput
//   - go-azure-sdk resource-manager/cosmosdb/2024-08-15/cosmosdb:
//     model_mongodbcollectioncreateupdateproperties.go, model_mongodbcollectionresource.go,
//     model_mongoindex.go, model_createupdateoptions.go
//
// Not encoded (deliberate):
//   - name / account_name / database_name are envelope-owned (URL segments).
//   - shard_key is a map[string]string (properties.resource.shardKey); it is ForceNew
//     but its keys/values are a map, not validatable declaratively.
//   - default_ttl_seconds is not a scalar ARM property: AzureRM expands it into a TTL
//     index entry inside properties.resource.indexes[*], so it has no single ARM path.
//   - index[*].keys / index[*].unique and system_indexes are array-element or
//     response-only index paths; azwise cannot lower a rule through an array element.
type CosmosDbMongoCollection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CosmosDbMongoCollection)(nil)

func NewCosmosDbMongoCollection() *CosmosDbMongoCollection {
	return &CosmosDbMongoCollection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/databaseAccounts/mongodbDatabases/collections",
			ApiVersions:  []string{"2024-08-15"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.resource.shardKey"},
			},
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
				{PropertyPath: "properties.resource.analyticalStorageTtl", MinValue: azwise.Ptr(int64(-1))},
				{PropertyPath: "properties.options.throughput", MinValue: azwise.Ptr(int64(400))},
				{PropertyPath: "properties.options.autoscaleSettings.maxThroughput", MinValue: azwise.Ptr(int64(1000))},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.options.throughput"}, // optional+computed; server decides
			},
		},
	}
}

func init() { azwise.Register(NewCosmosDbMongoCollection()) }
