// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PolicyAssignment provides resource knowledge for
// Microsoft.Authorization/policyAssignments.
//
// Merges the four azurerm policy-assignment resources that map to the same ARM type
// at different scopes (all share assignmentBaseResource):
//   - azurerm_management_group_policy_assignment (scope: management_group_id)
//   - azurerm_resource_group_policy_assignment   (scope: resource_group_id)
//   - azurerm_resource_policy_assignment          (scope: resource_id)
//   - azurerm_subscription_policy_assignment      (scope: subscription_id)
//
// Only knowledge true for EVERY contributing resource is included. The envelope
// `name` length differs by scope (management-group caps at 24; the rest at 64), so
// the name rule below carries only the universal parts (forbidden characters + not
// ending in space/period + non-empty); the scope-specific max lengths are omitted.
// The scope fields and `name` are ForceNew by construction (envelope), so they are
// not repeated as body ForceNew rules.
//
// Array-element enum validators (overrides[].kind, overrides[].selectors[].kind,
// resource_selectors[].selectors[].kind) target array elements and are skipped —
// azwise cannot address "[*]" element paths.
//
// Sources:
//   - terraform-provider-azurerm internal/services/policy/assignment_resource_base.go
//     shared arguments (policy_definition_id Required+ForceNew; enforce Default true;
//     createFunc maps enforce->enforcementMode, not_scopes->notScopes, etc.), 30m create
//   - terraform-provider-azurerm internal/services/policy/{management_group,resource_group,resource,subscription}_policy_assignment_resource.go
//     per-scope name validators (StringDoesNotContainAny("#<>%&:\\?/"),
//     StringLenBetween 1-24 / 1-64, trailing-char StringMatch "[^ .]$")
//   - go-azure-sdk resource-manager/resources/2022-06-01/policyassignments
//     PolicyAssignmentProperties: policyDefinitionId/displayName/description/
//     enforcementMode/metadata/notScopes/nonComplianceMessages/parameters/overrides/
//     resourceSelectors/scope; EnforcementMode enum Default/DoNotEnforce
type PolicyAssignment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PolicyAssignment)(nil)

// NewPolicyAssignment returns knowledge for the policyAssignments resource.
func NewPolicyAssignment() *PolicyAssignment {
	return &PolicyAssignment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Authorization/policyAssignments",
			ApiVersions:  []string{"2022-06-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.policyDefinitionId"},
			},
			RequiredFields: []string{
				"properties.policyDefinitionId",
			},
			// azurerm defaults `enforce` to true, which maps to enforcementMode "Default".
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.enforcementMode", Value: "Default"},
			},
			StringRules: []azwise.StringRule{
				// Universal envelope-name constraints shared by all four scopes:
				// no #<>%&:\?/ characters, must not end with a space or period,
				// at least one character. Per-scope max length is intentionally omitted.
				{
					PropertyPath: "",
					Regex:        "^[^#<>%&:\\\\?/]*[^#<>%&:\\\\?/ .]$",
					Message:      "policy assignment name must not contain any of #<>%&:\\?/ and must not end with a space or period",
				},
				// EnforcementMode enum — full ARM SDK set (PossibleValuesForEnforcementMode).
				{
					PropertyPath:  "properties.enforcementMode",
					AllowedValues: []string{"Default", "DoNotEnforce"},
					Message:       "enforcementMode must be Default or DoNotEnforce",
				},
			},
		},
	}
}

func init() { azwise.Register(NewPolicyAssignment()) }
