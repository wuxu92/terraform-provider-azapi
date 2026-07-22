package systemcentervirtualmachinemanager

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SystemCenterVirtualMachineManagerVirtualMachineInstance provides resource knowledge for
// Microsoft.ScVmm/virtualMachineInstances.
//
// Mirrors azurerm_system_center_virtual_machine_manager_virtual_machine_instance.
//
// This is an EXTENSION resource: a singleton "default" virtualMachineInstances child
// scoped to a Microsoft.HybridCompute/machines resource (the ID is
// "<machine-scope>/providers/Microsoft.ScVmm/virtualMachineInstances/default"). The
// instance has no user-chosen name, so there is no name-validation StringRule.
//
// Sources:
//   - terraform-provider-azurerm internal/services/systemcentervirtualmachinemanager/system_center_virtual_machine_manager_virtual_machine_instance_resource.go:103-421,470-564 (schema + Create/Update/Delete)
//   - Timeouts: Create 60m (:358) / Read 5m (:425) / Update 60m (:472) / Delete 60m (:545)
//   - Validation: validate/system_center_virtual_machine_manager_virtual_machine_instance_computer_name.go:18 / _storage_disk_name.go:18 / _mac_address.go:18
//   - go-azure-sdk resource-manager/systemcentervirtualmachinemanager/2023-10-07/virtualmachineinstances
//     model_virtualmachineinstanceproperties.go / model_infrastructureprofile.go / model_hardwareprofile.go /
//     model_osprofileforvminstance.go / model_networkinterface.go / model_virtualdisk.go / model_storageprofile.go / constants.go
//   - Scope/parenting verified via parse/virtual_machine_instance.go ID(): ".../Microsoft.ScVmm/virtualMachineInstances/default".
//
// Notes:
//   - extendedLocation.type is hardcoded "customLocation" by AzureRM.
//   - network_interface (properties.networkProfile.networkInterfaces[*]), storage_disk
//     (properties.storageProfile.disks[*]) and system_center_virtual_machine_manager_availability_set_ids
//     (properties.availabilitySets[*]) are array-element paths: their per-element rules
//     (AllocationMethod enums, bus/lun/disk_size ranges, disk name/uuid) cannot be
//     lowered declaratively and are omitted here.
//   - The infrastructure/hardware/operating_system blocks are single objects (MaxItems 1),
//     so their AtLeastOneOf / RequiredWith cross-property constraints resolve to real ARM
//     paths and are emitted below.
type SystemCenterVirtualMachineManagerVirtualMachineInstance struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SystemCenterVirtualMachineManagerVirtualMachineInstance)(nil)

// NewSystemCenterVirtualMachineManagerVirtualMachineInstance returns knowledge for the SCVMM VM instance resource.
func NewSystemCenterVirtualMachineManagerVirtualMachineInstance() *SystemCenterVirtualMachineManagerVirtualMachineInstance {
	return &SystemCenterVirtualMachineManagerVirtualMachineInstance{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ScVmm/virtualMachineInstances",
			ApiVersions:  []string{"2023-10-07"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "extendedLocation.name"},
				{PropertyPath: "properties.hardwareProfile"},
				{PropertyPath: "properties.osProfile"},
				{PropertyPath: "properties.infrastructureProfile.cloudId"},
				{PropertyPath: "properties.infrastructureProfile.inventoryItemId"},
				{PropertyPath: "properties.infrastructureProfile.templateId"},
				{PropertyPath: "properties.infrastructureProfile.vmmServerId"},
			},
			StringRules: []azwise.StringRule{
				// checkpoint_type: SDK field is a free-form *string; AzureRM restricts to these values.
				{PropertyPath: "properties.infrastructureProfile.checkpointType", AllowedValues: []string{"Disabled", "Production", "ProductionOnly", "Standard"}, Message: "must be one of Disabled, Production, ProductionOnly, Standard"},
				// operating_system.computer_name: alphanumeric only.
				{PropertyPath: "properties.osProfile.computerName", Regex: `^[a-zA-Z0-9]{1,}$`, MinLength: 1, Message: "must only contain alphanumeric characters"},
			},
			IntRules: []azwise.IntRule{
				// hardware.cpu_count: IntBetween(1, 64).
				{PropertyPath: "properties.hardwareProfile.cpuCount", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(64))},
				// hardware.dynamic_memory_max_in_mb: IntBetween(32, 1048576).
				{PropertyPath: "properties.hardwareProfile.dynamicMemoryMaxMB", MinValue: azwise.Ptr(int64(32)), MaxValue: azwise.Ptr(int64(1048576))},
				// hardware.dynamic_memory_min_in_mb: IntBetween(32, 1048576).
				{PropertyPath: "properties.hardwareProfile.dynamicMemoryMinMB", MinValue: azwise.Ptr(int64(32)), MaxValue: azwise.Ptr(int64(1048576))},
				// hardware.memory_in_mb: IntBetween(32, 1048576).
				{PropertyPath: "properties.hardwareProfile.memoryMB", MinValue: azwise.Ptr(int64(32)), MaxValue: azwise.Ptr(int64(1048576))},
			},
			SensitiveFields: []string{"properties.osProfile.adminPassword"},
			RequiredFields: []string{
				"extendedLocation.name",
				"properties.infrastructureProfile",
			},
			ComputedFields: []string{
				"properties.powerState",
				"properties.provisioningState",
			},
			RequiredWith: []azwise.RelationalRule{
				// cloud_id RequiredWith template_id (and vice versa).
				{Paths: []string{"properties.infrastructureProfile.cloudId", "properties.infrastructureProfile.templateId"}},
				{Paths: []string{"properties.infrastructureProfile.templateId", "properties.infrastructureProfile.cloudId"}},
				// inventory_item_id RequiredWith virtual_machine_server_id.
				{Paths: []string{"properties.infrastructureProfile.inventoryItemId", "properties.infrastructureProfile.vmmServerId"}},
			},
			AtLeastOneOf: []azwise.RelationalRule{
				// infrastructure: at least one source of the VM definition.
				{Paths: []string{
					"properties.infrastructureProfile.cloudId",
					"properties.infrastructureProfile.inventoryItemId",
					"properties.infrastructureProfile.templateId",
					"properties.infrastructureProfile.vmmServerId",
				}},
				// hardware: at least one hardware attribute must be set.
				{Paths: []string{
					"properties.hardwareProfile.cpuCount",
					"properties.hardwareProfile.dynamicMemoryMaxMB",
					"properties.hardwareProfile.dynamicMemoryMinMB",
					"properties.hardwareProfile.limitCpuForMigration",
					"properties.hardwareProfile.memoryMB",
				}},
				// operating_system: at least one of computer_name / admin_password.
				{Paths: []string{
					"properties.osProfile.computerName",
					"properties.osProfile.adminPassword",
				}},
			},
		},
	}
}

func init() { azwise.Register(NewSystemCenterVirtualMachineManagerVirtualMachineInstance()) }
