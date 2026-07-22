package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualNetworkGatewayConnection provides resource knowledge for
// Microsoft.Network/connections.
//
// Mirrors azurerm_virtual_network_gateway_connection. name and resource_group_name
// are envelope-owned. The ARM resource type is Microsoft.Network/connections (the
// SDK package is virtualnetworkgatewayconnections, but the resource ID segment is
// /providers/Microsoft.Network/connections/{name}).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/virtual_network_gateway_connection_resource.go
//     (schema 50-353, expand 780-872, timeouts 43-48)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualnetworkgatewayconnections:
//     id_connection.go:104 (segment casing), model_virtualnetworkgatewayconnectionpropertiesformat.go,
//     constants.go (ConnectionType 702-706, ConnectionProtocol 614-617, ConnectionMode 571-573)
//
// Not encoded (deliberate):
//   - traffic_selector_policy / custom_bgp_addresses / ipsec_policy are array/nested
//     shapes whose per-element enums cannot be lowered by azwise/azapin.
type VirtualNetworkGatewayConnection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualNetworkGatewayConnection)(nil)

// NewVirtualNetworkGatewayConnection returns knowledge for the connections resource.
func NewVirtualNetworkGatewayConnection() *VirtualNetworkGatewayConnection {
	return &VirtualNetworkGatewayConnection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/connections",
			ApiVersions:  []string{"2025-01-01"},
			// location, type, gateway refs, dpd_timeout, circuit, peer gateway,
			// local_azure_ip_address, connection_protocol and connection_mode all
			// replace the connection.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "properties.connectionType"},
				{PropertyPath: "properties.virtualNetworkGateway1.id"},
				{PropertyPath: "properties.dpdTimeoutSeconds"},
				{PropertyPath: "properties.peer.id"},
				{PropertyPath: "properties.virtualNetworkGateway2.id"},
				{PropertyPath: "properties.useLocalAzureIpAddress"},
				{PropertyPath: "properties.connectionProtocol"},
				{PropertyPath: "properties.connectionMode"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			SensitiveFields: []string{
				"properties.sharedKey",
				"properties.authorizationKey",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.connectionType",
					AllowedValues: []string{"ExpressRoute", "IPsec", "Vnet2Vnet", "VPNClient"},
					Message:       "type must be ExpressRoute, IPsec, Vnet2Vnet or VPNClient",
				},
				{
					PropertyPath:  "properties.connectionProtocol",
					AllowedValues: []string{"IKEv1", "IKEv2"},
					Message:       "connection_protocol must be IKEv1 or IKEv2",
				},
				{
					PropertyPath:  "properties.connectionMode",
					AllowedValues: []string{"Default", "InitiatorOnly", "ResponderOnly"},
					Message:       "connection_mode must be Default, InitiatorOnly or ResponderOnly",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.routingWeight",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(32000)),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.enableBgp", Value: false},
				{PropertyPath: "properties.enablePrivateLinkFastPath", Value: false},
				{PropertyPath: "properties.connectionMode", Value: "Default"},
			},
			RequiredFields: []string{
				"properties.connectionType",
				"properties.virtualNetworkGateway1.id",
			},
		},
	}
}

func init() { azwise.Register(NewVirtualNetworkGatewayConnection()) }
