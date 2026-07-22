// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PolicySetDefinition provides resource knowledge for
// Microsoft.Authorization/policySetDefinitions.
//
// Merges two azurerm resources that map to the same ARM type at different scopes:
//   - azurerm_policy_set_definition (subscription / management-group via optional
//     management_group_id)
//   - azurerm_management_group_policy_set_definition (management-group scope; the
//     scope is the required management_group_id envelope parent)
//
// Both carry identical body schemas, so all rules below are universal (true for
// every contributing resource). Envelope fields (name, management_group_id) are
// ForceNew by construction and are not repeated as body ForceNew rules.
//
// `parameters` is conditionally ForceNew in azurerm (CustomizeDiff forces
// replacement only when parameter definitions are removed/reduced). azwise cannot
// express "fewer than before", so the declarative rule triggers on any change; the
// condition is noted for the maintainer.
//
// Sources:
//   - terraform-provider-azurerm internal/services/policy/policy_set_definition_resource.go
//     schema (name/policy_type/display_name Required; policy_type ForceNew;
//     policy_definition_reference Required; parameters conditional ForceNew)
//   - terraform-provider-azurerm internal/services/policy/management_group_policy_set_definition_resource.go
//     identical body schema at management-group scope
//   - go-azure-sdk resource-manager/resources/2025-01-01/policysetdefinitions
//     PolicySetDefinitionProperties: policyType/displayName/description/metadata/
//     parameters/policyDefinitions/policyDefinitionGroups (policyDefinitions is a
//     non-pointer Required slice); PolicyType enum BuiltIn/Custom/NotSpecified/Static
type PolicySetDefinition struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PolicySetDefinition)(nil)

// NewPolicySetDefinition returns knowledge for the policySetDefinitions resource.
func NewPolicySetDefinition() *PolicySetDefinition {
	return &PolicySetDefinition{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Authorization/policySetDefinitions",
			ApiVersions:  []string{"2025-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.policyType"},
				// Conditional in azurerm: replacement only when `parameters` are
				// removed/reduced. Declared here as an unconditional path change.
				{PropertyPath: "properties.parameters"},
			},
			RequiredFields: []string{
				"properties.policyType",
				"properties.displayName",
				"properties.policyDefinitions",
			},
			StringRules: []azwise.StringRule{
				// PolicyType enum — full ARM SDK set (PossibleValuesForPolicyType).
				{
					PropertyPath:  "properties.policyType",
					AllowedValues: []string{"BuiltIn", "Custom", "NotSpecified", "Static"},
					Message:       "policy_type must be one of BuiltIn, Custom, NotSpecified, Static",
				},
			},
		},
	}
}

func init() { azwise.Register(NewPolicySetDefinition()) }
