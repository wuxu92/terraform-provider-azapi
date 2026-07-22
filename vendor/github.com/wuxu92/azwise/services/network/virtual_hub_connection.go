package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualHubConnection provides resource knowledge for
// Microsoft.Network/virtualHubs/hubVirtualNetworkConnections.
//
// Mirrors azurerm_virtual_hub_connection. name, virtual_hub_id (parent) are
// envelope/parent references.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/virtual_hub_connection_resource.go
//     (schema 50-188, expand 230-237, timeouts 37-42)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualwans:
//     id_hubvirtualnetworkconnection.go:110 (segment casing),
//     model_hubvirtualnetworkconnectionproperties.go
//
// Not encoded (deliberate):
//   - the routing{} block (properties.routingConfiguration) carries AzureRM
//     AtLeastOneOf constraints on nested sub-objects (associated_route_table_id /
//     propagated_route_table / static_vnet_route, and labels / route_table_ids).
//     These live under a single (MaxItems:1) object but the propagated_route_table
//     labels/route_table_ids constraint is only meaningful when that nested object
//     is present; representing it as a resource-level RelationalRule would misfire
//     when routing is absent, so it is documented here, not emitted.
type VirtualHubConnection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualHubConnection)(nil)

// NewVirtualHubConnection returns knowledge for the hubVirtualNetworkConnections resource.
func NewVirtualHubConnection() *VirtualHubConnection {
	return &VirtualHubConnection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/virtualHubs/hubVirtualNetworkConnections",
			ApiVersions:  []string{"2025-01-01"},
			// remote_virtual_network_id is ForceNew; only internet_security and
			// routing are updatable.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.remoteVirtualNetwork.id"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.enableInternetSecurity", Value: false},
			},
			RequiredFields: []string{
				"properties.remoteVirtualNetwork.id",
			},
		},
	}
}

func init() { azwise.Register(NewVirtualHubConnection()) }
