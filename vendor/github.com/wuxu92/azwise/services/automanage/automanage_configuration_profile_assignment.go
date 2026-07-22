package automanage

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomanageConfigurationProfileAssignment provides resource knowledge for
// Microsoft.Automanage/configurationProfileAssignments.
//
// This ARM type is the merge target for two AzureRM resources that both create a
// configuration profile assignment (identical ARM type; they differ only in the
// target resource the assignment attaches to):
//   - azurerm_virtual_machine_automanage_configuration_assignment (target = a VM)
//   - azurerm_arc_machine_automanage_configuration_assignment      (target = an Arc machine)
//
// Both hardcode the assignment name to "default" (an ARM API requirement) and set
// properties.configurationProfile to the profile resource ID. The VM variant also
// sets properties.targetId to the VM ID; the Arc variant lets the server populate
// targetId from the parent scope, so targetId is not marked ForceNew/Required here
// to avoid breaking the Arc path.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automanage/virtual_machine_automanage_configuration_assignment_resource.go
//     (schema 25-40, Create 54-106; name hardcoded "default"; timeouts 30m create / 5m read / 30m delete)
//   - terraform-provider-azurerm internal/services/automanage/arc_machine_automanage_configuration_assignment_resource.go
//     (schema 26-41, Create 55-105; targetId server-populated)
//   - go-azure-sdk resource-manager/automanage/2022-05-04/configurationprofileassignments
//     ConfigurationProfileAssignmentProperties{configurationProfile, targetId, status}.
type AutomanageConfigurationProfileAssignment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomanageConfigurationProfileAssignment)(nil)

// NewAutomanageConfigurationProfileAssignment returns knowledge for the
// configurationProfileAssignments resource.
func NewAutomanageConfigurationProfileAssignment() *AutomanageConfigurationProfileAssignment {
	return &AutomanageConfigurationProfileAssignment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automanage/configurationProfileAssignments",
			ApiVersions:  []string{"2022-05-04"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// configuration_id (properties.configurationProfile) is ForceNew in both
			// AzureRM resources. The target (virtual_machine_id / arc_machine_id) is
			// also ForceNew; for the VM resource it is written to properties.targetId.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.configurationProfile"},
			},
			// ARM requires the configuration profile reference on the body.
			RequiredFields: []string{
				"properties.configurationProfile",
			},
			// The assignment name must be "default" (ARM API requirement; AzureRM
			// hardcodes it). Validated against the resource-name attribute (empty path).
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "",
					AllowedValues: []string{"default"},
					Message:       "automanage configuration profile assignment name must be \"default\"",
				},
			},
		},
	}
}

func init() { azwise.Register(NewAutomanageConfigurationProfileAssignment()) }
