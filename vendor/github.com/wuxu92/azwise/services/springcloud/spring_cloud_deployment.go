package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudDeployment provides resource knowledge for
// Microsoft.AppPlatform/Spring/apps/deployments.
//
// Merges three AzureRM resources that all target the same ARM deployments type,
// discriminated by the source `type`:
//   - azurerm_spring_cloud_java_deployment      (source.type = Jar)
//   - azurerm_spring_cloud_container_deployment (source.type = Container)
//   - azurerm_spring_cloud_build_deployment     (source.type = BuildResult)
//
// Only universal knowledge is unioned (name validation, timeouts). Kind-specific
// source fields are not unioned as ForceNew/Required; source enum constraints that
// only apply when a given sub-object is present remain safe (they fire only when
// that path exists).
//
// Sources:
//   - internal/services/springcloud/spring_cloud_java_deployment_resource.go:326-400
//     (Timeouts :45-50 30m/5m/30m/30m; name :329-334; runtime_version :391-397)
//   - internal/services/springcloud/spring_cloud_container_deployment_resource.go:52-160
//   - internal/services/springcloud/spring_cloud_build_deployment_resource.go:52-125
//   - name validate: internal/services/springcloud/validate/spring_cloud_deployment_name.go:23
//   - go-azure-sdk .../appplatform model_deploymentresourceproperties.go (Source UserSourceInfo),
//     model_jaruploadedusersourceinfo.go (RuntimeVersion json "runtimeVersion"),
//     constants.go SupportedRuntimeValue (Java_8/Java_11/Java_17/Java_21/NetCore_31)
type SpringCloudDeployment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudDeployment)(nil)

// NewSpringCloudDeployment returns knowledge for the Spring Cloud deployment resource.
func NewSpringCloudDeployment() *SpringCloudDeployment {
	return &SpringCloudDeployment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/apps/deployments",
			ApiVersions:  []string{"2024-01-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// name (deployment): begins with a letter, ends alphanumeric,
				// 4-32 lowercase/digits/hyphens.
				{
					PropertyPath: "",
					Regex:        `^([a-z])([a-z\d-]{2,30})([a-z\d])$`,
					Message:      "spring cloud deployment name must begin with a letter, end with a letter or number, contain only lowercase letters, numbers and hyphens, and be 4-32 characters long",
				},
				// runtime_version applies only to Jar / NetCore source deployments; the rule
				// fires only when properties.source.runtimeVersion is present, so it is safe
				// across the merged kinds. azwise uses the full ARM SDK set (AzureRM's Java
				// resource restricts to Java_8/Java_11/Java_17).
				{
					PropertyPath:  "properties.source.runtimeVersion",
					AllowedValues: []string{"Java_8", "Java_11", "Java_17", "Java_21", "NetCore_31"},
				},
			},
		},
	}
}

func init() { azwise.Register(NewSpringCloudDeployment()) }
