package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkManagerRoutingRule provides resource knowledge for
// Microsoft.Network/networkManagers/routingConfigurations/ruleCollections/rules.
//
// Mirrors azurerm_network_manager_routing_rule. name and the parent rule_collection_id
// are envelope-owned (Required + RequiresReplace).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_manager_routing_rule_resource.go
//     (schema Arguments, CustomizeDiff, expand functions)
//   - go-azure-sdk resource-manager/network/2025-01-01/routingrules:
//     model_routingrulepropertiesformat.go, model_routingruleroutedestination.go,
//     model_routingrulenexthop.go, constants.go
//
// Not encoded (deliberate):
//   - CustomizeDiff requires destination.address to be a valid CIDR when
//     destination.type is AddressPrefix, and next_hop.address to be set when
//     next_hop.type is VirtualAppliance — cross-field semantic rules that belong on an
//     azapin customizer, noted only.
type NetworkManagerRoutingRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkManagerRoutingRule)(nil)

func NewNetworkManagerRoutingRule() *NetworkManagerRoutingRule {
	return &NetworkManagerRoutingRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkManagers/routingConfigurations/ruleCollections/rules",
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
				{
					PropertyPath:  "properties.destination.type",
					AllowedValues: []string{"AddressPrefix", "ServiceTag"},
					Message:       "must be AddressPrefix or ServiceTag",
				},
				{
					PropertyPath:  "properties.nextHop.nextHopType",
					AllowedValues: []string{"Internet", "NoNextHop", "VirtualAppliance", "VirtualNetworkGateway", "VnetLocal"},
					Message:       "must be one of Internet, NoNextHop, VirtualAppliance, VirtualNetworkGateway, VnetLocal",
				},
			},
			// destination (with address + type) and next_hop (with type) are Required.
			RequiredFields: []string{
				"properties.destination.destinationAddress",
				"properties.destination.type",
				"properties.nextHop.nextHopType",
			},
		},
	}
}

func init() { azwise.Register(NewNetworkManagerRoutingRule()) }
