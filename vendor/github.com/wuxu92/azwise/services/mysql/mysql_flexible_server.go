package mysql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MySQLFlexibleServer provides resource knowledge for Microsoft.DBforMySQL/flexibleServers.
//
// Contributing Terraform resource: azurerm_mysql_flexible_server.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mysql/mysql_flexible_server_resource.go
//     (schema L45-435, Create body L437-599, storage.0.size_gb shrink ForceNew CustomizeDiff L343-347,
//     expand network L936-952, storage L972-1008, backup L1035-1050, sku L1052-1074)
//   - terraform-provider-azurerm internal/services/mysql/validate/flexible_server_name.go (FlexibleServerName L10-32),
//     flexible_server_administrator_login.go (FlexibleServerAdministratorLogin L10-38),
//     flexible_server_administrator_password.go (FlexibleServerAdministratorPassword L10-42),
//     flexible_server_sku_name.go (FlexibleServerSkuName L10-24), flexible_server_version.go (8.4 L7-9)
//   - go-azure-sdk resource-manager/mysql/2023-12-30/servers:
//     model_server.go, model_serverproperties.go, model_storage.go, model_backup.go,
//     model_network.go, model_highavailability.go, model_maintenancewindow.go,
//     model_dataencryption.go, model_mysqlserversku.go, constants.go (enums)
//
// Notes:
//   - name/location/resource_group_name/tags/identity are envelope-owned; not emitted as body rules.
//   - administrator_password (properties.administratorLoginPassword) has a complexity requirement
//     (3 of 4 char categories, len 8-128) that cannot be expressed declaratively; only the length is
//     captured via a StringRule and the field is listed as sensitive.
//   - administrator_login also forbids a reserved-name set (azure_superuser/admin/administrator/root/
//     guest/public) which is not expressible as a regex here; only the character regex + length is kept.
//   - sku_name maps one-to-many onto sku.name + sku.tier (tier derived from the "B_"/"GP_"/"MO_" prefix);
//     the compound AzureRM regex does not target a single ARM field, so no declarative rule is emitted.
//   - customer_managed_key maps onto properties.dataEncryption (primaryKeyURI/primaryUserAssignedIdentityId/
//     geoBackupKeyURI/geoBackupUserAssignedIdentityId); the key-vault URI validators are semantic and not
//     expressed here.
//   - maintenance_window is applied via a follow-up Update call (create cannot set it), but it lives on the
//     same ARM type so its ranges/defaults are still emitted.
//   - high_availability.mode, public_network_access, replication_role and version use the full ARM SDK
//     enum sets (AzAPI sends raw ARM values); AzureRM restricts the settable subset.
//   - storage.0.size_gb (properties.storage.storageSizeGB) is ForceNew only when shrinking; handled in
//     CheckForceNew below, not as an unconditional declarative rule.
type MySQLFlexibleServer struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MySQLFlexibleServer)(nil)

// NewMySQLFlexibleServer returns knowledge for the flexibleServers resource.
func NewMySQLFlexibleServer() *MySQLFlexibleServer {
	return &MySQLFlexibleServer{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforMySQL/flexibleServers",
			ApiVersions:  []string{"2023-12-30"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 2 * time.Hour,
				Read:   5 * time.Minute,
				Update: 2 * time.Hour,
				Delete: 1 * time.Hour,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.administratorLogin"},
				{PropertyPath: "properties.createMode"},
				{PropertyPath: "properties.network.delegatedSubnetResourceId"},
				{PropertyPath: "properties.network.privateDnsZoneResourceId"},
				{PropertyPath: "properties.backup.geoRedundantBackup"},
				{PropertyPath: "properties.restorePointInTime"},
				{PropertyPath: "properties.sourceServerResourceId"},
			},
			SensitiveFields: []string{
				"properties.administratorLoginPassword",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[a-z0-9]([a-z0-9-]+[a-z0-9])?$`,
					MinLength:    3,
					MaxLength:    63,
					Message:      "name must be 3-63 characters, contain only lowercase letters, numbers and '-', and not start or end with '-'",
				},
				{
					PropertyPath: "properties.administratorLogin",
					Regex:        `^[a-zA-Z0-9_]*$`,
					MinLength:    1,
					MaxLength:    32,
					Message:      "administrator_login must be 1-32 characters containing only letters, numbers or '_'",
				},
				{
					PropertyPath:  "properties.createMode",
					AllowedValues: []string{"Default", "GeoRestore", "PointInTimeRestore", "Replica"},
					Message:       "create_mode must be a valid CreateMode value",
				},
				{
					PropertyPath:  "properties.highAvailability.mode",
					AllowedValues: []string{"Disabled", "SameZone", "ZoneRedundant"},
					Message:       "high_availability.mode must be a valid HighAvailabilityMode value",
				},
				{
					PropertyPath:  "properties.network.publicNetworkAccess",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "public_network_access must be Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.replicationRole",
					AllowedValues: []string{"None", "Replica", "Source"},
					Message:       "replication_role must be a valid ReplicationRole value",
				},
				{
					PropertyPath:  "properties.version",
					AllowedValues: []string{"5.7", "8.0.21", "8.4"},
					Message:       "version must be a supported MySQL server version",
				},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.backup.backupRetentionDays", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(35))},
				{PropertyPath: "properties.maintenanceWindow.dayOfWeek", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(6))},
				{PropertyPath: "properties.maintenanceWindow.startHour", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(23))},
				{PropertyPath: "properties.maintenanceWindow.startMinute", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(59))},
				{PropertyPath: "properties.storage.iops", MinValue: azwise.Ptr(int64(360)), MaxValue: azwise.Ptr(int64(48000))},
				{PropertyPath: "properties.storage.storageSizeGB", MinValue: azwise.Ptr(int64(20)), MaxValue: azwise.Ptr(int64(16384))},
			},
			ComputedFields: []string{
				"properties.fullyQualifiedDomainName",
				"properties.replicaCapacity",
				"properties.state",
				"properties.backup.earliestRestoreDate",
				"properties.highAvailability.state",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.backup.backupRetentionDays", Value: 7},
				{PropertyPath: "properties.backup.geoRedundantBackup", Value: "Disabled"},
				{PropertyPath: "properties.storage.autoGrow", Value: "Enabled"},
				{PropertyPath: "properties.storage.logOnDisk", Value: "Disabled"},
				{PropertyPath: "properties.storage.autoIoScaling", Value: "Disabled"},
				{PropertyPath: "properties.maintenanceWindow.dayOfWeek", Value: 0},
				{PropertyPath: "properties.maintenanceWindow.startHour", Value: 0},
				{PropertyPath: "properties.maintenanceWindow.startMinute", Value: 0},
			},
		},
	}
}

// CheckForceNew extends the declarative check with the conditional storage-shrink rule: reducing
// properties.storage.storageSizeGB forces replacement, while growing it does not (mirrors the AzureRM
// CustomizeDiff ForceNewIfChange in mysql_flexible_server_resource.go L343-347).
func (s *MySQLFlexibleServer) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
	if s.BaseKnowledge.CheckForceNew(oldBody, newBody) {
		return true
	}

	oldSize, oldOK := storageSizeGB(oldBody)
	newSize, newOK := storageSizeGB(newBody)
	if oldOK && newOK && newSize < oldSize {
		return true
	}

	return false
}

// storageSizeGB extracts properties.storage.storageSizeGB from a body as a float64.
func storageSizeGB(body map[string]interface{}) (float64, bool) {
	props, ok := body["properties"].(map[string]interface{})
	if !ok {
		return 0, false
	}
	storage, ok := props["storage"].(map[string]interface{})
	if !ok {
		return 0, false
	}
	v, ok := storage["storageSizeGB"].(float64)
	return v, ok
}

func init() { azwise.Register(NewMySQLFlexibleServer()) }
