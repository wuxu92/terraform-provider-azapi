package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualHubBgpConnection provides resource knowledge for
// Microsoft.Network/virtualHubs/bgpConnections.
//
// Mirrors azurerm_virtual_hub_bgp_connection. name, virtual_hub_id (parent) are
// envelope/parent references. Nearly every field is ForceNew (create+delete only).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/virtual_hub_bgp_connection_resource.go
//     (schema 45-80, expand 112-124, timeouts 33-37)
//   - go-azure-helpers commonids/virtual_hub_bgp_connection.go:118-121 (segment casing)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualwans:
//     model_bgpconnectionproperties.go
type VirtualHubBgpConnection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualHubBgpConnection)(nil)

// NewVirtualHubBgpConnection returns knowledge for the bgpConnections resource.
func NewVirtualHubBgpConnection() *VirtualHubBgpConnection {
	return &VirtualHubBgpConnection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/virtualHubs/bgpConnections",
			ApiVersions:  []string{"2025-01-01"},
			// The resource has no Update func: peer_asn, peer_ip and
			// virtual_network_connection_id are all ForceNew.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.peerAsn"},
				{PropertyPath: "properties.peerIp"},
				{PropertyPath: "properties.hubVirtualNetworkConnection.id"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.peerAsn",
					MinValue:     azwise.Ptr(int64(0)),
				},
			},
			RequiredFields: []string{
				"properties.peerAsn",
				"properties.peerIp",
			},
		},
	}
}

func init() { azwise.Register(NewVirtualHubBgpConnection()) }
