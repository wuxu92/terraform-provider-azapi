// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package machinelearning

import (
	"time"

	"github.com/wuxu92/azwise"
)

// OutboundRules provides resource knowledge for
// Microsoft.MachineLearningServices/workspaces/outboundRules.
//
// One ARM type, three AzureRM resources discriminated by `properties.type`
// (see managednetwork.OutboundRuleBasicResource.Properties, a discriminated union):
//   - azurerm_machine_learning_workspace_network_outbound_rule_fqdn             (type = "FQDN")
//   - azurerm_machine_learning_workspace_network_outbound_rule_private_endpoint  (type = "PrivateEndpoint")
//   - azurerm_machine_learning_workspace_network_outbound_rule_service_tag       (type = "ServiceTag")
//
// Universal knowledge encoded here:
//   - `name` and the parent workspace_id are ForceNew on all three (workspace_id is a
//     parent reference — excluded from body rules).
//   - `properties.type` is the discriminator; the RuleType enum below constrains it to
//     the full ARM SDK set and is valid for every contributing resource.
//
// NOTE (deliberately omitted):
//   - service_tag: AzureRM restricts it with a curated StringInSlice, but ARM models
//     ServiceTagDestination.serviceTag as a free `*string` (no SDK enum). Encoding the
//     AzureRM subset would reject legitimate ARM values AzAPI can send, so no rule.
//   - protocol {*, TCP, UDP, ICMP}: likewise a free `*string` in the ARM model — no rule.
//   - destination_fqdn / port_ranges are per-kind free strings — no universal rule.
//
// Sources:
//   - AzureRM internal/services/machinelearning/machine_learning_workspace_network_outbound_rule_fqdn_resource.go
//     (schema :45-68, Create :74-121, RuleTypeFQDN :106)
//   - AzureRM internal/services/machinelearning/machine_learning_workspace_network_outbound_rule_private_endpoint_resource.go
//     (RuleTypePrivateEndpoint)
//   - AzureRM internal/services/machinelearning/machine_learning_workspace_network_outbound_rule_service_tag_resource.go
//     (schema :46-158, Create :164-215, RuleTypeServiceTag :196, service_tag list :65-142)
//   - go-azure-sdk .../machinelearningservices/2025-06-01/managednetwork:
//     model_outboundrulebasicresource.go, model_outboundrule.go, model_fqdnoutboundrule.go,
//     model_servicetagoutboundrule.go, model_servicetagdestination.go,
//     model_privateendpointoutboundrule.go, constants.go (RuleType, RuleCategory)
type OutboundRules struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*OutboundRules)(nil)

func NewOutboundRules() *OutboundRules {
	return &OutboundRules{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.MachineLearningServices/workspaces/outboundRules",
			ApiVersions:  []string{"2025-06-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Discriminator (properties.type). Full ARM SDK RuleType set.
					PropertyPath:  "properties.type",
					AllowedValues: []string{"FQDN", "PrivateEndpoint", "ServiceTag"},
					Message:       "must be one of FQDN, PrivateEndpoint, or ServiceTag",
				},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewOutboundRules()) }
