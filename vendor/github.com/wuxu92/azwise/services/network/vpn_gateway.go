package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VpnGateway provides resource knowledge for Microsoft.Network/vpnGateways.
//
// Mirrors azurerm_vpn_gateway. name and resource_group_name are envelope-owned.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/vpn_gateway_resource.go
//     (schema 48-222, expand 250-260, timeouts 41-46)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualwans:
//     model_vpngatewayproperties.go, model_bgpsettings.go
//
// Not encoded (deliberate):
//   - routing_preference is a TF-only "Microsoft Network"/"Internet" enum that maps
//     to the ARM bool properties.isRoutingPreferenceInternet (true == "Internet"),
//     so no StringRule is emitted; its ForceNew + default are captured on the bool.
type VpnGateway struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VpnGateway)(nil)

// NewVpnGateway returns knowledge for the vpnGateways resource.
func NewVpnGateway() *VpnGateway {
	return &VpnGateway{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/vpnGateways",
			ApiVersions:  []string{"2025-01-01"},
			// location, virtual_hub_id, routing_preference and the bgp asn/peer_weight
			// all replace the gateway.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "properties.virtualHub.id"},
				{PropertyPath: "properties.isRoutingPreferenceInternet"},
				{PropertyPath: "properties.bgpSettings.asn"},
				{PropertyPath: "properties.bgpSettings.peerWeight"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 90 * time.Minute,
				Delete: 90 * time.Minute,
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.vpnGatewayScaleUnit",
					MinValue:     azwise.Ptr(int64(0)),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.enableBgpRouteTranslationForNat", Value: false},
				{PropertyPath: "properties.vpnGatewayScaleUnit", Value: float64(1)},
				{PropertyPath: "properties.isRoutingPreferenceInternet", Value: false},
			},
			RequiredFields: []string{
				"properties.virtualHub.id",
			},
		},
	}
}

func init() { azwise.Register(NewVpnGateway()) }
