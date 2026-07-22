package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudGatewayRouteConfig provides resource knowledge for
// Microsoft.AppPlatform/Spring/gateways/routeConfigs.
//
// Mirrors azurerm_spring_cloud_gateway_route_config.
//
// Sources:
//   - internal/services/springcloud/spring_cloud_gateway_route_config_resource.go:39-170
//     (Timeouts :39-44 30m/5m/30m/30m; name :52-56 ForceNew; protocol :80-87 Required enum;
//     route.order :118-121 Required)
//   - go-azure-sdk .../appplatform model_gatewayrouteconfigproperties.go (Protocol json
//     "protocol", AppResourceId "appResourceId"), constants.go GatewayRouteConfigProtocol
//     (HTTP/HTTPS)
type SpringCloudGatewayRouteConfig struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudGatewayRouteConfig)(nil)

// NewSpringCloudGatewayRouteConfig returns knowledge for the gateway route config resource.
func NewSpringCloudGatewayRouteConfig() *SpringCloudGatewayRouteConfig {
	return &SpringCloudGatewayRouteConfig{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/gateways/routeConfigs",
			ApiVersions:  []string{"2024-01-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{"properties.protocol"},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.protocol",
					AllowedValues: []string{"HTTP", "HTTPS"},
				},
			},
		},
	}
}

func init() { azwise.Register(NewSpringCloudGatewayRouteConfig()) }
