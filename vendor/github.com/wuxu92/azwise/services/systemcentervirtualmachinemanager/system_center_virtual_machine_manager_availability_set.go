package systemcentervirtualmachinemanager

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SystemCenterVirtualMachineManagerAvailabilitySet provides resource knowledge for
// Microsoft.ScVmm/availabilitySets.
//
// Mirrors azurerm_system_center_virtual_machine_manager_availability_set.
//
// Sources:
//   - terraform-provider-azurerm internal/services/systemcentervirtualmachinemanager/system_center_virtual_machine_manager_availability_set_resource.go:59-138 (schema + Create)
//   - Timeouts: Create 30m (:86) / Read 5m (:142) / Update 30m (:193) / Delete 30m (:224)
//   - Name validation: validate/system_center_virtual_machine_manager_availability_set_name.go:18 (regex, 1-54 chars)
//   - go-azure-sdk resource-manager/systemcentervirtualmachinemanager/2023-10-07/availabilitysets
//     model_availabilitysetproperties.go / model_extendedlocation.go
//   - ARM type casing verified via availabilitysets/id_availabilityset.go StaticSegment("staticAvailabilitySets", "availabilitySets").
//
// Note: AzureRM hardcodes properties.availabilitySetName to the resource name.
type SystemCenterVirtualMachineManagerAvailabilitySet struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SystemCenterVirtualMachineManagerAvailabilitySet)(nil)

// NewSystemCenterVirtualMachineManagerAvailabilitySet returns knowledge for the SCVMM availability set resource.
func NewSystemCenterVirtualMachineManagerAvailabilitySet() *SystemCenterVirtualMachineManagerAvailabilitySet {
	return &SystemCenterVirtualMachineManagerAvailabilitySet{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ScVmm/availabilitySets",
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
				{PropertyPath: "properties.availabilitySetName"},
				{PropertyPath: "properties.vmmServerId"},
			},
			StringRules: []azwise.StringRule{
				// name: must start/end alphanumeric, may contain -.; 1-54 chars.
				{PropertyPath: "", Regex: `^[a-zA-Z0-9]([-.a-zA-Z0-9]{0,52}[a-zA-Z0-9])?$`, MinLength: 1, MaxLength: 54, Message: "must start and end with an alphanumeric character, may contain alphanumeric characters, dashes or periods and must be between 1 and 54 characters long"},
			},
			// extendedLocation.type is hardcoded "customLocation" by AzureRM.
			RequiredFields: []string{
				"extendedLocation.name",
				"properties.availabilitySetName",
				"properties.vmmServerId",
			},
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewSystemCenterVirtualMachineManagerAvailabilitySet()) }
