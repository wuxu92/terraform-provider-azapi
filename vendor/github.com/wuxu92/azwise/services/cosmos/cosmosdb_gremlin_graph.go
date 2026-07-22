package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CosmosDbGremlinGraph provides resource knowledge for
// Microsoft.DocumentDB/databaseAccounts/gremlinDatabases/graphs
// (azurerm_cosmosdb_gremlin_graph).
//
// It overrides CheckForceNew for the value-conditional replacement AzureRM expresses
// via CustomizeDiff (cosmosdb_gremlin_graph_resource.go:189-194): analytical_storage_ttl
// cannot be disabled in place.
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_gremlin_graph_resource.go
//     :46-51 (timeouts 30/5/30/30), :53-187 (schema: name CosmosEntityName,
//     analytical_storage_ttl IntBetween(-1,MaxInt32)!=0, default_ttl, throughput,
//     partition_key_path ForceNew, partition_key_version IntBetween(1,2) ForceNew,
//     index_policy indexing_mode enum, conflict_resolution_policy, unique_key ForceNew),
//     :189-194 (CustomizeDiff conditional ForceNew), :198-274/:454-531 (expand under
//     properties.resource / properties.options)
//   - go-azure-sdk resource-manager/cosmosdb/2024-08-15/cosmosdb:
//     model_gremlingraphcreateupdateproperties.go, model_gremlingraphresource.go,
//     model_containerpartitionkey.go, model_indexingpolicy.go, model_conflictresolutionpolicy.go,
//     constants.go PossibleValuesForIndexingMode/ConflictResolutionMode
//
// Not encoded (deliberate):
//   - name / account_name / database_name are envelope-owned (URL segments).
//   - partition_key_path expands to properties.resource.partitionKey.paths[0]
//     (one-to-many), so ForceNew/required target the whole paths array.
//   - unique_key[*].paths and index_policy composite/included/excluded/spatial paths
//     are array-element paths; azwise cannot lower a rule through an array element.
//     The unique_key policy as a whole is ForceNew.
type CosmosDbGremlinGraph struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CosmosDbGremlinGraph)(nil)

// CheckForceNew extends BaseKnowledge with AzureRM's value-conditional replacement
// (cosmosdb_gremlin_graph_resource.go:189-194): analytical_storage_ttl cannot be
// disabled (set to 0) once enabled.
func (s *CosmosDbGremlinGraph) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
	if s.BaseKnowledge.CheckForceNew(oldBody, newBody) {
		return true
	}

	const ttlPath = "properties.resource.analyticalStorageTtl"
	oldTTL, oldOK := cosmosIntValue(oldBody, ttlPath)
	newTTL, newOK := cosmosIntValue(newBody, ttlPath)
	if oldOK && newOK && (oldTTL == -1 || oldTTL >= 1) && newTTL == 0 {
		return true
	}

	return false
}

func NewCosmosDbGremlinGraph() *CosmosDbGremlinGraph {
	return &CosmosDbGremlinGraph{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/databaseAccounts/gremlinDatabases/graphs",
			ApiVersions:  []string{"2024-08-15"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.resource.partitionKey.paths"},
				{PropertyPath: "properties.resource.partitionKey.version"},
				{PropertyPath: "properties.resource.uniqueKeyPolicy"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{MinLength: 1, MaxLength: 255, Message: "Cosmos DB entity name must be between 1 and 255 characters"},
				{PropertyPath: "properties.resource.indexingPolicy.indexingMode", AllowedValues: []string{"consistent", "lazy", "none"}},
				{PropertyPath: "properties.resource.conflictResolutionPolicy.mode", AllowedValues: []string{"Custom", "LastWriterWins"}},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.resource.partitionKey.version", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(2))},
				{PropertyPath: "properties.resource.analyticalStorageTtl", MinValue: azwise.Ptr(int64(-1))},
				{PropertyPath: "properties.options.throughput", MinValue: azwise.Ptr(int64(400))},
				{PropertyPath: "properties.options.autoscaleSettings.maxThroughput", MinValue: azwise.Ptr(int64(1000))},
			},
			RequiredFields: []string{
				"properties.resource.partitionKey.paths",
			},
		},
	}
}

func init() { azwise.Register(NewCosmosDbGremlinGraph()) }
