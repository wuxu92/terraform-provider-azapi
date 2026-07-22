package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudAPIPortal provides resource knowledge for
// Microsoft.AppPlatform/Spring/apiPortals.
//
// Mirrors azurerm_spring_cloud_api_portal.
//
// Sources:
//   - internal/services/springcloud/spring_cloud_api_portal_resource.go:64-145
//     (Timeouts Create :168 30m / Update :235 30m / Read :311 5m / Delete :359 30m;
//     name :67-74 ForceNew "default"; instance_count :102-107 default 1;
//     https_only_enabled :97-100; public_network_access_enabled :109-112)
//   - go-azure-sdk .../appplatform model_apiportalproperties.go (HTTPSOnly json "httpsOnly",
//     Public "public")
type SpringCloudAPIPortal struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudAPIPortal)(nil)

// NewSpringCloudAPIPortal returns knowledge for the Spring Cloud API portal resource.
func NewSpringCloudAPIPortal() *SpringCloudAPIPortal {
	return &SpringCloudAPIPortal{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/apiPortals",
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
		},
	}
}

func init() { azwise.Register(NewSpringCloudAPIPortal()) }
