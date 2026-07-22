package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VpnGatewayConnection provides resource knowledge for
// Microsoft.Network/vpnGateways/vpnConnections.
//
// Mirrors azurerm_vpn_gateway_connection. name and vpn_gateway_id (parent) are
// envelope/parent references.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/vpn_gateway_connection_resource.go
//     (schema 45-337, expand 370-383, timeouts 38-43)
//   - go-azure-helpers commonids/vpn_connection.go:105 (segment casing)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualwans:
//     model_vpnconnectionproperties.go
//
// Not encoded (deliberate):
//   - vpn_link is a Required list expanding to properties.vpnLinkConnections[*];
//     its per-link rules (bandwidth, bgp, ipsec_policy enums, connection_mode, etc.)
//     live under an array element and cannot be lowered by azwise/azapin.
//   - traffic_selector_policy and the routing{} block are likewise nested/array
//     shapes; documented here, not emitted.
type VpnGatewayConnection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VpnGatewayConnection)(nil)

// NewVpnGatewayConnection returns knowledge for the vpnConnections resource.
func NewVpnGatewayConnection() *VpnGatewayConnection {
	return &VpnGatewayConnection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/vpnGateways/vpnConnections",
			ApiVersions:  []string{"2025-01-01"},
			// remote_vpn_site_id is ForceNew; internet_security and links are updatable.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.remoteVpnSite.id"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.enableInternetSecurity", Value: false},
			},
			RequiredFields: []string{
				"properties.remoteVpnSite.id",
			},
		},
	}
}

func init() { azwise.Register(NewVpnGatewayConnection()) }
