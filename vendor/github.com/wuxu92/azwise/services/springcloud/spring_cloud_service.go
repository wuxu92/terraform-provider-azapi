package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudService provides resource knowledge for Microsoft.AppPlatform/Spring.
//
// Mirrors azurerm_spring_cloud_service (the top-level Azure Spring Apps service instance).
//
// Sources:
//   - terraform-provider-azurerm internal/services/springcloud/spring_cloud_service_resource.go:33-260
//     (Timeouts :47-52 Create 60m / Read 5m / Update 30m / Delete 30m; name :60-65;
//     sku_name :73-83; sku_tier :85-96; network block :197-260)
//   - name validate: internal/services/springcloud/validate/spring_cloud_service_name.go:23
//   - go-azure-sdk resource-manager/appplatform/2024-01-01-preview/appplatform
//     model_serviceresource.go (top-level Sku), model_sku.go, model_networkprofile.go,
//     constants.go (no dedicated Sku enum — name/tier are free strings restricted by AzureRM)
type SpringCloudService struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudService)(nil)

// NewSpringCloudService returns knowledge for the Spring Cloud service resource.
func NewSpringCloudService() *SpringCloudService {
	return &SpringCloudService{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring",
			ApiVersions:  []string{"2024-01-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// sku (top-level) and the entire networkProfile are replace-only.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "sku.name"},
				{PropertyPath: "sku.tier"},
				{PropertyPath: "properties.networkProfile.appSubnetId"},
				{PropertyPath: "properties.networkProfile.serviceRuntimeSubnetId"},
				{PropertyPath: "properties.networkProfile.serviceCidr"},
				{PropertyPath: "properties.networkProfile.outboundType"},
				{PropertyPath: "properties.networkProfile.appNetworkResourceGroup"},
				{PropertyPath: "properties.networkProfile.serviceRuntimeNetworkResourceGroup"},
			},
			StringRules: []azwise.StringRule{
				// name: begins with a letter, ends alphanumeric, 4-32 lowercase/digits/hyphens.
				{
					PropertyPath: "",
					Regex:        `^([a-z])([a-z\d-]{2,30})([a-z\d])$`,
					Message:      "spring cloud service name must begin with a letter, end with a letter or number, contain only lowercase letters, numbers and hyphens, and be 4-32 characters long",
				},
				// sku_name: AzureRM restricts to B0/S0/E0.
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"B0", "S0", "E0"},
				},
				// sku_tier: AzureRM restricts to Basic/Enterprise/Standard/StandardGen2.
				{
					PropertyPath:  "sku.tier",
					AllowedValues: []string{"Basic", "Enterprise", "Standard", "StandardGen2"},
				},
				// network.outbound_type.
				{
					PropertyPath:  "properties.networkProfile.outboundType",
					AllowedValues: []string{"loadBalancer", "userDefinedRouting"},
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "sku.name", Value: "S0"},
				{PropertyPath: "properties.networkProfile.outboundType", Value: "loadBalancer"},
			},
			// NOTE: config_server_git_setting (properties.configServerProperties) carries
			// sensitive git credentials (password, private_key, host_key). These live under
			// deeply-nested blocks; they are documented here but not emitted as declarative
			// sensitive paths. container_registry and default_build_service are managed via the
			// child Spring/containerRegistries and Spring/buildServices resources, and
			// build_agent_pool_size via Spring/buildServices/agentPools — not part of this body.
		},
	}
}

func init() { azwise.Register(NewSpringCloudService()) }
