// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PolicyExemption provides resource knowledge for
// Microsoft.Authorization/policyExemptions.
//
// Merges the four azurerm policy-exemption resources that map to the same ARM type
// at different scopes (all share an identical body schema):
//   - azurerm_management_group_policy_exemption (scope: management_group_id)
//   - azurerm_resource_group_policy_exemption   (scope: resource_group_id)
//   - azurerm_resource_policy_exemption          (scope: resource_id)
//   - azurerm_subscription_policy_exemption      (scope: subscription_id)
//
// All rules below are universal. Envelope fields (name, scope) are ForceNew by
// construction and are not repeated as body ForceNew rules.
//
// Sources:
//   - terraform-provider-azurerm internal/services/policy/{management_group,resource_group,resource,subscription}_policy_exemption_resource.go
//     schema (exemption_category Required enum; policy_assignment_id Required+ForceNew;
//     display_name len 1-128; description len 1-512; expires_on ISO8601), timeouts 30m/5m/30m/30m
//   - azure-sdk-for-go services/preview/resources/mgmt/2021-06-01-preview/policy
//     ExemptionProperties: policyAssignmentId/policyDefinitionReferenceIds/
//     exemptionCategory/expiresOn/displayName/description (json tags policyAssignmentId,
//     exemptionCategory, expiresOn, displayName, description); ExemptionCategory enum
//     Mitigated/Waiver
type PolicyExemption struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PolicyExemption)(nil)

// NewPolicyExemption returns knowledge for the policyExemptions resource.
func NewPolicyExemption() *PolicyExemption {
	return &PolicyExemption{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Authorization/policyExemptions",
			ApiVersions:  []string{"2021-06-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.policyAssignmentId"},
			},
			RequiredFields: []string{
				"properties.exemptionCategory",
				"properties.policyAssignmentId",
			},
			StringRules: []azwise.StringRule{
				// ExemptionCategory enum — full ARM set (azurerm StringInSlice matches it).
				{
					PropertyPath:  "properties.exemptionCategory",
					AllowedValues: []string{"Mitigated", "Waiver"},
					Message:       "exemption_category must be Mitigated or Waiver",
				},
				{PropertyPath: "properties.displayName", MinLength: 1, MaxLength: 128, Message: "display_name must be 1-128 characters"},
				{PropertyPath: "properties.description", MinLength: 1, MaxLength: 512, Message: "description must be 1-512 characters"},
			},
		},
	}
}

func init() { azwise.Register(NewPolicyExemption()) }
