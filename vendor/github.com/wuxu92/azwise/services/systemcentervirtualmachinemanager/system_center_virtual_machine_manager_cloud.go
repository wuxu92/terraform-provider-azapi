package systemcentervirtualmachinemanager

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SystemCenterVirtualMachineManagerCloud provides resource knowledge for
// Microsoft.ScVmm/clouds.
//
// Mirrors azurerm_system_center_virtual_machine_manager_cloud.
//
// Sources:
//   - terraform-provider-azurerm internal/services/systemcentervirtualmachinemanager/system_center_virtual_machine_manager_cloud_resource.go:60-140 (schema + Create)
//   - Timeouts: Create 30m (:87) / Read 5m (:144) / Update 30m (:191) / Delete 30m (:222)
//   - Name validation: validate/system_center_virtual_machine_manager_cloud_name.go:18 (regex, 1-54 chars)
//   - go-azure-sdk resource-manager/systemcentervirtualmachinemanager/2023-10-07/clouds
//     model_cloudproperties.go / model_extendedlocation.go
//   - ARM type casing verified via clouds/id_cloud.go StaticSegment("staticClouds", "clouds").
//
// Note: system_center_virtual_machine_manager_server_inventory_item_id maps to
// properties.inventoryItemId; AzureRM also derives properties.uuid and
// properties.vmmServerId from it (one-to-many), so those are set on create.
type SystemCenterVirtualMachineManagerCloud struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SystemCenterVirtualMachineManagerCloud)(nil)

// NewSystemCenterVirtualMachineManagerCloud returns knowledge for the SCVMM cloud resource.
func NewSystemCenterVirtualMachineManagerCloud() *SystemCenterVirtualMachineManagerCloud {
	return &SystemCenterVirtualMachineManagerCloud{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ScVmm/clouds",
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
				"properties.cloudCapacity",
				"properties.cloudName",
				"properties.provisioningState",
				"properties.storageQoSPolicies",
			},
		},
	}
}

func init() { azwise.Register(NewSystemCenterVirtualMachineManagerCloud()) }
