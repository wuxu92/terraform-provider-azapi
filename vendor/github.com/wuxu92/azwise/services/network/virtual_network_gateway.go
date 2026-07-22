package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualNetworkGateway provides resource knowledge for
// Microsoft.Network/virtualNetworkGateways.
//
// Mirrors azurerm_virtual_network_gateway. name and resource_group_name are
// envelope-owned.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/virtual_network_gateway_resource.go
//     (schema 54-672, expand 1046-1075, timeouts 45-50)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualnetworkgateways:
//     model_virtualnetworkgatewaypropertiesformat.go, model_virtualnetworkgatewaysku.go,
//     constants.go (VirtualNetworkGatewayType 1374-1378, VpnType 1676-1679,
//     VpnGatewayGeneration 1506-1510)
//
// Not encoded (deliberate):
//   - sku uses a composite Any() validator whose valid set depends on `type` and
//     `vpn_type` (validateVirtualNetworkGateway{PolicyBased,RouteBased,ExpressRoute}Sku);
//     this cross-field semantic rule is not representable declaratively.
//   - sku ForceNew is CONDITIONAL: only changes that cross the availability-zone
//     boundary replace the gateway (CustomizeDiff, resource line ~1003), so no
//     unconditional ForceNew is emitted for the sku.
//   - ip_configuration / policy_group / vpn_client_configuration / bgp_settings are
//     array/nested shapes whose per-element rules cannot be lowered by azwise/azapin.
type VirtualNetworkGateway struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualNetworkGateway)(nil)

// NewVirtualNetworkGateway returns knowledge for the virtualNetworkGateways resource.
func NewVirtualNetworkGateway() *VirtualNetworkGateway {
	return &VirtualNetworkGateway{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/virtualNetworkGateways",
			ApiVersions:  []string{"2025-01-01"},
			// location, edge_zone, type, vpn_type, generation,
			// private_ip_address_enabled and the whole ip_configuration array all
			// replace the gateway.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "extendedLocation"},
				{PropertyPath: "properties.gatewayType"},
				{PropertyPath: "properties.vpnType"},
				{PropertyPath: "properties.vpnGatewayGeneration"},
				{PropertyPath: "properties.enablePrivateIpAddress"},
				{PropertyPath: "properties.ipConfigurations"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 120 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.gatewayType",
					AllowedValues: []string{"ExpressRoute", "Vpn", "LocalGateway"},
					Message:       "type must be ExpressRoute, Vpn or LocalGateway",
				},
				{
					PropertyPath:  "properties.vpnType",
					AllowedValues: []string{"RouteBased", "PolicyBased"},
					Message:       "vpn_type must be RouteBased or PolicyBased",
				},
				{
					PropertyPath:  "properties.vpnGatewayGeneration",
					AllowedValues: []string{"Generation1", "Generation2", "None"},
					Message:       "generation must be Generation1, Generation2 or None",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.autoScaleConfiguration.bounds.min",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(40)),
				},
				{
					PropertyPath: "properties.autoScaleConfiguration.bounds.max",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(40)),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.vpnType", Value: "RouteBased"},
				{PropertyPath: "properties.enableBgp", Value: false},
				{PropertyPath: "properties.enableBgpRouteTranslationForNat", Value: false},
				// ip_sec_replay_protection_enabled defaults true, mapped to the
				// inverse ARM flag disableIPSecReplayProtection.
				{PropertyPath: "properties.disableIPSecReplayProtection", Value: false},
				{PropertyPath: "properties.allowRemoteVnetTraffic", Value: false},
				{PropertyPath: "properties.allowVirtualWanTraffic", Value: false},
			},
			RequiredFields: []string{
				"properties.gatewayType",
				"properties.sku.name",
				"properties.ipConfigurations",
			},
			// maximum_scale_unit and minimum_scale_unit are RequiredWith each other;
			// both live under the same autoScaleConfiguration.bounds object.
			RequiredWith: []azwise.RelationalRule{
				{
					Paths: []string{
						"properties.autoScaleConfiguration.bounds.max",
						"properties.autoScaleConfiguration.bounds.min",
					},
					Message: "maximum_scale_unit and minimum_scale_unit must be set together",
				},
			},
		},
	}
}

func init() { azwise.Register(NewVirtualNetworkGateway()) }
