package maintenance

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ConfigurationAssignment provides resource knowledge for
// Microsoft.Maintenance/configurationAssignments.
//
// This ARM type is created by five AzureRM resources, all of which POST a
// ConfigurationAssignment to the same provider type at different scopes. They are
// merged here; only knowledge universal to every contributing body is unioned.
//
// Contributing TF resources (all go-azure-sdk .../configurationassignments):
//   - azurerm_maintenance_assignment_dedicated_host
//     (maintenance_assignment_dedicated_host_resource.go: schema 57-76, body 110-116;
//      scope = commonids DedicatedHost; NewScopedConfigurationAssignmentID)
//   - azurerm_maintenance_assignment_virtual_machine
//     (maintenance_assignment_virtual_machine_resource.go: schema 55-72, body ~110)
//   - azurerm_maintenance_assignment_virtual_machine_scale_set
//     (maintenance_assignment_virtual_machine_scale_set_resource.go: schema 55-74)
//   - azurerm_maintenance_assignment_dynamic_scope
//     (maintenance_assignment_dynamic_scope_resource.go: schema 45-146, Create 160-240;
//      subscription-scoped NewConfigurationAssignmentID; sets properties.filter, no resourceId)
//
// The scoped resources (dedicated_host/vm/vmss) set properties.resourceId and
// properties.maintenanceConfigurationId; the dynamic_scope resource sets
// properties.maintenanceConfigurationId and properties.filter (no resourceId).
//
// Sources:
//   - go-azure-sdk resource-manager/maintenance/2023-04-01/configurationassignments
//     ConfigurationAssignmentProperties (maintenanceConfigurationId/resourceId/filter),
//     ConfigurationAssignmentFilterProperties (locations/osTypes/resourceGroups/
//     resourceTypes/tagSettings), TagSettingsProperties (filterOperator/tags),
//     constants.go (PossibleValuesForTagOperators), id_configurationassignment.go /
//     id_scopedconfigurationassignment.go (ARM type Microsoft.Maintenance/configurationAssignments).
type ConfigurationAssignment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ConfigurationAssignment)(nil)

// NewConfigurationAssignment returns knowledge for the configurationAssignments resource.
func NewConfigurationAssignment() *ConfigurationAssignment {
	return &ConfigurationAssignment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Maintenance/configurationAssignments",
			ApiVersions:  []string{"2023-04-01"},
			SoftDelete:   false,
			// Create/Read/Delete 30m/5m/30m across all; dynamic_scope adds Update 30m.
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				// name is ForceNew (envelope); dynamic_scope exposes it explicitly, the
				// scoped resources derive it from the configuration name.
				{PropertyPath: "name"},
				// maintenance_configuration_id is ForceNew in every contributing resource.
				{PropertyPath: "properties.maintenanceConfigurationId"},
				// resourceId is ForceNew for the scoped resources; dynamic_scope never sets
				// it, so this rule only fires when resourceId is actually present.
				{PropertyPath: "properties.resourceId"},
			},
			// maintenanceConfigurationId is Required in all five resources.
			// resourceId (scoped only) and filter (dynamic_scope only) are kind-specific
			// and intentionally NOT unioned here.
			RequiredFields: []string{
				"properties.maintenanceConfigurationId",
			},
			StringRules: []azwise.StringRule{
				{
					// filter.tag_filter enum (PossibleValuesForTagOperators). Only the
					// dynamic_scope body sets properties.filter.tagSettings; the rule fires
					// only when that sub-object is present, so it is safe to union.
					PropertyPath: "properties.filter.tagSettings.filterOperator",
					AllowedValues: []string{
						"All",
						"Any",
					},
					Message: "tag_filter must be All or Any",
				},
			},
			// NOTES:
			//   - dynamic_scope filter.os_types ("Linux"/"Windows") and filter.resource_types
			//     ("Microsoft.Compute/virtualMachines"/"Microsoft.HybridCompute/machines") are
			//     array elements (properties.filter.osTypes[*] / resourceTypes[*]); array-element
			//     paths are not expressible as declarative rules, so they are skipped.
			//   - tag_filter's AzureRM Default ("Any") is NOT emitted as a DefaultValue: it is
			//     kind-specific to dynamic_scope, and injecting it unconditionally would fabricate
			//     properties.filter on the scoped (resourceId) bodies.
		},
	}
}

func init() { azwise.Register(NewConfigurationAssignment()) }
