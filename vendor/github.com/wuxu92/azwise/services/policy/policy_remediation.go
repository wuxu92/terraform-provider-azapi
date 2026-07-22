// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PolicyRemediation provides resource knowledge for
// Microsoft.PolicyInsights/remediations.
//
// Merges the four azurerm policy-remediation resources that map to the same ARM type
// at different scopes (all share readRemediationProperties):
//   - azurerm_management_group_policy_remediation (scope: management_group_id)
//   - azurerm_resource_group_policy_remediation   (scope: resource_group_id)
//   - azurerm_resource_policy_remediation          (scope: resource_id)
//   - azurerm_subscription_policy_remediation      (scope: subscription_id)
//
// Only knowledge true for EVERY contributing resource is included. `policy_assignment_id`
// is Required but NOT ForceNew. Envelope fields (name, scope) are ForceNew by
// construction and are not repeated as body ForceNew rules.
//
// `resource_discovery_mode` is present on the subscription/resource-group/resource
// resources but NOT on the management-group resource, so its azurerm Default
// (ExistingNonCompliant) is non-universal and is intentionally omitted from
// DefaultValues (it would inject a value the management-group kind never sets). The
// enum constraint is still safe: it fires only when the property is present.
//
// Sources:
//   - terraform-provider-azurerm internal/services/policy/{management_group,resource_group,resource,subscription}_policy_remediation_resource.go
//     schema (policy_assignment_id Required; failure_percentage FloatBetween(0,1.0);
//     parallel_deployments/resource_count IntPositive; resource_discovery_mode enum
//     w/ Default ExistingNonCompliant except management-group), timeouts 30m/5m/30m/30m
//   - go-azure-sdk resource-manager/policyinsights/2021-10-01/remediations
//     RemediationProperties: policyAssignmentId/policyDefinitionReferenceId/
//     resourceDiscoveryMode/parallelDeployments(*int64)/resourceCount(*int64);
//     failureThreshold.percentage(*float64); filters.locations([]string);
//     ResourceDiscoveryMode enum ExistingNonCompliant/ReEvaluateCompliance
type PolicyRemediation struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PolicyRemediation)(nil)

// NewPolicyRemediation returns knowledge for the remediations resource.
func NewPolicyRemediation() *PolicyRemediation {
	return &PolicyRemediation{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.PolicyInsights/remediations",
			ApiVersions:  []string{"2021-10-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.policyAssignmentId",
			},
			FloatRules: []azwise.FloatRule{
				{
					PropertyPath: "properties.failureThreshold.percentage",
					MinValue:     azwise.Ptr(0.0),
					MaxValue:     azwise.Ptr(1.0),
					Message:      "failure_percentage must be between 0 and 1.0",
				},
			},
			IntRules: []azwise.IntRule{
				// azurerm IntPositive => strictly greater than 0 => minimum 1.
				{PropertyPath: "properties.parallelDeployments", MinValue: azwise.Ptr(int64(1)), Message: "parallel_deployments must be a positive integer"},
				{PropertyPath: "properties.resourceCount", MinValue: azwise.Ptr(int64(1)), Message: "resource_count must be a positive integer"},
			},
			StringRules: []azwise.StringRule{
				// ResourceDiscoveryMode enum — full ARM SDK set. Fires only when the
				// property is present (non-universal across scopes, so no DefaultValue).
				{
					PropertyPath:  "properties.resourceDiscoveryMode",
					AllowedValues: []string{"ExistingNonCompliant", "ReEvaluateCompliance"},
					Message:       "resource_discovery_mode must be ExistingNonCompliant or ReEvaluateCompliance",
				},
			},
		},
	}
}

func init() { azwise.Register(NewPolicyRemediation()) }
