package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualNetworkPeering provides resource knowledge for
// Microsoft.Network/virtualNetworks/virtualNetworkPeerings.
//
// Mirrors azurerm_virtual_network_peering. name, resource_group_name and
// virtual_network_name (parent) are envelope/parent references.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/virtual_network_peering_resource.go
//     (schema 49-133, expand 159-181, timeouts 42-47)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualnetworkpeerings:
//     id_virtualnetworkpeering.go:110 (segment casing),
//     model_virtualnetworkpeeringpropertiesformat.go
//
// Not encoded (deliberate):
//   - triggers is a TF-only map with no ARM body equivalent; skipped.
type VirtualNetworkPeering struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualNetworkPeering)(nil)

// NewVirtualNetworkPeering returns knowledge for the virtualNetworkPeerings resource.
func NewVirtualNetworkPeering() *VirtualNetworkPeering {
	return &VirtualNetworkPeering{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/virtualNetworks/virtualNetworkPeerings",
			ApiVersions:  []string{"2025-01-01"},
			// remote_virtual_network_id, only_ipv6_peering_enabled and
			// peer_complete_virtual_networks_enabled replace the peering.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.remoteVirtualNetwork.id"},
				{PropertyPath: "properties.enableOnlyIPv6Peering"},
				{PropertyPath: "properties.peerCompleteVnets"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.allowVirtualNetworkAccess", Value: true},
				{PropertyPath: "properties.allowForwardedTraffic", Value: false},
				{PropertyPath: "properties.allowGatewayTransit", Value: false},
				{PropertyPath: "properties.peerCompleteVnets", Value: true},
				{PropertyPath: "properties.useRemoteGateways", Value: false},
			},
			RequiredFields: []string{
				"properties.remoteVirtualNetwork.id",
			},
		},
	}
}

func init() { azwise.Register(NewVirtualNetworkPeering()) }
