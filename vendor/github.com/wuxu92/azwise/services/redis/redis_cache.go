package redis

import (
	"time"

	"github.com/wuxu92/azwise"
)

// RedisCache provides resource knowledge for Microsoft.Cache/Redis
// (the classic Azure Cache for Redis).
//
// Mirrors azurerm_redis_cache. Note that two things embedded in the AzureRM
// resource are SEPARATE ARM child resources and are intentionally NOT modelled
// here:
//   - patch_schedule -> Microsoft.Cache/Redis/patchSchedules/default (created via a
//     separate PatchSchedulesCreateOrUpdate call), so its day_of_week /
//     maintenance_window / start_hour_utc constraints do not belong in the cache body.
//   - the redis access keys / connection strings are read-only outputs.
//
// Sources:
//   - terraform-provider-azurerm internal/services/redis/redis_cache_resource.go:51-439
//     (schema), :441-577 (create -> RedisCreateParameters), :877-990 (expandRedisConfiguration)
//   - Timeouts (:63-68): Create 180m / Read 5m / Update 180m / Delete 180m
//   - CustomizeDiff (:380-422): sku_name is conditionally ForceNew only on a SKU
//     *downgrade* (skuWeight[old] > skuWeight[new]); it is NOT emitted as a declarative
//     ForceNew rule because forcing replacement on every sku change would break valid
//     upgrades, and azwise cannot evaluate the directional check. (Same rationale as the
//     managedredis cluster knowledge.)
//   - Semantic validators routed to the azapin customizer, not declarative rules:
//     redis_configuration.storage_account_subscription_id (validation.IsUUID ->
//     properties.redisConfiguration.storage-subscription-id, validators.UUID); subnet_id
//     (commonids.ValidateSubnetID -> properties.subnetId, resource-id validator).
//   - go-azure-sdk resource-manager/redis/2024-11-01/redisresources
//     model_rediscreateparameters.go / model_rediscreateproperties.go / model_sku.go /
//     model_rediscommonpropertiesredisconfiguration.go / model_redisproperties.go / constants.go
type RedisCache struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*RedisCache)(nil)

// NewRedisCache returns knowledge for the classic Redis cache resource.
func NewRedisCache() *RedisCache {
	return &RedisCache{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cache/Redis",
			ApiVersions:  []string{"2024-11-01"},
			ForceNew: []azwise.ForceNewRule{
				// zones: commonschema.ZonesMultipleOptionalForceNew() -> top-level
				// zones; zonal placement is immutable.
				{PropertyPath: "zones"},
				// subnet_id (ForceNew) -> properties.subnetId.
				{PropertyPath: "properties.subnetId"},
				// private_static_ip_address (ForceNew, O+C) -> properties.staticIP.
				{PropertyPath: "properties.staticIP"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 180 * time.Minute,
				Read:   5 * time.Minute,
				Update: 180 * time.Minute,
				Delete: 180 * time.Minute,
			},
			// sku is Required (name, family, capacity). location/name are envelope-owned.
			RequiredFields: []string{
				"properties.sku.name",
				"properties.sku.family",
				"properties.sku.capacity",
			},
			DefaultValues: []azwise.DefaultValue{
				// minimum_tls_version Default "1.2".
				{PropertyPath: "properties.minimumTlsVersion", Value: "1.2"},
				// non_ssl_port_enabled Default false -> enableNonSslPort.
				{PropertyPath: "properties.enableNonSslPort", Value: false},
				// public_network_access_enabled Default true -> "Enabled".
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				// redis_version Default "6".
				{PropertyPath: "properties.redisVersion", Value: "6"},
				// access_keys_authentication_enabled Default true maps to the inverse
				// disableAccessKeyAuthentication = !true = false.
				{PropertyPath: "properties.disableAccessKeyAuthentication", Value: false},
				// redis_configuration.maxmemory_policy Default "volatile-lru".
				{PropertyPath: "properties.redisConfiguration.maxmemory-policy", Value: "volatile-lru"},
			},
			StringRules: []azwise.StringRule{
				// family: full ARM SkuFamily enum.
				{
					PropertyPath:  "properties.sku.family",
					AllowedValues: []string{"C", "P"},
					Message:       "sku family must be C (Basic/Standard) or P (Premium)",
				},
				// sku_name: full ARM SkuName enum.
				{
					PropertyPath:  "properties.sku.name",
					AllowedValues: []string{"Basic", "Standard", "Premium"},
					Message:       "sku name must be Basic, Standard or Premium",
				},
				// minimum_tls_version: full ARM TlsVersion enum (AzureRM restricts new
				// caches to 1.2, but the ARM type accepts the legacy values and AzAPI
				// sends raw values).
				{
					PropertyPath:  "properties.minimumTlsVersion",
					AllowedValues: []string{"1.0", "1.1", "1.2"},
					Message:       "minimumTlsVersion must be 1.0, 1.1 or 1.2",
				},
				// public_network_access enum.
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "publicNetworkAccess must be Enabled or Disabled",
				},
				// redis_configuration.maxmemory_policy: validate.MaxMemoryPolicy.
				// ARM field is a free string; AzureRM restricts to this set.
				{
					PropertyPath: "properties.redisConfiguration.maxmemory-policy",
					AllowedValues: []string{
						"allkeys-lfu", "allkeys-lru", "allkeys-random", "noeviction",
						"volatile-lru", "volatile-lfu", "volatile-random", "volatile-ttl",
					},
					Message: "maxmemory-policy must be a valid Redis eviction policy",
				},
				// redis_configuration.data_persistence_authentication_method ->
				// preferred-data-persistence-auth-method (StringInSlice).
				{
					PropertyPath:  "properties.redisConfiguration.preferred-data-persistence-auth-method",
					AllowedValues: []string{"SAS", "ManagedIdentity"},
					Message:       "data persistence auth method must be SAS or ManagedIdentity",
				},
				// redis_configuration.rdb_backup_frequency: validate.CacheBackupFrequency.
				// TF is an int but the ARM field rdb-backup-frequency is a string, so this
				// is a StringRule over the discrete allowed values (not an IntRule).
				{
					PropertyPath:  "properties.redisConfiguration.rdb-backup-frequency",
					AllowedValues: []string{"15", "30", "60", "360", "720", "1440"},
					Message:       "rdb-backup-frequency must be one of 15, 30, 60, 360, 720 or 1440",
				},
			},
			IntRules: []azwise.IntRule{
				// replicas_per_master: IntBetween(1, 3) -> properties.replicasPerMaster.
				{
					PropertyPath: "properties.replicasPerMaster",
					MinValue:     azwise.Ptr[int64](1),
					MaxValue:     azwise.Ptr[int64](3),
					Message:      "replicasPerMaster must be between 1 and 3",
				},
				// replicas_per_primary: IntBetween(1, 3) -> properties.replicasPerPrimary.
				{
					PropertyPath: "properties.replicasPerPrimary",
					MinValue:     azwise.Ptr[int64](1),
					MaxValue:     azwise.Ptr[int64](3),
					Message:      "replicasPerPrimary must be between 1 and 3",
				},
			},
			SensitiveFields: []string{
				"properties.redisConfiguration.rdb-storage-connection-string",
				"properties.redisConfiguration.aof-storage-connection-string-0",
				"properties.redisConfiguration.aof-storage-connection-string-1",
				"properties.accessKeys",
			},
			// Read-only: present in RedisProperties (GET) but absent from
			// RedisCreateProperties (create model).
			ComputedFields: []string{
				"properties.hostName",
				"properties.port",
				"properties.sslPort",
				"properties.accessKeys",
				"properties.provisioningState",
				"properties.instances",
				"properties.linkedServers",
				"properties.privateEndpointConnections",
			},
		},
	}
}

func init() { azwise.Register(NewRedisCache()) }
