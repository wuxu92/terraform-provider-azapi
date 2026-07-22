package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudGatewayCustomDomain provides resource knowledge for
// Microsoft.AppPlatform/Spring/gateways/domains.
//
// Mirrors azurerm_spring_cloud_gateway_custom_domain.
//
// Sources:
//   - internal/services/springcloud/spring_cloud_gateway_custom_domain_resource.go:34-70
//     (Timeouts :34-39 30m/5m/30m/30m; name :51-56 ForceNew; thumbprint :65-68)
//   - go-azure-sdk .../appplatform id_gatewaydomain.go, model_gatewaycustomdomainproperties.go
//     (Thumbprint json "thumbprint")
type SpringCloudGatewayCustomDomain struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudGatewayCustomDomain)(nil)

// NewSpringCloudGatewayCustomDomain returns knowledge for the gateway custom domain resource.
func NewSpringCloudGatewayCustomDomain() *SpringCloudGatewayCustomDomain {
	return &SpringCloudGatewayCustomDomain{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/gateways/domains",
			ApiVersions:  []string{"2024-01-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// name (the domain) is ForceNew; no declarative name pattern in AzureRM.
		},
	}
}

func init() { azwise.Register(NewSpringCloudGatewayCustomDomain()) }
