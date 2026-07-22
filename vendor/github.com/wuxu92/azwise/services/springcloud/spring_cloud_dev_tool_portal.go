package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudDevToolPortal provides resource knowledge for
// Microsoft.AppPlatform/Spring/DevToolPortals.
//
// Mirrors azurerm_spring_cloud_dev_tool_portal.
//
// Sources:
//   - internal/services/springcloud/spring_cloud_dev_tool_portal_resource.go:64-130
//     (Timeouts Create :140 30m / Update :189 30m / Read :230 5m / Delete :280 30m;
//     name :65-70 ForceNew "default")
//   - go-azure-sdk .../appplatform id_devtoolportal.go, model_devtoolportalproperties.go
type SpringCloudDevToolPortal struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudDevToolPortal)(nil)

// NewSpringCloudDevToolPortal returns knowledge for the dev tool portal resource.
func NewSpringCloudDevToolPortal() *SpringCloudDevToolPortal {
	return &SpringCloudDevToolPortal{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/DevToolPortals",
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

func init() { azwise.Register(NewSpringCloudDevToolPortal()) }
