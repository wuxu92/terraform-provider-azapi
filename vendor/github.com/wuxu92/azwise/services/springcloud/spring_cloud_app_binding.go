package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudAppBinding provides resource knowledge for
// Microsoft.AppPlatform/Spring/apps/bindings.
//
// Merges three AzureRM "association" resources that all target the same ARM bindings
// type, discriminated by the bound resource:
//   - azurerm_spring_cloud_app_cosmosdb_association (properties.resourceId -> Cosmos DB account)
//   - azurerm_spring_cloud_app_mysql_association    (properties.resourceId -> MySQL server)
//   - azurerm_spring_cloud_app_redis_association    (properties.resourceId -> Redis cache)
//
// Only universal knowledge is unioned (name validation, timeouts, ForceNew on the
// bound resourceId). The bound access key (properties.key) is sensitive; per-kind
// binding parameters (api_type, database_name, ssl flags) go into
// properties.bindingParameters (map) and are kind-specific.
//
// Sources:
//   - internal/services/springcloud/spring_cloud_app_cosmosdb_association_resource.go:59-145
//     (Timeouts :59-64 30m/5m/30m/30m; name :67-72; cosmosdb_account_id :81-86 ForceNew)
//   - internal/services/springcloud/spring_cloud_app_mysql_association_resource.go:51-100
//     (mysql_server_id :73-82 ForceNew)
//   - internal/services/springcloud/spring_cloud_app_redis_association_resource.go:48-90
//     (redis_cache_id :70-75 ForceNew)
//   - name validate: internal/services/springcloud/validate/spring_cloud_app_association_name.go:23
//   - go-azure-sdk .../appplatform model_bindingresourceproperties.go (ResourceId json
//     "resourceId", Key "key", BindingParameters "bindingParameters")
type SpringCloudAppBinding struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudAppBinding)(nil)

// NewSpringCloudAppBinding returns knowledge for the Spring Cloud app binding resource.
func NewSpringCloudAppBinding() *SpringCloudAppBinding {
	return &SpringCloudAppBinding{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/apps/bindings",
			ApiVersions:  []string{"2024-01-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// The bound resource id is ForceNew across all three association kinds.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.resourceId"},
			},
			RequiredFields: []string{"properties.resourceId"},
			StringRules: []azwise.StringRule{
				// name: begins with a letter, ends alphanumeric, 4-32 lowercase/digits/hyphens.
				{
					PropertyPath: "",
					Regex:        `^([a-z])([a-z\d-]{2,30})([a-z\d])$`,
					Message:      "spring cloud app association name must begin with a letter, end with a letter or number, contain only lowercase letters, numbers and hyphens, and be 4-32 characters long",
				},
			},
			// properties.key holds the bound service access key and is sensitive.
			SensitiveFields: []string{"properties.key"},
		},
	}
}

func init() { azwise.Register(NewSpringCloudAppBinding()) }
