package elastic

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ElasticCloudElasticsearch provides resource knowledge for Microsoft.Elastic/monitors.
//
// Mirrors azurerm_elastic_cloud_elasticsearch.
//
// Sources:
//   - terraform-provider-azurerm internal/services/elastic/elastic_cloud_elasticsearch_resource.go
//     schema (lines 46-158: name/sku_name/elastic_cloud_email_address/monitoring_enabled ForceNew;
//     elastic_cloud_* + *_service_url + kibana_sso_uri Computed; logs block), Create (162-221),
//     Read (223-299), timeouts 60m/5m/60m/60m.
//   - internal/services/elastic/validate/elasticsearch_name.go (name regex, 2-32 chars).
//   - go-azure-sdk resource-manager/elastic/2023-06-01/monitorsresource
//     ElasticMonitorResource (Sku.Name required json:"name"), MonitorProperties (monitoringStatus/
//     userInfo settable; elasticProperties/version/provisioningState/liftr* read-only),
//     UserInfo (emailAddress), ResourceSku (name).
//
// Note: the `logs` block is written to a separate ARM resource
// (Microsoft.Elastic/monitors/tagRules "default" via the rules SDK), not the monitor body — its
// rules belong to that sub-service, not here.
type ElasticCloudElasticsearch struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ElasticCloudElasticsearch)(nil)

// NewElasticCloudElasticsearch returns knowledge for the Elastic monitors resource.
func NewElasticCloudElasticsearch() *ElasticCloudElasticsearch {
	return &ElasticCloudElasticsearch{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Elastic/monitors",
			ApiVersions:  []string{"2023-06-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "sku.name"},
				{PropertyPath: "properties.userInfo.emailAddress"},
				{PropertyPath: "properties.monitoringStatus"},
			},
			RequiredFields: []string{
				"sku.name",
				"properties.userInfo.emailAddress",
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). AzureRM ElasticsearchName:
					// 2-32 chars, alphanumeric plus underscore and hyphen.
					Regex:     `^[a-zA-Z0-9_-]{2,32}$`,
					MinLength: 2,
					MaxLength: 32,
					Message:   "must be 2-32 characters and contain only alphanumeric characters, underscores and hyphens",
				},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.elasticProperties",
				"properties.version",
				"properties.provisioningState",
				"properties.liftrResourceCategory",
				"properties.liftrResourcePreference",
			},
		},
	}
}

func init() { azwise.Register(NewElasticCloudElasticsearch()) }
