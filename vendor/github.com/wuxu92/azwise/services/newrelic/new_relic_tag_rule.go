package newrelic

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NewRelicTagRule provides resource knowledge for NewRelic.Observability/monitors/tagRules.
//
// Mirrors azurerm_new_relic_tag_rule.
//
// ARM type casing verified against go-azure-sdk tagrules/id_tagrule.go Segments:
// "NewRelic.Observability" / "monitors" / "tagRules" (the tag rule name is hardcoded "default").
//
// Sources:
//   - terraform-provider-azurerm internal/services/newrelic/new_relic_tag_rule_resource.go
//     Arguments (52-141: monitor_id parent ref ForceNew; azure_active_directory_log_enabled/
//     activity_log_enabled/subscription_log_enabled/metric_enabled bool Default false;
//     log_tag_filter/metric_tag_filter blocks with name/action(Include/Exclude)/value), Create
//     (147-224: bools mapped to Enabled/Disabled status enums; UserEmail resolved from monitor),
//     timeouts Create 30m / Read 5m / Update 30m / Delete 30m.
//   - go-azure-sdk resource-manager/newrelic/2024-03-01/tagrules
//     MonitoringTagRulesProperties (logRules/metricRules settable; provisioningState read-only),
//     LogRules (filteringTags/sendAadLogs/sendActivityLogs/sendSubscriptionLogs), MetricRules
//     (filteringTags/sendMetrics/userEmail), constants.go Send*Status [Disabled, Enabled] /
//     TagAction [Exclude, Include].
//
// Note: filtering-tag action (Include/Exclude) lives at the array-element paths
// properties.logRules.filteringTags[*].action and properties.metricRules.filteringTags[*].action;
// azwise cannot resolve array-element paths so its enum is documented but not emitted as a rule.
type NewRelicTagRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NewRelicTagRule)(nil)

// NewNewRelicTagRule returns knowledge for the NewRelic tagRules resource.
func NewNewRelicTagRule() *NewRelicTagRule {
	return &NewRelicTagRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "NewRelic.Observability/monitors/tagRules",
			ApiVersions:  []string{"2024-03-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.logRules.sendAadLogs",
					AllowedValues: []string{"Enabled", "Disabled"},
				},
				{
					PropertyPath:  "properties.logRules.sendActivityLogs",
					AllowedValues: []string{"Enabled", "Disabled"},
				},
				{
					PropertyPath:  "properties.logRules.sendSubscriptionLogs",
					AllowedValues: []string{"Enabled", "Disabled"},
				},
				{
					PropertyPath:  "properties.metricRules.sendMetrics",
					AllowedValues: []string{"Enabled", "Disabled"},
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.logRules.sendAadLogs", Value: "Disabled"},
				{PropertyPath: "properties.logRules.sendActivityLogs", Value: "Disabled"},
				{PropertyPath: "properties.logRules.sendSubscriptionLogs", Value: "Disabled"},
				{PropertyPath: "properties.metricRules.sendMetrics", Value: "Disabled"},
			},
			// Server-populated, read-only ARM property.
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewNewRelicTagRule()) }
