package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudBuildPackBinding provides resource knowledge for
// Microsoft.AppPlatform/Spring/buildServices/builders/buildpackBindings.
//
// Mirrors azurerm_spring_cloud_build_pack_binding.
//
// Sources:
//   - internal/services/springcloud/spring_cloud_build_pack_binding_resource.go:24-100
//     (Timeouts :33-38 30m/5m/30m/30m; name :51-55 ForceNew; binding_type :64-75;
//     launch.properties/secrets :77-96)
//   - go-azure-sdk .../appplatform model_buildpackbindingproperties.go (BindingType json
//     "bindingType", LaunchProperties "launchProperties"),
//     constants.go BindingType (ApacheSkyWalking/AppDynamics/ApplicationInsights/
//     CACertificates/Dynatrace/ElasticAPM/NewRelic)
type SpringCloudBuildPackBinding struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudBuildPackBinding)(nil)

// NewSpringCloudBuildPackBinding returns knowledge for the buildpack binding resource.
func NewSpringCloudBuildPackBinding() *SpringCloudBuildPackBinding {
	return &SpringCloudBuildPackBinding{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/buildServices/builders/buildpackBindings",
			ApiVersions:  []string{"2024-01-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// binding_type: AzureRM restricts to a subset; azwise uses the full ARM SDK set.
				{
					PropertyPath:  "properties.bindingType",
					AllowedValues: []string{"ApacheSkyWalking", "AppDynamics", "ApplicationInsights", "CACertificates", "Dynatrace", "ElasticAPM", "NewRelic"},
				},
			},
			// launch.secrets (properties.launchProperties.secrets) is a sensitive string map
			// (map-keyed) and is not expressible as a single declarative ARM path.
		},
	}
}

func init() { azwise.Register(NewSpringCloudBuildPackBinding()) }
