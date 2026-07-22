package systemcentervirtualmachinemanager

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SystemCenterVirtualMachineManagerVirtualMachineTemplate provides resource knowledge for
// Microsoft.ScVmm/virtualMachineTemplates.
//
// Mirrors azurerm_system_center_virtual_machine_manager_virtual_machine_template.
//
// Sources:
//   - terraform-provider-azurerm internal/services/systemcentervirtualmachinemanager/system_center_virtual_machine_manager_virtual_machine_template_resource.go:60-140 (schema + Create)
//   - Timeouts: Create 30m (:87) / Read 5m (:144) / Update 30m (:191) / Delete 30m (:222)
//   - Name validation: validate/system_center_virtual_machine_manager_virtual_machine_template_name.go:18 (regex, 1-54 chars)
//   - go-azure-sdk resource-manager/systemcentervirtualmachinemanager/2023-10-07/virtualmachinetemplates
//     model_virtualmachinetemplateproperties.go / model_extendedlocation.go
//   - ARM type casing verified via virtualmachinetemplates/id_virtualmachinetemplate.go StaticSegment("staticVirtualMachineTemplates", "virtualMachineTemplates").
//
// Note: only inventoryItemId (and derived uuid/vmmServerId) are set on create; all
// other properties (cpuCount, memoryMB, disks, networkInterfaces, osType, ...) are
// server-populated from the discovered SCVMM inventory item and are read-only.
type SystemCenterVirtualMachineManagerVirtualMachineTemplate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SystemCenterVirtualMachineManagerVirtualMachineTemplate)(nil)

// NewSystemCenterVirtualMachineManagerVirtualMachineTemplate returns knowledge for the SCVMM VM template resource.
func NewSystemCenterVirtualMachineManagerVirtualMachineTemplate() *SystemCenterVirtualMachineManagerVirtualMachineTemplate {
	return &SystemCenterVirtualMachineManagerVirtualMachineTemplate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ScVmm/virtualMachineTemplates",
			ApiVersions:  []string{"2023-10-07"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "extendedLocation.name"},
				{PropertyPath: "properties.inventoryItemId"},
			},
			StringRules: []azwise.StringRule{
				// name: must start/end alphanumeric, may contain -.; 1-54 chars.
				{PropertyPath: "", Regex: `^[a-zA-Z0-9]([-.a-zA-Z0-9]{0,52}[a-zA-Z0-9])?$`, MinLength: 1, MaxLength: 54, Message: "must start and end with an alphanumeric character, may contain alphanumeric characters, dashes or periods and must be between 1 and 54 characters long"},
			},
			// extendedLocation.type is hardcoded "customLocation" by AzureRM.
			RequiredFields: []string{
				"extendedLocation.name",
				"properties.inventoryItemId",
			},
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewSystemCenterVirtualMachineManagerVirtualMachineTemplate()) }
