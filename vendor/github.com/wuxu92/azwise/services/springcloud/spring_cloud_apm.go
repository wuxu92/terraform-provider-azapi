package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudApm provides resource knowledge for Microsoft.AppPlatform/Spring/apms.
//
// Merges five AzureRM resources that all target the same ARM apms type, discriminated
// by properties.type:
//   - azurerm_spring_cloud_app_dynamics_application_performance_monitoring         (type = AppDynamics)
//   - azurerm_spring_cloud_application_insights_application_performance_monitoring  (type = ApplicationInsights)
//   - azurerm_spring_cloud_dynatrace_application_performance_monitoring             (type = Dynatrace)
//   - azurerm_spring_cloud_elastic_application_performance_monitoring               (type = ElasticAPM)
//   - azurerm_spring_cloud_new_relic_application_performance_monitoring             (type = NewRelic)
//
// Only universal knowledge is unioned (timeouts, the type discriminator enum/required).
// Each APM's provider-specific credentials go into properties.properties (config) and
// properties.secrets (sensitive) as string maps — map-keyed, so not expressible as
// declarative ARM-path rules.
//
// Sources:
//   - internal/services/springcloud/spring_cloud_app_dynamics_application_performance_monitoring_resource.go:139-141 (Timeout 30m)
//   - internal/services/springcloud/spring_cloud_application_insights_application_performance_monitoring_resource.go:114-116
//   - internal/services/springcloud/spring_cloud_dynatrace_application_performance_monitoring_resource.go:120-122
//   - internal/services/springcloud/spring_cloud_elastic_application_performance_monitoring_resource.go:101-103
//   - internal/services/springcloud/spring_cloud_new_relic_application_performance_monitoring_resource.go:141-143
//   - go-azure-sdk .../appplatform model_apmproperties.go (Type json "type", Secrets "secrets"),
//     constants.go ApmType (AppDynamics/ApplicationInsights/Dynatrace/ElasticAPM/NewRelic)
type SpringCloudApm struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudApm)(nil)

// NewSpringCloudApm returns knowledge for the Spring Cloud APM resource.
func NewSpringCloudApm() *SpringCloudApm {
	return &SpringCloudApm{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/apms",
			ApiVersions:  []string{"2024-01-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// properties.type is required (the APM discriminator).
			RequiredFields: []string{"properties.type"},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.type",
					AllowedValues: []string{"AppDynamics", "ApplicationInsights", "Dynatrace", "ElasticAPM", "NewRelic"},
				},
			},
			// Per-provider secrets live in properties.secrets (map[string]string) and are
			// sensitive; map-keyed values have no single ARM path, so they are documented
			// here rather than emitted as declarative SensitiveFields.
		},
	}
}

func init() { azwise.Register(NewSpringCloudApm()) }
