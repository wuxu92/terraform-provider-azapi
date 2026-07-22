package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ExpressRouteCircuitConnection provides resource knowledge for
// Microsoft.Network/expressRouteCircuits/peerings/connections.
//
// Mirrors azurerm_express_route_circuit_connection. The connection is parented to a
// circuit peering: the SDK PeeringConnectionId is
// .../expressRouteCircuits/{circuit}/peerings/{peering}/connections/{connection},
// so the ARM type nests under peerings. name, peering_id and peer_peering_id are
// envelope/parent references (all ForceNew) and are not repeated here.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/express_route_circuit_connection_resource.go
//     (schema resourceExpressRouteCircuitConnection, Create body lines 121-153, CRUD timeouts)
//   - go-azure-sdk resource-manager/network/2025-01-01/expressroutecircuitconnections:
//     id_peeringconnection.go (ARM parenting "expressRouteCircuits/peerings/connections",
//     ID() line 116), model_expressroutecircuitconnectionpropertiesformat.go,
//     model_ipv6circuitconnectionconfig.go
//
// Not encoded (deliberate):
//   - address_prefix_ipv4 is Required + ForceNew and CIDR-validated; the CIDR check is
//     a semantic validator ported to the azapin customizer, not a declarative rule.
//   - authorization_key is IsUUID-validated (semantic) — ported to the customizer.
type ExpressRouteCircuitConnection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ExpressRouteCircuitConnection)(nil)

// NewExpressRouteCircuitConnection returns knowledge for the connections resource.
func NewExpressRouteCircuitConnection() *ExpressRouteCircuitConnection {
	return &ExpressRouteCircuitConnection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/expressRouteCircuits/peerings/connections",
			ApiVersions:  []string{"2025-01-01"},
			// address_prefix_ipv4 is ForceNew (schema lines 67-72).
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.addressPrefix"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// address_prefix_ipv4 -> properties.addressPrefix is Required.
			RequiredFields: []string{
				"properties.addressPrefix",
			},
			SensitiveFields: []string{
				"properties.authorizationKey",
			},
		},
	}
}

func init() { azwise.Register(NewExpressRouteCircuitConnection()) }
