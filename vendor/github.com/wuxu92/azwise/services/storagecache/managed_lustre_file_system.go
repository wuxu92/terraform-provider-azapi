package storagecache

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ManagedLustreFileSystem provides resource knowledge for
// Microsoft.StorageCache/amlFilesystems.
//
// Mirrors azurerm_managed_lustre_file_system.
//
// Sources:
//   - internal/services/storagecache/managed_lustre_file_system_resource.go
//     (Arguments 131-284: name ForceNew ManagedLustreFileSystemName; location ForceNew;
//     maintenance_window Required {day_of_week StringInSlice(MaintenanceDayOfWeekType),
//     time_of_day_in_utc}; root_squash {mode StringInSlice(All,RootOnly), no_squash_nids,
//     squash_gid/uid Optional Default 0 IntAtLeast(1)}; sku_name Required ForceNew
//     StringInSlice(4 AMLFS SKUs); storage_capacity_in_tb Required ForceNew IntAtLeast(4);
//     subnet_id Required ForceNew; zones RequiredForceNew; hsm_setting Optional ForceNew;
//     identity OptionalForceNew; encryption_key {key_url, source_vault_id};
//     Attributes 286-293: mgs_address Computed; CustomizeDiff 295-322: encryption_key
//     removal ForceNew + sku/capacity increment check; create 353-368 →
//     AmlFilesystem{Properties{Hsm, EncryptionSettings, MaintenanceWindow, FilesystemSubnet,
//     StorageCapacityTiB, RootSquashSettings}, Sku.Name, Zones}; timeouts C/U 30m R 5m D 30m).
//   - internal/services/storagecache/validate/managed_lustre_file_system_name.go
//     (regex ^[0-9a-zA-Z][-0-9a-zA-Z_]{0,78}[0-9a-zA-Z]$).
//   - go-azure-sdk resource-manager/storagecache/2024-07-01/amlfilesystems:
//     model_amlfilesystemproperties.go (filesystemSubnet/storageCapacityTiB/hsm/
//     encryptionSettings/maintenanceWindow/rootSquashSettings; clientInfo/health/
//     provisioningState/throughputProvisionedMBps read-only),
//     model_amlfilesystempropertiesmaintenancewindow.go (dayOfWeek/timeOfDayUTC),
//     model_amlfilesystemrootsquashsettings.go (mode/noSquashNidLists/squashGID/squashUID),
//     model_amlfilesystemhsmsettings.go (container/loggingContainer/importPrefix),
//     model_skuname.go (name), constants.go (MaintenanceDayOfWeekType 7 days,
//     AmlFilesystemSquashMode All/None/RootOnly),
//     id_amlfilesystem.go (type segment casing "amlFilesystems").
type ManagedLustreFileSystem struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ManagedLustreFileSystem)(nil)

// NewManagedLustreFileSystem returns knowledge for the amlFilesystems resource.
func NewManagedLustreFileSystem() *ManagedLustreFileSystem {
	return &ManagedLustreFileSystem{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StorageCache/amlFilesystems",
			ApiVersions:  []string{"2024-07-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "sku.name"},
				{PropertyPath: "properties.storageCapacityTiB"},
				{PropertyPath: "properties.filesystemSubnet"},
				{PropertyPath: "zones"},
				{PropertyPath: "properties.hsm"},
				// identity uses UserAssignedIdentityOptionalForceNew.
				{PropertyPath: "identity"},
			},
			StringRules: []azwise.StringRule{
				{
					// name: validate.ManagedLustreFileSystemName.
					Regex:     `^[0-9a-zA-Z][-0-9a-zA-Z_]{0,78}[0-9a-zA-Z]$`,
					MinLength: 2,
					MaxLength: 80,
					Message:   "must be 2-80 characters, alphanumeric/hyphen/underscore, starting and ending with alphanumeric",
				},
				{
					PropertyPath: "sku.name",
					AllowedValues: []string{
						"AMLFS-Durable-Premium-40",
						"AMLFS-Durable-Premium-125",
						"AMLFS-Durable-Premium-250",
						"AMLFS-Durable-Premium-500",
					},
				},
				{
					PropertyPath: "properties.maintenanceWindow.dayOfWeek",
					AllowedValues: []string{
						"Friday",
						"Monday",
						"Saturday",
						"Sunday",
						"Thursday",
						"Tuesday",
						"Wednesday",
					},
				},
				{
					// root_squash mode. AzureRM restricts to {All, RootOnly}; ARM SDK set
					// is {All, None, RootOnly} and AzAPI sends raw ARM values.
					PropertyPath:  "properties.rootSquashSettings.mode",
					AllowedValues: []string{"All", "None", "RootOnly"},
				},
			},
			FloatRules: []azwise.FloatRule{
				{
					// storage_capacity_in_tb: IntAtLeast(4); ARM storageCapacityTiB is float64.
					PropertyPath: "properties.storageCapacityTiB",
					MinValue:     azwise.Ptr(float64(4)),
				},
			},
			IntRules: []azwise.IntRule{
				{
					// squash_gid: IntAtLeast(1).
					PropertyPath: "properties.rootSquashSettings.squashGID",
					MinValue:     azwise.Ptr(int64(1)),
				},
				{
					// squash_uid: IntAtLeast(1).
					PropertyPath: "properties.rootSquashSettings.squashUID",
					MinValue:     azwise.Ptr(int64(1)),
				},
			},
			RequiredFields: []string{
				"sku.name",
				"properties.storageCapacityTiB",
				"properties.filesystemSubnet",
				"properties.maintenanceWindow.dayOfWeek",
				"properties.maintenanceWindow.timeOfDayUTC",
				"zones",
			},
			// Server-populated, read-only ARM properties (absent from the create body).
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.clientInfo",
				"properties.health",
				"properties.throughputProvisionedMBps",
			},
			// NOTE: encryption_key maps to
			// properties.encryptionSettings.keyEncryptionKey.{keyUrl,sourceVault.id};
			// it is conditionally ForceNew (only when removed, via CustomizeDiff) and its
			// key_url is a Key Vault nested-item ID validated semantically — neither is an
			// unconditional declarative rule.
			// NOTE: storage_capacity_in_tb also has a CustomizeDiff cross-field check
			// requiring the value to be a multiple of the SKU's minimum increment
			// (48/16/8/4 for the four SKUs); that cross-field constraint is not declarative.
			// NOTE: hsm_setting maps to properties.hsm.settings.{container,loggingContainer,
			// importPrefix}; its Storage-Container-ID validators are semantic and its whole
			// block is ForceNew (covered by the properties.hsm ForceNew rule above).
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewManagedLustreFileSystem()) }
