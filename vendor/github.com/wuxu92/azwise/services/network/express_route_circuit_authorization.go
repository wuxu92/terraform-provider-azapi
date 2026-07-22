package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ExpressRouteCircuitAuthorization provides resource knowledge for
// Microsoft.Network/expressRouteCircuits/authorizations.
//
// Mirrors azurerm_express_route_circuit_authorization. name, resource_group_name
// and express_route_circuit_name live on the operational envelope (all ForceNew by
// construction). The create body carries an empty AuthorizationPropertiesFormat;
// authorization_key and authorization_use_status are server-computed.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/express_route_circuit_authorization_resource.go
//     (schema resourceExpressRouteCircuitAuthorization, CRUD timeouts, Create body lines 92-94)
//   - go-azure-sdk resource-manager/network/2025-01-01/expressroutecircuitauthorizations:
//     id_authorization.go (ARM type casing "expressRouteCircuits/authorizations"),
//     model_authorizationpropertiesformat.go (json tags)
type ExpressRouteCircuitAuthorization struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ExpressRouteCircuitAuthorization)(nil)

// NewExpressRouteCircuitAuthorization returns knowledge for the authorizations resource.
func NewExpressRouteCircuitAuthorization() *ExpressRouteCircuitAuthorization {
	return &ExpressRouteCircuitAuthorization{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/expressRouteCircuits/authorizations",
			ApiVersions:  []string{"2025-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			SensitiveFields: []string{
				"properties.authorizationKey",
			},
			ComputedFields: []string{
				"properties.authorizationKey",
				"properties.authorizationUseStatus",
			},
		},
	}
}

func init() { azwise.Register(NewExpressRouteCircuitAuthorization()) }
