package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PointToSiteVpnGateway provides resource knowledge for
// Microsoft.Network/p2sVpnGateways.
//
// Mirrors azurerm_point_to_site_vpn_gateway. name and resource_group_name are
// envelope-owned.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/point_to_site_vpn_gateway_resource.go
//     (schema 47-188, expand 214-231, timeouts 40-45)
//   - go-azure-helpers commonids/virtual_wan_p2s_vpn_gateway.go:99 (segment casing)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualwans:
//     model_p2svpngatewayproperties.go
//
// Not encoded (deliberate):
//   - connection_configuration (Required) expands to
//     properties.p2SConnectionConfigurations[*]; its nested address pool / routing
//     rules live under array elements and cannot be lowered by azwise/azapin.
type PointToSiteVpnGateway struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PointToSiteVpnGateway)(nil)

// NewPointToSiteVpnGateway returns knowledge for the p2sVpnGateways resource.
func NewPointToSiteVpnGateway() *PointToSiteVpnGateway {
	return &PointToSiteVpnGateway{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/p2sVpnGateways",
			ApiVersions:  []string{"2025-01-01"},
			// location, virtual_hub_id, vpn_server_configuration_id and
			// routing_preference_internet_enabled all replace the gateway.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "properties.virtualHub.id"},
				{PropertyPath: "properties.vpnServerConfiguration.id"},
				{PropertyPath: "properties.isRoutingPreferenceInternet"},
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
				{PropertyPath: "properties.isRoutingPreferenceInternet", Value: false},
			},
			RequiredFields: []string{
				"properties.virtualHub.id",
				"properties.vpnServerConfiguration.id",
				"properties.vpnGatewayScaleUnit",
				"properties.p2SConnectionConfigurations",
			},
		},
	}
}

func init() { azwise.Register(NewPointToSiteVpnGateway()) }
