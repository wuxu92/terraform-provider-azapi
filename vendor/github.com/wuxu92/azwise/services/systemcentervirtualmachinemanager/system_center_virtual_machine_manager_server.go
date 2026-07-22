package systemcentervirtualmachinemanager

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SystemCenterVirtualMachineManagerServer provides resource knowledge for
// Microsoft.ScVmm/vmmServers.
//
// Mirrors azurerm_system_center_virtual_machine_manager_server.
//
// Sources:
//   - terraform-provider-azurerm internal/services/systemcentervirtualmachinemanager/system_center_virtual_machine_manager_server_resource.go:63-187 (schema + Create)
//   - Timeouts: Create 180m (:117) / Read 5m (:191) / Update 180m (:247) / Delete 180m (:276)
//   - Name validation: validate/system_center_virtual_machine_manager_server_name.go:18 (regex, 1-54 chars)
//   - go-azure-sdk resource-manager/systemcentervirtualmachinemanager/2023-10-07/vmmservers
//     model_vmmserverproperties.go / model_vmmcredential.go / model_extendedlocation.go
//   - ARM type casing verified via vmmservers/id_vmmserver.go StaticSegment("staticVmmServers", "vmmServers").
type SystemCenterVirtualMachineManagerServer struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SystemCenterVirtualMachineManagerServer)(nil)

// NewSystemCenterVirtualMachineManagerServer returns knowledge for the SCVMM server resource.
func NewSystemCenterVirtualMachineManagerServer() *SystemCenterVirtualMachineManagerServer {
	return &SystemCenterVirtualMachineManagerServer{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ScVmm/vmmServers",
			ApiVersions:  []string{"2023-10-07"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 180 * time.Minute,
				Read:   5 * time.Minute,
				Update: 180 * time.Minute,
				Delete: 180 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "extendedLocation.name"},
				{PropertyPath: "properties.fqdn"},
				{PropertyPath: "properties.credentials.username"},
				{PropertyPath: "properties.credentials.password"},
				{PropertyPath: "properties.port"},
			},
			StringRules: []azwise.StringRule{
				// name: must start/end alphanumeric, may contain -.; 1-54 chars.
				{PropertyPath: "", Regex: `^[a-zA-Z0-9]([-.a-zA-Z0-9]{0,52}[a-zA-Z0-9])?$`, MinLength: 1, MaxLength: 54, Message: "must start and end with an alphanumeric character, may contain alphanumeric characters, dashes or periods and must be between 1 and 54 characters long"},
			},
			IntRules: []azwise.IntRule{
				// port: validation.IntBetween(1, 65535).
				{PropertyPath: "properties.port", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(65535))},
			},
			SensitiveFields: []string{"properties.credentials.password"},
			// extendedLocation.type is hardcoded "customLocation" by AzureRM.
			RequiredFields: []string{
				"extendedLocation.name",
				"properties.fqdn",
				"properties.credentials.username",
				"properties.credentials.password",
			},
			ComputedFields: []string{
				"properties.connectionStatus",
				"properties.errorMessage",
				"properties.provisioningState",
				"properties.uuid",
				"properties.version",
			},
		},
	}
}

func init() { azwise.Register(NewSystemCenterVirtualMachineManagerServer()) }
