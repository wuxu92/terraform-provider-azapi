package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PostgreSQLCluster provides resource knowledge for
// Microsoft.DBforPostgreSQL/serverGroupsv2 (azurerm_cosmosdb_postgresql_cluster).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_postgresql_cluster_resource.go
//     :88-314  (Arguments schema: name StringLenBetween(1,260), node_count
//     IntBetween(0,20)\1 Required, administrator_login_password StringLenBetween(8,256)
//     Sensitive, citus_version/coordinator_server_edition/node_server_edition/sql_version
//     enums, coordinator_storage_quota_in_mb & node_storage_quota_in_mb IntBetween/DivBy,
//     coordinator_vcore_count & node_vcores IntInSlice, maintenance_window IntBetween,
//     point_in_time_in_utc/source_location/source_resource_id ForceNew+RequiredWith,
//     coordinator_public_ip_access_enabled/ha_enabled/node_public_ip_access_enabled defaults)
//     :341-443 (create mapping to clusters.Cluster / clusters.ClusterProperties)
//     :343,447,556,627 (timeouts: create/update/delete 3h, read 5m)
//   - go-azure-sdk resource-manager/cosmosdb/../postgresqlhsc/2022-11-08/clusters:
//     model_clusterproperties.go (properties.* body paths),
//     model_maintenancewindow.go (properties.maintenanceWindow.*),
//     constants.go:14-16 (Microsoft.DBforPostgreSQL/serverGroupsv2)
//
// Not encoded (deliberate):
//   - coordinator_vcore_count (properties.coordinatorVCores) IntInSlice[1,2,4,8,16,32,64,96]
//     and node_vcores (properties.nodeVCores) IntInSlice[1,2,4,8,16,32,64,96,104] are
//     discrete integer enums; IntRule only expresses min/max bounds, so only the enclosing
//     range is emitted (a loose guard) with the allowed set noted in the message.
//   - coordinator_storage_quota_in_mb / node_storage_quota_in_mb additionally require
//     divisibility by 1024 (IntDivisibleBy); only the range is expressible.
//   - node_count additionally excludes 1 (IntNotInSlice); only the 0..20 range is expressible.
//   - administrator_login_password is only conditionally Required (when source_resource_id
//     is unset — resource :415-416); not a universal RequiredField.
//   - properties.administratorLogin / earliestRestoreTime / serverNames / provisioningState
//     / state / readReplicas / privateEndpointConnections are server-populated but present
//     in the shared ClusterProperties Create model, so they are NOT in ComputedFields.
//   - preferred_primary_zone (properties.preferredPrimaryZone) is a free-form availability
//     zone string with no enum/regex constraint.
type PostgreSQLCluster struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PostgreSQLCluster)(nil)

// NewPostgreSQLCluster returns knowledge for the serverGroupsv2 resource.
func NewPostgreSQLCluster() *PostgreSQLCluster {
	return &PostgreSQLCluster{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforPostgreSQL/serverGroupsv2",
			ApiVersions:  []string{"2022-11-08"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.pointInTimeUTC"},
				{PropertyPath: "properties.sourceLocation"},
				{PropertyPath: "properties.sourceResourceId"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 3 * time.Hour,
				Read:   5 * time.Minute,
				Update: 3 * time.Hour,
				Delete: 3 * time.Hour,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					MinLength:    1,
					MaxLength:    260,
					Message:      "name must be between 1 and 260 characters",
				},
				{
					PropertyPath: "properties.administratorLoginPassword",
					MinLength:    8,
					MaxLength:    256,
					Message:      "administrator_login_password must be between 8 and 256 characters",
				},
				{
					// AzureRM restricts citus_version via StringInSlice; the SDK field is a
					// free-form *string (no enum type).
					PropertyPath:  "properties.citusVersion",
					AllowedValues: []string{"8.3", "9.0", "9.1", "9.2", "9.3", "9.4", "9.5", "10.0", "10.1", "10.2", "11.0", "11.1", "11.2", "11.3", "12.1"},
					Message:       "citus_version must be a supported Citus version",
				},
				{
					PropertyPath:  "properties.coordinatorServerEdition",
					AllowedValues: []string{"BurstableGeneralPurpose", "BurstableMemoryOptimized", "GeneralPurpose", "MemoryOptimized"},
					Message:       "coordinator_server_edition must be one of Burstable*/GeneralPurpose/MemoryOptimized",
				},
				{
					PropertyPath:  "properties.nodeServerEdition",
					AllowedValues: []string{"BurstableGeneralPurpose", "BurstableMemoryOptimized", "GeneralPurpose", "MemoryOptimized"},
					Message:       "node_server_edition must be one of Burstable*/GeneralPurpose/MemoryOptimized",
				},
				{
					// AzureRM restricts sql_version via StringInSlice; the SDK field is a
					// free-form *string (no enum type).
					PropertyPath:  "properties.postgresqlVersion",
					AllowedValues: []string{"11", "12", "13", "14", "15", "16"},
					Message:       "sql_version must be one of 11, 12, 13, 14, 15, 16",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.nodeCount",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(20)),
					Message:      "node_count must be between 0 and 20 (excluding 1)",
				},
				{
					PropertyPath: "properties.coordinatorStorageQuotaInMb",
					MinValue:     azwise.Ptr(int64(32768)),
					MaxValue:     azwise.Ptr(int64(16777216)),
					Message:      "coordinator_storage_quota_in_mb must be between 32768 and 16777216 (multiple of 1024)",
				},
				{
					PropertyPath: "properties.coordinatorVCores",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(96)),
					Message:      "coordinator_vcore_count must be one of 1, 2, 4, 8, 16, 32, 64, 96",
				},
				{
					PropertyPath: "properties.nodeStorageQuotaInMb",
					MinValue:     azwise.Ptr(int64(32768)),
					MaxValue:     azwise.Ptr(int64(16777216)),
					Message:      "node_storage_quota_in_mb must be between 32768 and 16777216 (multiple of 1024)",
				},
				{
					PropertyPath: "properties.nodeVCores",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(104)),
					Message:      "node_vcores must be one of 1, 2, 4, 8, 16, 32, 64, 96, 104",
				},
				{
					PropertyPath: "properties.maintenanceWindow.dayOfWeek",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(6)),
					Message:      "maintenance_window day_of_week must be between 0 and 6",
				},
				{
					PropertyPath: "properties.maintenanceWindow.startHour",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(23)),
					Message:      "maintenance_window start_hour must be between 0 and 23",
				},
				{
					PropertyPath: "properties.maintenanceWindow.startMinute",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(59)),
					Message:      "maintenance_window start_minute must be between 0 and 59",
				},
			},
			SensitiveFields: []string{"properties.administratorLoginPassword"},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.coordinatorEnablePublicIpAccess", Value: true},
				{PropertyPath: "properties.coordinatorServerEdition", Value: "GeneralPurpose"},
				{PropertyPath: "properties.enableHa", Value: false},
				{PropertyPath: "properties.nodeEnablePublicIpAccess", Value: false},
				{PropertyPath: "properties.nodeServerEdition", Value: "MemoryOptimized"},
			},
			RequiredFields: []string{"properties.nodeCount"},
			RequiredWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.pointInTimeUTC", "properties.sourceLocation", "properties.sourceResourceId"},
					Message: "point_in_time_in_utc requires source_location and source_resource_id",
				},
				{
					Paths:   []string{"properties.sourceLocation", "properties.sourceResourceId"},
					Message: "source_location requires source_resource_id",
				},
				{
					Paths:   []string{"properties.sourceResourceId", "properties.sourceLocation"},
					Message: "source_resource_id requires source_location",
				},
			},
		},
	}
}

func init() { azwise.Register(NewPostgreSQLCluster()) }
