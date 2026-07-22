package azurestackhci

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StackHCIVirtualHardDisk provides resource knowledge for Microsoft.AzureStackHCI/virtualHardDisks.
//
// Mirrors azurerm_stack_hci_virtual_hard_disk. name / resource_group_name / location live on
// the operational envelope; location is surfaced here as ForceNew (commonschema.Location).
// Every argument on this resource is ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/azurestackhci/stack_hci_virtual_hard_disk_resource.go:60-142
//     (schema: ForceNew, validators, enums, defaults)
//   - terraform-provider-azurerm internal/services/azurestackhci/stack_hci_virtual_hard_disk_resource.go:172-208
//     (create mapping to virtualharddisks.VirtualHardDiskProperties + extendedLocation)
//   - go-azure-sdk resource-manager/azurestackhci/2024-01-01/virtualharddisks:
//     model_virtualharddiskproperties.go (diskSizeGB, dynamic, blockSizeBytes, containerId,
//     diskFileFormat, hyperVGeneration, logicalSectorBytes, physicalSectorBytes),
//     constants.go:14-24 (DiskFileFormat vhd/vhdx), 93-96 (HyperVGeneration V1/V2)
//
// Not encoded (deliberate):
//   - custom_location_id (commonschema.ResourceIDReferenceRequiredForceNew) and storage_path_id
//     (storagecontainers.ValidateStorageContainerID) carry resource-ID semantic validators that
//     belong in an azapin customizer (validators.AzureResourceID) attached to
//     extendedLocation.name / properties.containerId, not a StringRule.
type StackHCIVirtualHardDisk struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StackHCIVirtualHardDisk)(nil)

// NewStackHCIVirtualHardDisk returns knowledge for the Azure Stack HCI virtualHardDisks resource.
func NewStackHCIVirtualHardDisk() *StackHCIVirtualHardDisk {
	return &StackHCIVirtualHardDisk{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AzureStackHCI/virtualHardDisks",
			ApiVersions:  []string{"2024-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "extendedLocation.name"},
				{PropertyPath: "properties.diskSizeGB"},
				{PropertyPath: "properties.blockSizeBytes"},
				{PropertyPath: "properties.diskFileFormat"},
				{PropertyPath: "properties.dynamic"},
				{PropertyPath: "properties.hyperVGeneration"},
				{PropertyPath: "properties.logicalSectorBytes"},
				{PropertyPath: "properties.physicalSectorBytes"},
				{PropertyPath: "properties.containerId"},
			},
			RequiredFields: []string{
				"properties.diskSizeGB",
				// AzureRM always sends the custom location as the extended location.
				"extendedLocation.name",
				"extendedLocation.type",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// name: begin/end alphanumeric, 2-64 chars, alphanumeric + - . _ inside.
					Regex:   `^[a-zA-Z0-9][\-\.\_a-zA-Z0-9]{0,62}[a-zA-Z0-9]$`,
					Message: "name must begin and end with an alphanumeric character, be between 2 and 64 characters in length and can only contain alphanumeric characters, hyphens, periods or underscores",
				},
				{
					PropertyPath:  "properties.diskFileFormat",
					AllowedValues: []string{"vhd", "vhdx"},
					Message:       "must be vhd or vhdx",
				},
				{
					PropertyPath:  "properties.hyperVGeneration",
					AllowedValues: []string{"V1", "V2"},
					Message:       "must be V1 or V2",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.diskSizeGB",
					MinValue:     azwise.Ptr(int64(1)),
					Message:      "disk size must be at least 1 GB",
				},
				{
					PropertyPath: "properties.blockSizeBytes",
					MinValue:     azwise.Ptr(int64(1)),
					Message:      "block size must be at least 1 byte",
				},
				{
					PropertyPath: "properties.logicalSectorBytes",
					MinValue:     azwise.Ptr(int64(1)),
					Message:      "logical sector size must be at least 1 byte",
				},
				{
					PropertyPath: "properties.physicalSectorBytes",
					MinValue:     azwise.Ptr(int64(1)),
					Message:      "physical sector size must be at least 1 byte",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.dynamic", Value: false},
			},
		},
	}
}

func init() { azwise.Register(NewStackHCIVirtualHardDisk()) }
