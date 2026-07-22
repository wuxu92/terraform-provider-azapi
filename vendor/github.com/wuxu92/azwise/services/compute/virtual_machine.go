package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualMachine provides resource knowledge for Microsoft.Compute/virtualMachines.
//
// This ARM type is produced by two AzureRM Terraform resources, merged here since
// azwise is keyed by ARM resource type. Only knowledge that is universally true for
// BOTH OS variants is unioned. Value constraints that live on OS-distinct ARM paths
// (linuxConfiguration vs windowsConfiguration) are safe to include because they only
// fire when that sub-object is present.
//
// Contributing TF resources:
//   - azurerm_linux_virtual_machine   (internal/services/compute/linux_virtual_machine_resource.go)
//   - azurerm_windows_virtual_machine (internal/services/compute/windows_virtual_machine_resource.go)
//
// Sources:
//   - linux_virtual_machine_resource.go:40-486 (schema, timeouts, ForceNew, enums)
//   - windows_virtual_machine_resource.go:39-535 (schema, timeouts, ForceNew, enums)
//   - internal/services/compute/validate/virtual_machine_name.go (name validation)
//   - go-azure-sdk compute/2024-03-01/virtualmachines model_*.go + constants.go
//
// Not unioned (OS-specific, would corrupt validation for the other variant):
//   - properties.licenseType: Linux allows RHEL_*/SLES_*/UBUNTU_PRO, Windows allows
//     None/Windows_Client/Windows_Server — conflicting enum sets on the same path.
type VirtualMachine struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualMachine)(nil)

func NewVirtualMachine() *VirtualMachine {
	return &VirtualMachine{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/virtualMachines",
			ApiVersions:  []string{"2022-03-01", "2023-04-02", "2024-03-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 45 * time.Minute,
				Read:   5 * time.Minute,
				Update: 45 * time.Minute,
				Delete: 45 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "zones"},            // zone
				{PropertyPath: "extendedLocation"}, // edge_zone
				// osProfile fields — the whole OsProfile is immutable when set.
				{PropertyPath: "properties.osProfile.adminUsername"},
				{PropertyPath: "properties.osProfile.adminPassword"},
				{PropertyPath: "properties.osProfile.customData"},
				{PropertyPath: "properties.osProfile.computerName"},
				// hardware / placement
				{PropertyPath: "properties.availabilitySet.id"},        // availability_set_id
				{PropertyPath: "properties.platformFaultDomain"},       // platform_fault_domain
				{PropertyPath: "properties.priority"},                  // priority
				{PropertyPath: "properties.evictionPolicy"},            // eviction_policy
				{PropertyPath: "properties.storageProfile.imageReference.id"},        // source_image_id
				{PropertyPath: "properties.storageProfile.osDisk.managedDisk.id"},    // os_managed_disk_id (existing OS disk)
				{PropertyPath: "properties.securityProfile.uefiSettings.secureBootEnabled"}, // secure_boot_enabled
				{PropertyPath: "properties.securityProfile.uefiSettings.vTpmEnabled"},       // vtpm_enabled
				// provision_vm_agent — ForceNew for both OSes on OS-distinct paths.
				{PropertyPath: "properties.osProfile.linuxConfiguration.provisionVMAgent"},
				{PropertyPath: "properties.osProfile.windowsConfiguration.provisionVMAgent"},
				// Linux-only ForceNew (OS-distinct path, safe to union).
				{PropertyPath: "properties.osProfile.linuxConfiguration.disablePasswordAuthentication"},
				// Windows-only ForceNew (OS-distinct paths, safe to union).
				{PropertyPath: "properties.osProfile.windowsConfiguration.enableAutomaticUpdates"},
				{PropertyPath: "properties.osProfile.windowsConfiguration.timeZone"},
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── VirtualMachineName: 1-80 chars, alphanumerics/dots/dashes/underscores.
				{
					Regex:     `^[a-zA-Z0-9._-]+$`,
					MinLength: 1,
					MaxLength: 80,
					Message:   "must be 1-80 characters and may only contain alphanumeric characters, dots, dashes and underscores",
				},
				// priority (VirtualMachinePriorityTypes)
				{
					PropertyPath:  "properties.priority",
					AllowedValues: []string{"Low", "Regular", "Spot"},
				},
				// eviction_policy (VirtualMachineEvictionPolicyTypes)
				{
					PropertyPath:  "properties.evictionPolicy",
					AllowedValues: []string{"Deallocate", "Delete"},
				},
				// disk_controller_type (DiskControllerTypes)
				{
					PropertyPath:  "properties.storageProfile.diskControllerType",
					AllowedValues: []string{"NVMe", "SCSI"},
				},
				// ── Linux patch settings (OS-distinct paths) ──
				{
					PropertyPath:  "properties.osProfile.linuxConfiguration.patchSettings.patchMode",
					AllowedValues: []string{"AutomaticByPlatform", "ImageDefault"},
				},
				{
					PropertyPath:  "properties.osProfile.linuxConfiguration.patchSettings.assessmentMode",
					AllowedValues: []string{"AutomaticByPlatform", "ImageDefault"},
				},
				{
					PropertyPath:  "properties.osProfile.linuxConfiguration.patchSettings.automaticByPlatformSettings.rebootSetting",
					AllowedValues: []string{"Always", "IfRequired", "Never", "Unknown"},
				},
				// ── Windows patch settings (OS-distinct paths) ──
				{
					PropertyPath:  "properties.osProfile.windowsConfiguration.patchSettings.patchMode",
					AllowedValues: []string{"AutomaticByOS", "AutomaticByPlatform", "Manual"},
				},
				{
					PropertyPath:  "properties.osProfile.windowsConfiguration.patchSettings.assessmentMode",
					AllowedValues: []string{"AutomaticByPlatform", "ImageDefault"},
				},
				{
					PropertyPath:  "properties.osProfile.windowsConfiguration.patchSettings.automaticByPlatformSettings.rebootSetting",
					AllowedValues: []string{"Always", "IfRequired", "Never", "Unknown"},
				},
			},
			IntRules: []azwise.IntRule{
				// platform_fault_domain: IntAtLeast(-1); -1 is the AzureRM "unset" sentinel.
				{PropertyPath: "properties.platformFaultDomain", MinValue: azwise.Ptr(int64(-1))},
			},
			FloatRules: []azwise.FloatRule{
				// max_bid_price: FloatAtLeast(-1.0); -1 means "pay up to on-demand price".
				{PropertyPath: "properties.billingProfile.maxPrice", MinValue: azwise.Ptr(float64(-1))},
			},
			SensitiveFields: []string{
				"properties.osProfile.adminPassword",
				"properties.osProfile.customData",
			},
			// vmId is populated by Azure and never present in the create/update model.
			ComputedFields: []string{
				"properties.vmId",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.priority", Value: "Regular"},
				{PropertyPath: "properties.extensionsTimeBudget", Value: "PT1H30M"},
			},
			// Required for both OS variants.
			RequiredFields: []string{
				"properties.hardwareProfile.vmSize",           // size
				"properties.storageProfile.osDisk",            // os_disk
				"properties.networkProfile.networkInterfaces", // network_interface_ids
			},
		},
	}
}

func init() { azwise.Register(NewVirtualMachine()) }
