package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ExpressRouteConnection provides resource knowledge for
// Microsoft.Network/expressRouteGateways/expressRouteConnections.
//
// Mirrors azurerm_express_route_connection. The connection is parented to an
// ExpressRoute gateway: the SDK ExpressRouteConnectionId is
// .../expressRouteGateways/{gateway}/expressRouteConnections/{connection}. name,
// express_route_gateway_id and express_route_circuit_peering_id are envelope/parent
// references (all ForceNew) and are not repeated here.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/express_route_connection_resource.go
//     (schema resourceExpressRouteConnection, Create body lines 214-229, CRUD timeouts)
//   - go-azure-sdk resource-manager/network/2025-01-01/expressrouteconnections:
//     id_expressrouteconnection.go (ARM parenting "expressRouteGateways/expressRouteConnections",
//     ID() line 110), model_expressrouteconnectionproperties.go, model_routingconfiguration.go
//
// Not encoded (deliberate):
//   - authorization_key is IsUUID-validated (semantic) — ported to the customizer.
//   - express_route_circuit_peering_id is Required but is the parent SubResource
//     reference (properties.expressRouteCircuitPeering.id) supplied via the config;
//     kept as a RequiredField.
type ExpressRouteConnection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ExpressRouteConnection)(nil)

// NewExpressRouteConnection returns knowledge for the expressRouteConnections resource.
func NewExpressRouteConnection() *ExpressRouteConnection {
	return &ExpressRouteConnection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/expressRouteGateways/expressRouteConnections",
			ApiVersions:  []string{"2025-01-01"},
			// express_route_circuit_peering_id is ForceNew (schema lines 54-59).
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.expressRouteCircuitPeering.id"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.routingWeight",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(32000)),
					Message:      "routing_weight must be between 0 and 32000",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// internet_security_enabled Default:false.
				{PropertyPath: "properties.enableInternetSecurity", Value: false},
				// express_route_gateway_bypass_enabled Default:false.
				{PropertyPath: "properties.expressRouteGatewayBypass", Value: false},
				// routing_weight Default:0.
				{PropertyPath: "properties.routingWeight", Value: int64(0)},
			},
			// express_route_circuit_peering_id -> properties.expressRouteCircuitPeering.id
			// is Required.
			RequiredFields: []string{
				"properties.expressRouteCircuitPeering.id",
			},
		},
	}
}

func init() { azwise.Register(NewExpressRouteConnection()) }
