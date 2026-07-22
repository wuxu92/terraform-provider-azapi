package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CosmosDbSqlContainer provides resource knowledge for
// Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers
// (azurerm_cosmosdb_sql_container).
//
// It overrides CheckForceNew for the two value-conditional replacements AzureRM
// expresses via CustomizeDiff (cosmosdb_sql_container_resource.go:147-157):
// analytical_storage_ttl cannot be disabled in place, and partition_key_version can
// only be updated from unset to 1.
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_sql_container_resource.go
//     :44-49 (timeouts 30/5/30/30), :51-145 (schema: name CosmosEntityName,
//     partition_key_paths ForceNew, partition_key_kind PartitionKind enum ForceNew
//     default Hash, partition_key_version IntBetween(1,2), conflict_resolution_policy,
//     throughput, analytical_storage_ttl IntAtLeast(-1), default_ttl IntAtLeast(-1),
//     unique_key ForceNew, indexing_policy), :147-157 (CustomizeDiff conditional ForceNew),
//     :188-233 (expand under properties.resource / properties.options)
//   - AzureRM internal/services/cosmos/common/conflict_resolution_policy.go (mode enum),
//     common/indexing_policy.go (indexing_mode enum)
//   - go-azure-sdk resource-manager/cosmosdb/2024-08-15/cosmosdb:
//     model_sqlcontainercreateupdateproperties.go, model_sqlcontainerresource.go,
//     model_containerpartitionkey.go, model_indexingpolicy.go, model_conflictresolutionpolicy.go,
//     constants.go PossibleValuesForPartitionKind/IndexingMode/ConflictResolutionMode
//
// Not encoded (deliberate):
//   - name / account_name / database_name are envelope-owned (URL segments).
//   - partition_key_paths[*], unique_key[*].paths, indexing_policy composite/included/
//     excluded/spatial paths are array-element paths; azwise cannot lower a rule
//     through an array element. The unique_key policy as a whole is ForceNew.
type CosmosDbSqlContainer struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CosmosDbSqlContainer)(nil)

// CheckForceNew extends BaseKnowledge with AzureRM's value-conditional replacements
// (cosmosdb_sql_container_resource.go:147-157): analytical_storage_ttl cannot be
// disabled once enabled, and partition_key_version may only move from unset (0) to 1.
func (s *CosmosDbSqlContainer) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
	if s.BaseKnowledge.CheckForceNew(oldBody, newBody) {
		return true
	}

	const ttlPath = "properties.resource.analyticalStorageTtl"
	oldTTL, oldOK := cosmosIntValue(oldBody, ttlPath)
	newTTL, newOK := cosmosIntValue(newBody, ttlPath)
	if oldOK && newOK && (oldTTL == -1 || oldTTL > 0) && newTTL == 0 {
		return true
	}

	const versionPath = "properties.resource.partitionKey.version"
	oldVer, oldVerOK := cosmosIntValue(oldBody, versionPath)
	newVer, newVerOK := cosmosIntValue(newBody, versionPath)
	if (oldVerOK || newVerOK) && (oldVer != 0 || newVer != 1) {
		return true
	}

	return false
}

func NewCosmosDbSqlContainer() *CosmosDbSqlContainer {
	return &CosmosDbSqlContainer{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers",
			ApiVersions:  []string{"2024-08-15"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.resource.partitionKey.paths"},
				{PropertyPath: "properties.resource.partitionKey.kind"},
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
				{PropertyPath: "properties.resource.partitionKey.kind", AllowedValues: []string{"Hash", "MultiHash", "Range"}},
				{PropertyPath: "properties.resource.conflictResolutionPolicy.mode", AllowedValues: []string{"Custom", "LastWriterWins"}},
				{PropertyPath: "properties.resource.indexingPolicy.indexingMode", AllowedValues: []string{"consistent", "lazy", "none"}},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.resource.partitionKey.version", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(2))},
				{PropertyPath: "properties.resource.analyticalStorageTtl", MinValue: azwise.Ptr(int64(-1))},
				{PropertyPath: "properties.resource.defaultTtl", MinValue: azwise.Ptr(int64(-1))},
				{PropertyPath: "properties.options.throughput", MinValue: azwise.Ptr(int64(400))},
				{PropertyPath: "properties.options.autoscaleSettings.maxThroughput", MinValue: azwise.Ptr(int64(1000))},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.resource.partitionKey.kind", Value: "Hash"},
			},
			RequiredFields: []string{
				"properties.resource.partitionKey.paths",
			},
		},
	}
}

func init() { azwise.Register(NewCosmosDbSqlContainer()) }

// cosmosIntValue returns the integer at path, or (0,false) when absent/non-numeric.
// JSON bodies decode numbers as float64, so both float64 and int/int64 are handled.
func cosmosIntValue(body map[string]interface{}, path string) (int64, bool) {
	switch v := azwise.ExtractNestedValue(body, path).(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case int:
		return int64(v), true
	default:
		return 0, false
	}
}
