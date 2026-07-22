// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PolicyVirtualMachineConfigurationAssignment provides resource knowledge for
// Microsoft.GuestConfiguration/guestConfigurationAssignments.
//
// Mirrors azurerm_policy_virtual_machine_configuration_assignment. The assignment is
// an extension resource on a virtual machine: its scope is the VM (azurerm's
// virtual_machine_id) and its ARM name is the assignment name — both live on the
// envelope (ForceNew by construction) and are not repeated as body ForceNew rules.
//
// The single required `configuration` block (MaxItems 1) maps to
// properties.guestConfiguration. `content_hash`/`content_uri` are Optional+Computed
// and left to the server (azurerm additionally clears them for built-in
// configurations whose content_uri points at a service-owned storage account); they
// are not declared as ComputedFields because a user MAY set them via ARM.
//
// The nested `parameter` set maps to configurationParameter (array elements), which
// azwise cannot address, so no rules target it.
//
// Sources:
//   - terraform-provider-azurerm internal/services/policy/policy_virtual_machine_configuration_assignment_resource.go
//     schema (name/virtual_machine_id ForceNew; location Required; configuration
//     Required MaxItems 1; assignment_type enum), timeouts 30m/5m/30m/30m; createFunc
//     builds Properties.GuestConfiguration
//   - go-azure-sdk resource-manager/guestconfiguration/2024-04-05/guestconfigurationassignments
//     GuestConfigurationAssignmentProperties.guestConfiguration ->
//     GuestConfigurationNavigation: assignmentType/contentHash/contentUri/
//     configurationParameter/version/name; AssignmentType enum
//     ApplyAndAutoCorrect/ApplyAndMonitor/Audit/DeployAndAutoCorrect
type PolicyVirtualMachineConfigurationAssignment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PolicyVirtualMachineConfigurationAssignment)(nil)

// NewPolicyVirtualMachineConfigurationAssignment returns knowledge for the
// guestConfigurationAssignments resource.
func NewPolicyVirtualMachineConfigurationAssignment() *PolicyVirtualMachineConfigurationAssignment {
	return &PolicyVirtualMachineConfigurationAssignment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.GuestConfiguration/guestConfigurationAssignments",
			ApiVersions:  []string{"2024-04-05"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.guestConfiguration",
			},
			StringRules: []azwise.StringRule{
				// AssignmentType enum — full ARM SDK set (PossibleValuesForAssignmentType).
				{
					PropertyPath:  "properties.guestConfiguration.assignmentType",
					AllowedValues: []string{"ApplyAndAutoCorrect", "ApplyAndMonitor", "Audit", "DeployAndAutoCorrect"},
					Message:       "assignment_type must be one of ApplyAndAutoCorrect, ApplyAndMonitor, Audit, DeployAndAutoCorrect",
				},
			},
		},
	}
}

func init() { azwise.Register(NewPolicyVirtualMachineConfigurationAssignment()) }
