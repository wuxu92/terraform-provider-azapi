package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ExpressRouteGateway provides resource knowledge for
// Microsoft.Network/expressRouteGateways.
//
// Mirrors azurerm_express_route_gateway. name and resource_group_name live on the
// operational envelope; only body/envelope-path knowledge is encoded here.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/express_route_gateway_resource.go
//     (schema resourceExpressRouteGateway, Create body lines 117-132, CRUD timeouts)
//   - go-azure-sdk resource-manager/network/2025-01-01/expressroutegateways:
//     id_expressroutegateway.go (ARM type casing "expressRouteGateways"),
//     model_expressroutegatewayproperties.go,
//     model_expressroutegatewaypropertiesautoscaleconfigurationbounds.go,
//     model_virtualhubid.go
//
// Not encoded (deliberate):
//   - scale_units (Required, IntBetween 1-10) maps into
//     properties.autoScaleConfiguration.bounds.min. The IntRule is emitted below.
//   - virtual_hub_id is a resource-ID validator (semantic) — ported to the customizer.
type ExpressRouteGateway struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ExpressRouteGateway)(nil)

// NewExpressRouteGateway returns knowledge for the expressRouteGateways resource.
func NewExpressRouteGateway() *ExpressRouteGateway {
	return &ExpressRouteGateway{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/expressRouteGateways",
			ApiVersions:  []string{"2025-01-01"},
			// location and virtual_hub_id are ForceNew in AzureRM (schema lines 54, 58-63).
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "properties.virtualHub.id"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 90 * time.Minute,
				Delete: 90 * time.Minute,
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.autoScaleConfiguration.bounds.min",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(10)),
					Message:      "scale_units must be between 1 and 10",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// allow_non_virtual_wan_traffic Default:false.
				{PropertyPath: "properties.allowNonVirtualWanTraffic", Value: false},
			},
			// scale_units (Required) -> autoScaleConfiguration.bounds.min;
			// virtual_hub_id (Required) -> properties.virtualHub.id.
			RequiredFields: []string{
				"properties.autoScaleConfiguration.bounds.min",
				"properties.virtualHub.id",
			},
		},
	}
}

func init() { azwise.Register(NewExpressRouteGateway()) }
