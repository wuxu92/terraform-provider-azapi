package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LocalNetworkGateway provides resource knowledge for
// Microsoft.Network/localNetworkGateways.
//
// Mirrors azurerm_local_network_gateway. name and resource_group_name are
// envelope-owned.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/local_network_gateway_resource.go
//     (schema 49-106, expand 139-151, timeouts 42-47)
//   - go-azure-sdk resource-manager/network/2025-01-01/localnetworkgateways:
//     model_localnetworkgatewaypropertiesformat.go, model_bgpsettings.go
type LocalNetworkGateway struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LocalNetworkGateway)(nil)

// NewLocalNetworkGateway returns knowledge for the localNetworkGateways resource.
func NewLocalNetworkGateway() *LocalNetworkGateway {
	return &LocalNetworkGateway{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/localNetworkGateways",
			ApiVersions:  []string{"2025-01-01"},
			// location replaces the gateway.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// gateway_address and gateway_fqdn are ExactlyOneOf.
			ExactlyOneOf: []azwise.RelationalRule{
				{
					Paths: []string{
						"properties.gatewayIpAddress",
						"properties.fqdn",
					},
					Message: "exactly one of gateway_address or gateway_fqdn must be set",
				},
			},
		},
	}
}

func init() { azwise.Register(NewLocalNetworkGateway()) }
