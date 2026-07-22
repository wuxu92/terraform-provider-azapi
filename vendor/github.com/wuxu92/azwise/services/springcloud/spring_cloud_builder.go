package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudBuilder provides resource knowledge for
// Microsoft.AppPlatform/Spring/buildServices/builders.
//
// Mirrors azurerm_spring_cloud_builder.
//
// Sources:
//   - internal/services/springcloud/spring_cloud_builder_resource.go:25-110
//     (Timeouts :34-39 30m/5m/30m/30m; name :52-56 ForceNew; build_pack_group :65-88 Required;
//     stack :90-106 Required)
//   - go-azure-sdk .../appplatform model_builderproperties.go (BuildpackGroups json
//     "buildpackGroups", Stack "stack")
type SpringCloudBuilder struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudBuilder)(nil)

// NewSpringCloudBuilder returns knowledge for the Spring Cloud builder resource.
func NewSpringCloudBuilder() *SpringCloudBuilder {
	return &SpringCloudBuilder{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/buildServices/builders",
			ApiVersions:  []string{"2024-01-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.stack",
				"properties.buildpackGroups",
			},
		},
	}
}

func init() { azwise.Register(NewSpringCloudBuilder()) }
