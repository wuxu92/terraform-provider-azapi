package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ExpressRoutePortAuthorization provides resource knowledge for
// Microsoft.Network/expressRoutePorts/authorizations.
//
// Mirrors azurerm_express_route_port_authorization. The authorization is parented to
// an ExpressRoute port: the SDK ExpressRoutePortAuthorizationId is
// .../expressRoutePorts/{port}/authorizations/{authorization}. name,
// resource_group_name and express_route_port_name live on the operational envelope
// (all ForceNew). The create body carries an empty
// ExpressRoutePortAuthorizationPropertiesFormat; authorization_key and
// authorization_use_status are server-computed.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/express_route_port_authorization_resource.go
//     (schema resourceExpressRoutePortAuthorization, Create body lines 90-92, CRUD timeouts)
//   - go-azure-sdk resource-manager/network/2025-01-01/expressrouteportauthorizations:
//     id_expressrouteportauthorization.go (ARM parenting
//     "expressRoutePorts/authorizations", ID() line 110),
//     model_expressrouteportauthorizationpropertiesformat.go
type ExpressRoutePortAuthorization struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ExpressRoutePortAuthorization)(nil)

// NewExpressRoutePortAuthorization returns knowledge for the port authorizations resource.
func NewExpressRoutePortAuthorization() *ExpressRoutePortAuthorization {
	return &ExpressRoutePortAuthorization{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/expressRoutePorts/authorizations",
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

func init() { azwise.Register(NewExpressRoutePortAuthorization()) }
