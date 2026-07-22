package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkManagerRoutingRuleCollection provides resource knowledge for
// Microsoft.Network/networkManagers/routingConfigurations/ruleCollections.
//
// Mirrors azurerm_network_manager_routing_rule_collection. name and the parent
// routing_configuration_id are envelope-owned (Required + RequiresReplace).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_manager_routing_rule_collection_resource.go
//     (schema Arguments, expandBgpRoutePropagation)
//   - go-azure-sdk resource-manager/network/2025-01-01/routingrulecollections:
//     model_routingrulecollectionpropertiesformat.go, constants.go
//
// Not encoded (deliberate):
//   - network_group_ids elements validate with a NetworkGroup resource-ID validator — a
//     semantic rule inside an array; belongs on an azapin customizer, noted only.
type NetworkManagerRoutingRuleCollection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkManagerRoutingRuleCollection)(nil)

func NewNetworkManagerRoutingRuleCollection() *NetworkManagerRoutingRuleCollection {
	return &NetworkManagerRoutingRuleCollection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkManagers/routingConfigurations/ruleCollections",
			ApiVersions:  []string{"2025-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name validation (empty PropertyPath = the name attribute).
					Regex:     `^[a-zA-Z0-9\_\.\-]{1,64}$`,
					MinLength: 1,
					MaxLength: 64,
					Message:   "must be 1-64 characters and contain only letters, numbers, underscores, periods and hyphens",
				},
			},
			// network_group_ids is Required (properties.appliesTo).
			RequiredFields: []string{"properties.appliesTo"},
			// bgp_route_propagation_enabled defaults false; AzureRM inverts this to
			// properties.disableBgpRoutePropagation = "True".
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.disableBgpRoutePropagation", Value: "True"},
			},
		},
	}
}

func init() { azwise.Register(NewNetworkManagerRoutingRuleCollection()) }
