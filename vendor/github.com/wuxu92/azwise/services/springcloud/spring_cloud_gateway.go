package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudGateway provides resource knowledge for
// Microsoft.AppPlatform/Spring/gateways.
//
// Mirrors azurerm_spring_cloud_gateway.
//
// Sources:
//   - internal/services/springcloud/spring_cloud_gateway_resource.go:107-250
//     (Timeouts Create :445 30m / Update :513 30m / Read :608 5m / Delete :672 30m;
//     name :108-115 ForceNew "default"; application_performance_monitoring_types :173-186)
//   - go-azure-sdk .../appplatform model_gatewayproperties.go (ApmTypes json "apmTypes",
//     HTTPSOnly "httpsOnly", Public "public"), constants.go ApmType
type SpringCloudGateway struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudGateway)(nil)

// NewSpringCloudGateway returns knowledge for the Spring Cloud gateway resource.
func NewSpringCloudGateway() *SpringCloudGateway {
	return &SpringCloudGateway{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/gateways",
			ApiVersions:  []string{"2024-01-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// name must be "default".
				{
					PropertyPath:  "",
					AllowedValues: []string{"default"},
				},
			},
			// application_performance_monitoring_types maps to properties.apmTypes[*] (array
			// element enum, values AppDynamics/ApplicationInsights/Dynatrace/ElasticAPM/NewRelic);
			// array-element paths are not expressible as declarative StringRules.
		},
	}
}

func init() { azwise.Register(NewSpringCloudGateway()) }
