// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package policy

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PolicyDefinition provides resource knowledge for
// Microsoft.Authorization/policyDefinitions.
//
// Mirrors azurerm_policy_definition. A single ARM type serves both subscription
// and management-group scopes: azurerm's `management_group_id` selects the scope
// (envelope parent_id) rather than a distinct ARM type, so there is no separate
// management-group resource to merge. (Policy *set* definitions, by contrast, live
// under Microsoft.Authorization/policySetDefinitions — see policy_set_definition.go.)
//
// Envelope fields (name, management_group_id) are ForceNew by construction and are
// not repeated as body ForceNew rules here.
//
// `parameters` is conditionally ForceNew in azurerm: a CustomizeDiff forces
// replacement only when parameter definitions are *removed* (the new set has fewer
// entries than the old, or is emptied). azwise cannot express "fewer than before",
// so the declarative rule below triggers replacement on any parameters change; the
// condition is noted for the maintainer.
//
// Sources:
//   - terraform-provider-azurerm internal/services/policy/policy_definition_resource.go
//     schema (name/policy_type/mode/display_name Required; policy_type ForceNew;
//     parameters conditional ForceNew via CustomizeDiff), timeouts 30m/5m/30m/30m
//   - azure-sdk-for-go services/preview/resources/mgmt/2021-06-01-preview/policy
//     DefinitionProperties: policyType/mode/displayName/description/policyRule/
//     metadata/parameters (json tags policyType, mode, displayName, description,
//     policyRule, metadata, parameters); Type enum BuiltIn/Custom/NotSpecified/Static
type PolicyDefinition struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PolicyDefinition)(nil)

// NewPolicyDefinition returns knowledge for the policyDefinitions resource.
func NewPolicyDefinition() *PolicyDefinition {
	return &PolicyDefinition{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Authorization/policyDefinitions",
			ApiVersions:  []string{"2021-06-01-preview"},
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
				"properties.mode",
			},
			StringRules: []azwise.StringRule{
				// policy.Type enum — full ARM set (azurerm StringInSlice matches it).
				{
					PropertyPath:  "properties.policyType",
					AllowedValues: []string{"BuiltIn", "Custom", "NotSpecified", "Static"},
					Message:       "policy_type must be one of BuiltIn, Custom, NotSpecified, Static",
				},
				// azurerm restricts `mode` to this documented list via StringInSlice.
				// ARM models mode as an open string; new data-plane modes may appear,
				// so this mirrors azurerm rather than an SDK enum.
				{
					PropertyPath: "properties.mode",
					AllowedValues: []string{
						"All",
						"Indexed",
						"Microsoft.ContainerService.Data",
						"Microsoft.CustomerLockbox.Data",
						"Microsoft.DataCatalog.Data",
						"Microsoft.KeyVault.Data",
						"Microsoft.Kubernetes.Data",
						"Microsoft.MachineLearningServices.Data",
						"Microsoft.Network.Data",
						"Microsoft.Synapse.Data",
					},
					Message: "mode must be a supported policy definition mode",
				},
			},
		},
	}
}

func init() { azwise.Register(NewPolicyDefinition()) }
