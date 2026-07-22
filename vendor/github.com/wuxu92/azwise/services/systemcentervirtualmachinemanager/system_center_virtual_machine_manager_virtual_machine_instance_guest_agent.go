package systemcentervirtualmachinemanager

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SystemCenterVirtualMachineManagerVirtualMachineInstanceGuestAgent provides resource
// knowledge for Microsoft.ScVmm/virtualMachineInstances/guestAgents.
//
// Mirrors azurerm_system_center_virtual_machine_manager_virtual_machine_instance_guest_agent.
//
// This is a singleton "default" guestAgents child of the singleton "default"
// virtualMachineInstances extension resource, itself scoped to a
// Microsoft.HybridCompute/machines resource (ID:
// "<machine-scope>/providers/Microsoft.ScVmm/virtualMachineInstances/default/guestAgents/default").
//
// Sources:
//   - terraform-provider-azurerm internal/services/systemcentervirtualmachinemanager/system_center_virtual_machine_manager_virtual_machine_instance_guest_agent_resource.go:46-186 (schema + Create)
//   - Timeouts: Create 30m (:86) / Read 5m (:131) / Delete 30m (:170) (no Update)
//   - go-azure-sdk resource-manager/systemcentervirtualmachinemanager/2023-10-07/guestagents
//     model_guestagentproperties.go / model_guestcredential.go / constants.go (ProvisioningAction)
//   - Scope/parenting verified via parse/virtual_machine_instance_guest_agent.go ID():
//     ".../virtualMachineInstances/default/guestAgents/default".
type SystemCenterVirtualMachineManagerVirtualMachineInstanceGuestAgent struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SystemCenterVirtualMachineManagerVirtualMachineInstanceGuestAgent)(nil)

// NewSystemCenterVirtualMachineManagerVirtualMachineInstanceGuestAgent returns knowledge for the SCVMM VM instance guest agent resource.
func NewSystemCenterVirtualMachineManagerVirtualMachineInstanceGuestAgent() *SystemCenterVirtualMachineManagerVirtualMachineInstanceGuestAgent {
	return &SystemCenterVirtualMachineManagerVirtualMachineInstanceGuestAgent{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ScVmm/virtualMachineInstances/guestAgents",
			ApiVersions:  []string{"2023-10-07"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.credentials.username"},
				{PropertyPath: "properties.credentials.password"},
				{PropertyPath: "properties.provisioningAction"},
			},
			StringRules: []azwise.StringRule{
				// provisioning_action: guestagents.PossibleValuesForProvisioningAction().
				{PropertyPath: "properties.provisioningAction", AllowedValues: []string{"install", "repair", "uninstall"}, Message: "must be one of install, repair, uninstall"},
			},
			SensitiveFields: []string{"properties.credentials.password"},
			RequiredFields: []string{
				"properties.credentials.username",
				"properties.credentials.password",
			},
			DefaultValues: []azwise.DefaultValue{
				// provisioning_action defaults to "install".
				{PropertyPath: "properties.provisioningAction", Value: "install"},
			},
			ComputedFields: []string{
				"properties.customResourceName",
				"properties.provisioningState",
				"properties.status",
				"properties.uuid",
			},
		},
	}
}

func init() {
	azwise.Register(NewSystemCenterVirtualMachineManagerVirtualMachineInstanceGuestAgent())
}
