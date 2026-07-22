package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudApplicationLiveView provides resource knowledge for
// Microsoft.AppPlatform/Spring/applicationLiveViews.
//
// Mirrors azurerm_spring_cloud_application_live_view. The resource carries no
// configurable ARM body beyond its name (which must be "default").
//
// Sources:
//   - internal/services/springcloud/spring_cloud_application_live_view_resource.go:52-150
//     (Timeouts Create :75 30m / Read :118 5m / Delete :146 30m; name :53-58 ForceNew "default")
//   - go-azure-sdk .../appplatform id_applicationliveview.go
type SpringCloudApplicationLiveView struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudApplicationLiveView)(nil)

// NewSpringCloudApplicationLiveView returns knowledge for the application live view resource.
func NewSpringCloudApplicationLiveView() *SpringCloudApplicationLiveView {
	return &SpringCloudApplicationLiveView{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/applicationLiveViews",
			ApiVersions:  []string{"2024-01-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
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

func init() { azwise.Register(NewSpringCloudApplicationLiveView()) }
