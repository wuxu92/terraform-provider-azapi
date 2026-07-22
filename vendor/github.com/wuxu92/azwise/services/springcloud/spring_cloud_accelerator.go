package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudAccelerator provides resource knowledge for
// Microsoft.AppPlatform/Spring/applicationAccelerators.
//
// Mirrors azurerm_spring_cloud_accelerator. The resource carries no configurable ARM
// body beyond its name (which must be "default") and provisioning.
//
// Sources:
//   - internal/services/springcloud/spring_cloud_accelerator_resource.go:60-160
//     (Timeouts Create :86 30m / Read :129 5m / Delete :157 30m; name :64-69 ForceNew "default")
//   - go-azure-sdk .../appplatform id_applicationaccelerator.go
type SpringCloudAccelerator struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudAccelerator)(nil)

// NewSpringCloudAccelerator returns knowledge for the Spring Cloud accelerator resource.
func NewSpringCloudAccelerator() *SpringCloudAccelerator {
	return &SpringCloudAccelerator{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/applicationAccelerators",
			ApiVersions:  []string{"2024-01-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// name must be "default" (Azure allows a single accelerator per service).
				{
					PropertyPath:  "",
					AllowedValues: []string{"default"},
				},
			},
		},
	}
}

func init() { azwise.Register(NewSpringCloudAccelerator()) }
