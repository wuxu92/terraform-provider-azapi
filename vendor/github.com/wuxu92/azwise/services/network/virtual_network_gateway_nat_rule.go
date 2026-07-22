package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualNetworkGatewayNatRule provides resource knowledge for
// Microsoft.Network/virtualNetworkGateways/natRules.
//
// Mirrors azurerm_virtual_network_gateway_nat_rule. name and
// virtual_network_gateway_id (parent) are envelope/parent references.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/virtual_network_gateway_nat_rule_resource.go
//     (schema 43-127, expand 157-168, timeouts 31-36)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualnetworkgateways:
//     id_virtualnetworkgatewaynatrule.go:110 (segment casing),
//     model_virtualnetworkgatewaynatruleproperties.go, constants.go (VpnNatRuleMode/Type)
type VirtualNetworkGatewayNatRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualNetworkGatewayNatRule)(nil)

// NewVirtualNetworkGatewayNatRule returns knowledge for the natRules resource.
func NewVirtualNetworkGatewayNatRule() *VirtualNetworkGatewayNatRule {
	return &VirtualNetworkGatewayNatRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/virtualNetworkGateways/natRules",
			ApiVersions:  []string{"2025-01-01"},
			// mode and type are ForceNew; mappings and ip_configuration_id are updatable.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.mode"},
				{PropertyPath: "properties.type"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.mode",
					AllowedValues: []string{"EgressSnat", "IngressSnat"},
					Message:       "mode must be EgressSnat or IngressSnat",
				},
				{
					PropertyPath:  "properties.type",
					AllowedValues: []string{"Static", "Dynamic"},
					Message:       "type must be Static or Dynamic",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.mode", Value: "EgressSnat"},
				{PropertyPath: "properties.type", Value: "Static"},
			},
			RequiredFields: []string{
				"properties.externalMappings",
				"properties.internalMappings",
			},
		},
	}
}

func init() { azwise.Register(NewVirtualNetworkGatewayNatRule()) }
