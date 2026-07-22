package postgres

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PostgresqlFlexibleServer provides resource knowledge for
// Microsoft.DBforPostgreSQL/flexibleServers (azurerm_postgresql_flexible_server).
//
// Sources:
//   - AzureRM internal/services/postgres/postgresql_flexible_server_resource.go
//     :46-575  (Arguments schema: name FlexibleServerName len<=63 regex, sku_name,
//     storage_mb IntInSlice, storage_tier enum, version enum, create_mode enum,
//     backup_retention_days IntBetween(7,35), geo_redundant_backup_enabled ForceNew,
//     high_availability.mode enum, maintenance_window IntBetween ranges, cluster ForceNew,
//     customer_managed_key ForceNew, delegated_subnet_id/point_in_time/source_server_id ForceNew)
//     :53-58   (timeouts: create/update/delete 1h, read 5m)
//     :389-548 (CustomizeDiff: conditional ForceNew — see "Not encoded")
//     :682-694 (create mapping to servers.Server / servers.ServerProperties)
//   - go-azure-sdk resource-manager/postgresql/2025-08-01/servers:
//     model_server.go (properties/sku/identity/tags/location envelope),
//     model_serverproperties.go, model_storage.go, model_backup.go,
//     model_highavailability.go, model_dataencryption.go, model_authconfig.go,
//     model_network.go, model_cluster.go, model_maintenancewindow.go,
//     constants.go (enum values), id_flexibleserver.go:117 (staticFlexibleServers)
//
// Not encoded (deliberate):
//   - storage_mb (validation.IntInSlice on the TF MB value) maps to
//     properties.storage.storageSizeGB, but AzureRM converts MB->GB (mb/1024) before
//     sending; the ARM field is in GB while the enum is in MB, so no clean IntRule.
//   - Conditional ForceNew from CustomizeDiff (not universally ForceNew, so omitted to
//     avoid false replacement): version downgrade (properties.version),
//     administrator_login when previously set (properties.administratorLogin),
//     storage_mb downgrade (properties.storage.storageSizeGB), identity.0.type
//     narrowing, cluster.0.size downgrade (properties.cluster.clusterSize).
//   - authentication.tenant_id (properties.authConfig.tenantId) is validation.IsUUID —
//     a semantic validator that belongs in an azapin customizer (validators.UUID),
//     not expressible as a declarative StringRule.
//   - customer_managed_key.key_vault_key_id / geo_backup_key_vault_key_id are Key Vault
//     nested-item URIs mapped to properties.dataEncryption.primaryKeyURI /
//     geoBackupKeyURI (semantic KV validators; not declarative).
//   - sku_name/version/administrator_login/administrator_password are only conditionally
//     Required (create_mode == Default); not universal RequiredFields.
type PostgresqlFlexibleServer struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PostgresqlFlexibleServer)(nil)

// NewPostgresqlFlexibleServer returns knowledge for the flexibleServers resource.
func NewPostgresqlFlexibleServer() *PostgresqlFlexibleServer {
	return &PostgresqlFlexibleServer{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforPostgreSQL/flexibleServers",
			ApiVersions:  []string{"2025-08-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.network.delegatedSubnetResourceId"},
				{PropertyPath: "properties.pointInTimeUTC"},
				{PropertyPath: "properties.sourceServerResourceId"},
				{PropertyPath: "properties.backup.geoRedundantBackup"},
				{PropertyPath: "properties.dataEncryption"},
				{PropertyPath: "properties.cluster"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 1 * time.Hour,
				Read:   5 * time.Minute,
				Update: 1 * time.Hour,
				Delete: 1 * time.Hour,
			},
			SensitiveFields: []string{
				"properties.administratorLoginPassword",
			},
			ComputedFields: []string{
				"properties.fullyQualifiedDomainName",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					MinLength:    1,
					MaxLength:    63,
					Regex:        `^[a-z0-9]([a-z0-9-]+[a-z0-9])?$`,
					Message:      "server name must be 1-63 characters: lowercase letters, numbers and '-', not starting or ending with '-'",
				},
				{
					PropertyPath:  "properties.storage.tier",
					AllowedValues: []string{"P1", "P2", "P3", "P4", "P6", "P10", "P15", "P20", "P30", "P40", "P50", "P60", "P70", "P80"},
					Message:       "storage tier must be a valid AzureManagedDiskPerformanceTier",
				},
				{
					PropertyPath:  "properties.version",
					AllowedValues: []string{"11", "12", "13", "14", "15", "16", "17", "18"},
					Message:       "version must be a supported PostgreSQL major version",
				},
				{
					PropertyPath:  "properties.createMode",
					AllowedValues: []string{"Create", "Default", "GeoRestore", "PointInTimeRestore", "Replica", "ReviveDropped", "Update"},
					Message:       "create_mode must be a valid CreateMode",
				},
				{
					PropertyPath:  "properties.highAvailability.mode",
					AllowedValues: []string{"Disabled", "SameZone", "ZoneRedundant"},
					Message:       "high_availability.mode must be a valid PostgreSqlFlexibleServerHighAvailabilityMode",
				},
				{
					PropertyPath:  "properties.replicationRole",
					AllowedValues: []string{"AsyncReplica", "GeoAsyncReplica", "None", "Primary"},
					Message:       "replication_role must be a valid ReplicationRole",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.backup.backupRetentionDays",
					MinValue:     azwise.Ptr(int64(7)),
					MaxValue:     azwise.Ptr(int64(35)),
					Message:      "backup_retention_days must be between 7 and 35",
				},
				{
					PropertyPath: "properties.cluster.clusterSize",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(20)),
					Message:      "cluster size must be between 1 and 20",
				},
				{
					PropertyPath: "properties.maintenanceWindow.dayOfWeek",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(6)),
					Message:      "maintenance_window.day_of_week must be between 0 and 6",
				},
				{
					PropertyPath: "properties.maintenanceWindow.startHour",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(23)),
					Message:      "maintenance_window.start_hour must be between 0 and 23",
				},
				{
					PropertyPath: "properties.maintenanceWindow.startMinute",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(59)),
					Message:      "maintenance_window.start_minute must be between 0 and 59",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.authConfig.activeDirectoryAuth", Value: "Disabled"},
				{PropertyPath: "properties.authConfig.passwordAuth", Value: "Enabled"},
				{PropertyPath: "properties.storage.autoGrow", Value: "Disabled"},
				{PropertyPath: "properties.backup.geoRedundantBackup", Value: "Disabled"},
				{PropertyPath: "properties.network.publicNetworkAccess", Value: "Enabled"},
				{PropertyPath: "properties.cluster.defaultDatabaseName", Value: "postgres"},
			},
		},
	}
}

func init() { azwise.Register(NewPostgresqlFlexibleServer()) }
