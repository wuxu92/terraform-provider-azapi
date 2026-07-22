package redisenterprise

import (
	"time"

	"github.com/wuxu92/azwise"
)

// RedisEnterpriseDatabase provides resource knowledge for
// Microsoft.Cache/redisEnterprise/databases (the classic Redis Enterprise database,
// the child resource of a redisEnterprise cluster).
//
// Mirrors azurerm_redis_enterprise_database. The parent cluster
// (Microsoft.Cache/redisEnterprise) is registered separately in
// services/managedredis; only accessPolicyAssignments existed under that ARM path —
// databases was genuinely missing, hence this new file.
//
// Update is only partially supported in AzureRM (linked_database_id can change, most
// other fields are ForceNew). azwise emits the ForceNew fields declaratively.
//
// Sources:
//   - terraform-provider-azurerm internal/services/redisenterprise/redis_enterprise_database_resource.go:53-313
//   - Timeouts (:35-40): Create 30m / Read 5m / Update 30m / Delete 30m
//   - Skipped array-element / semantic constraints (not declarative):
//     module.name (StringInSlice[RedisBloom,RedisTimeSeries,RediSearch,RedisJSON] ->
//     properties.modules[*].name — array-element path); linked_database_id element
//     (databases.ValidateDatabaseID -> properties.geoReplication.linkedDatabases[*] —
//     array-element resource-id validator); cluster_id
//     (redisenterprise.ValidateRedisEnterpriseID -> parent cluster ref, resource-id
//     validator, routed to the azapin customizer).
//   - name is limited to "default" (validate.RedisEnterpriseDatabaseName) — an Azure
//     constraint (redisEnterprise supports a single database named "default"), emitted
//     as a name StringRule.
//   - go-azure-sdk resource-manager/redisenterprise/2024-10-01/databases
//     model_database.go / model_databaseproperties.go / model_module.go /
//     model_databasepropertiesgeoreplication.go / constants.go
type RedisEnterpriseDatabase struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*RedisEnterpriseDatabase)(nil)

// NewRedisEnterpriseDatabase returns knowledge for the redisEnterprise database resource.
func NewRedisEnterpriseDatabase() *RedisEnterpriseDatabase {
	return &RedisEnterpriseDatabase{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cache/redisEnterprise/databases",
			ApiVersions:  []string{"2024-10-01"},
			ForceNew: []azwise.ForceNewRule{
				// client_protocol / clustering_policy / eviction_policy / module / port /
				// linked_database_group_nickname are ForceNew. linked_database_id
				// (properties.geoReplication.linkedDatabases) is intentionally NOT here —
				// it is updatable via the force-unlink flow.
				{PropertyPath: "properties.clientProtocol"},
				{PropertyPath: "properties.clusteringPolicy"},
				{PropertyPath: "properties.evictionPolicy"},
				{PropertyPath: "properties.modules"},
				{PropertyPath: "properties.port"},
				{PropertyPath: "properties.geoReplication.groupNickname"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			DefaultValues: []azwise.DefaultValue{
				// client_protocol Default "Encrypted".
				{PropertyPath: "properties.clientProtocol", Value: "Encrypted"},
				// clustering_policy Default "OSSCluster".
				{PropertyPath: "properties.clusteringPolicy", Value: "OSSCluster"},
				// eviction_policy Default "VolatileLRU".
				{PropertyPath: "properties.evictionPolicy", Value: "VolatileLRU"},
				// port Default 10000.
				{PropertyPath: "properties.port", Value: int64(10000)},
			},
			StringRules: []azwise.StringRule{
				// name limited to "default".
				{
					PropertyPath:  "",
					AllowedValues: []string{"default"},
					Message:       "redisEnterprise database name is currently limited to 'default'",
				},
				// client_protocol: full ARM Protocol enum.
				{
					PropertyPath:  "properties.clientProtocol",
					AllowedValues: []string{"Encrypted", "Plaintext"},
					Message:       "clientProtocol must be Encrypted or Plaintext",
				},
				// clustering_policy: full ARM ClusteringPolicy enum.
				{
					PropertyPath:  "properties.clusteringPolicy",
					AllowedValues: []string{"EnterpriseCluster", "OSSCluster"},
					Message:       "clusteringPolicy must be EnterpriseCluster or OSSCluster",
				},
				// eviction_policy: full ARM EvictionPolicy enum.
				{
					PropertyPath: "properties.evictionPolicy",
					AllowedValues: []string{
						"AllKeysLFU", "AllKeysLRU", "AllKeysRandom", "NoEviction",
						"VolatileLFU", "VolatileLRU", "VolatileRandom", "VolatileTTL",
					},
					Message: "evictionPolicy must be a valid redisEnterprise eviction policy",
				},
			},
			IntRules: []azwise.IntRule{
				// port: IntBetween(0, 65353).
				{
					PropertyPath: "properties.port",
					MinValue:     azwise.Ptr[int64](0),
					MaxValue:     azwise.Ptr[int64](65353),
					Message:      "port must be between 0 and 65353",
				},
			},
			ArrayRules: []azwise.ArrayRule{
				// module MaxItems 4.
				{
					PropertyPath: "properties.modules",
					MaxItems:     4,
					Message:      "at most 4 modules are supported",
				},
				// linked_database_id MaxItems 5.
				{
					PropertyPath: "properties.geoReplication.linkedDatabases",
					MaxItems:     5,
					Message:      "at most 5 linked databases are supported",
				},
			},
			// linked_database_group_nickname RequiredWith linked_database_id.
			RequiredWith: []azwise.RelationalRule{
				{
					Paths: []string{
						"properties.geoReplication.groupNickname",
						"properties.geoReplication.linkedDatabases",
					},
					Message: "groupNickname and linkedDatabases must be set together",
				},
			},
		},
	}
}

func init() { azwise.Register(NewRedisEnterpriseDatabase()) }
