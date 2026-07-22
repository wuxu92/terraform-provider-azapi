package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudConfigurationService provides resource knowledge for
// Microsoft.AppPlatform/Spring/configurationServices.
//
// Mirrors azurerm_spring_cloud_configuration_service.
//
// Sources:
//   - internal/services/springcloud/spring_cloud_configuration_service_resource.go:64-180
//     (Timeouts Create :208 30m / Update :257 30m / Read :309 5m / Delete :355 30m;
//     name :65-72 ForceNew "default"; generation :81-88; repository password/private_key :146-158)
//   - go-azure-sdk .../appplatform model_configurationserviceproperties.go (Generation json
//     "generation"), constants.go ConfigurationServiceGeneration (Gen1/Gen2)
type SpringCloudConfigurationService struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudConfigurationService)(nil)

// NewSpringCloudConfigurationService returns knowledge for the configuration service resource.
func NewSpringCloudConfigurationService() *SpringCloudConfigurationService {
	return &SpringCloudConfigurationService{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/configurationServices",
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
				{
					PropertyPath:  "properties.generation",
					AllowedValues: []string{"Gen1", "Gen2"},
				},
			},
			// repository[*].password / private_key / host_key are sensitive but live under an
			// array of git repositories (array-element / map paths), so they are not emitted
			// as declarative SensitiveFields.
		},
	}
}

func init() { azwise.Register(NewSpringCloudConfigurationService()) }
