package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ManagedDisk provides resource knowledge for Microsoft.Compute/disks.
//
// Contributing Terraform resource: azurerm_managed_disk.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/managed_disk_resource.go
//     (schema L62-316, Create body L319-592)
//   - terraform-provider-azurerm internal/services/compute/validate/managed_disk_size_gb.go
//   - go-azure-sdk resource-manager/compute/2023-04-02/disks:
//     model_diskproperties.go, model_creationdata.go, model_encryption.go,
//     model_disksecurityprofile.go, model_disksku.go, constants.go (enums).
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; name is ForceNew but not an
//     ARM-body property, so it is not emitted as a body ForceNew rule.
//   - encryption_settings (properties.encryptionSettingsCollection) is ForceNew only when
//     removed once set (CustomizeDiff ForceNewIfChange, managed_disk_resource.go L311-315);
//     a value-conditional ForceNew that cannot be expressed declaratively, left as a note.
//   - logical_sector_size uses IntInSlice{512,4096} (an enum of ints, not a range), which
//     the IntRule min/max model cannot express — left as a note.
//   - trusted_launch_enabled (bool) and security_type (enum) both drive
//     properties.securityProfile.securityType; both are ForceNew, encoded once on that path.
//   - disk_iops_read_only/disk_mbps_read_only/logical_sector_size are only valid for
//     UltraSSD/PremiumV2 disks with shared disk enabled — value-conditional, left as a note.
type ManagedDisk struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ManagedDisk)(nil)

// NewManagedDisk returns knowledge for the disks resource.
func NewManagedDisk() *ManagedDisk {
	return &ManagedDisk{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/disks",
			ApiVersions:  []string{"2023-04-02"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "extendedLocation.name"},
				{PropertyPath: "zones"},
				{PropertyPath: "properties.hyperVGeneration"},
				{PropertyPath: "properties.creationData.createOption"},
				{PropertyPath: "properties.creationData.logicalSectorSize"},
				{PropertyPath: "properties.creationData.performancePlus"},
				{PropertyPath: "properties.creationData.sourceUri"},
				{PropertyPath: "properties.creationData.sourceResourceId"},
				{PropertyPath: "properties.creationData.storageAccountId"},
				{PropertyPath: "properties.creationData.imageReference.id"},
				{PropertyPath: "properties.creationData.galleryImageReference.id"},
				{PropertyPath: "properties.creationData.uploadSizeBytes"},
				{PropertyPath: "properties.securityProfile.securityType"},
				{PropertyPath: "properties.securityProfile.secureVMDiskEncryptionSetId"},
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "sku.name", AllowedValues: []string{
					"Premium_LRS",
					"PremiumV2_LRS",
					"Premium_ZRS",
					"Standard_LRS",
					"StandardSSD_LRS",
					"StandardSSD_ZRS",
					"UltraSSD_LRS",
				}},
				{PropertyPath: "properties.creationData.createOption", AllowedValues: []string{
					"Attach",
					"Copy",
					"CopyFromSanSnapshot",
					"CopyStart",
					"Empty",
					"FromImage",
					"Import",
					"ImportSecure",
					"Restore",
					"Upload",
					"UploadPreparedSecure",
				}},
				{PropertyPath: "properties.osType", AllowedValues: []string{
					"Linux",
					"Windows",
				}},
				{PropertyPath: "properties.networkAccessPolicy", AllowedValues: []string{
					"AllowAll",
					"AllowPrivate",
					"DenyAll",
				}},
				{PropertyPath: "properties.securityProfile.securityType", AllowedValues: []string{
					"ConfidentialVM_DiskEncryptedWithCustomerKey",
					"ConfidentialVM_DiskEncryptedWithPlatformKey",
					"ConfidentialVM_VMGuestStateOnlyEncryptedWithPlatformKey",
					"TrustedLaunch",
				}},
				{PropertyPath: "properties.hyperVGeneration", AllowedValues: []string{
					"V1",
					"V2",
				}},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.diskSizeGB", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(65536))},
				{PropertyPath: "properties.maxShares", MinValue: azwise.Ptr(int64(2)), MaxValue: azwise.Ptr(int64(10))},
				{PropertyPath: "properties.diskIOPSReadWrite", MinValue: azwise.Ptr(int64(1))},
				{PropertyPath: "properties.diskMBpsReadWrite", MinValue: azwise.Ptr(int64(1))},
				{PropertyPath: "properties.diskIOPSReadOnly", MinValue: azwise.Ptr(int64(1))},
				{PropertyPath: "properties.diskMBpsReadOnly", MinValue: azwise.Ptr(int64(1))},
				{PropertyPath: "properties.creationData.uploadSizeBytes", MinValue: azwise.Ptr(int64(1))},
			},
			// image_reference_id conflicts with gallery_image_reference_id; disk_encryption_set_id
			// conflicts with secure_vm_disk_encryption_set_id.
			ConflictsWith: []azwise.RelationalRule{
				{Paths: []string{"properties.creationData.imageReference.id", "properties.creationData.galleryImageReference.id"}, Message: "image_reference_id conflicts with gallery_image_reference_id"},
				{Paths: []string{"properties.encryption.diskEncryptionSetId", "properties.securityProfile.secureVMDiskEncryptionSetId"}, Message: "disk_encryption_set_id conflicts with secure_vm_disk_encryption_set_id"},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.optimizedForFrequentAttach", Value: false},
				{PropertyPath: "properties.creationData.performancePlus", Value: false},
				{PropertyPath: "properties.networkAccessPolicy", Value: "AllowAll"},
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
			},
			RequiredFields: []string{
				"sku.name",
				"properties.creationData.createOption",
			},
			// Read-only server-populated properties returned by GET but never authored.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.diskState",
				"properties.diskSizeBytes",
				"properties.timeCreated",
				"properties.uniqueId",
				"properties.shareInfo",
				"properties.propertyUpdatesInProgress",
				"properties.completionPercent",
			},
		},
	}
}

func init() { azwise.Register(NewManagedDisk()) }
