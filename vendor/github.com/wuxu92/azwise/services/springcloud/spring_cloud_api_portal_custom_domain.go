package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudAPIPortalCustomDomain provides resource knowledge for
// Microsoft.AppPlatform/Spring/apiPortals/domains.
//
// Mirrors azurerm_spring_cloud_api_portal_custom_domain.
//
// Sources:
//   - internal/services/springcloud/spring_cloud_api_portal_custom_domain_resource.go:24-70
//     (Timeouts :38-43 30m/5m/30m/30m; name :51-55 ForceNew; thumbprint :64-67)
//   - go-azure-sdk .../appplatform id_apiportaldomain.go, model_apiportalcustomdomainproperties.go
//     (Thumbprint json "thumbprint")
type SpringCloudAPIPortalCustomDomain struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudAPIPortalCustomDomain)(nil)

// NewSpringCloudAPIPortalCustomDomain returns knowledge for the API portal custom domain resource.
func NewSpringCloudAPIPortalCustomDomain() *SpringCloudAPIPortalCustomDomain {
	return &SpringCloudAPIPortalCustomDomain{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/apiPortals/domains",
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

func init() { azwise.Register(NewSpringCloudAPIPortalCustomDomain()) }
