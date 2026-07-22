package datadog

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DatadogMonitorTagRule provides resource knowledge for
// Microsoft.Datadog/monitors/tagRules.
//
// Mirrors azurerm_datadog_monitor_tag_rule.
//
// Sources:
//   - terraform-provider-azurerm internal/services/datadog/datadog_monitor_tag_rule_resource.go
//     schema (lines 41-129): datadog_monitor_id parent reference (ForceNew, envelope);
//     name optional default "default" (resource name); log block (aad_log_enabled/
//     subscription_log_enabled/resource_log_enabled bools + filter list) and metric block
//     (filter list); each filter.action is StringInSlice(TagAction); Create (133-169);
//     timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/datadog/2021-03-01/rules:
//     MonitoringTagRulesProperties (logRules/metricRules), LogRules
//     (sendAadLogs/sendSubscriptionLogs/sendResourceLogs + filteringTags), MetricRules
//     (filteringTags), FilteringTag (name/value/action), constants TagAction{Exclude,Include};
//     id_tagrule.go type Microsoft.Datadog/monitors/tagRules.
//
// No body-level ForceNew: the only ForceNew field is datadog_monitor_id, which is the
// parent monitor reference (envelope), not part of this resource's ARM body. All body
// properties (logRules/metricRules) are optional, so there are no RequiredFields.
//
// TODO: the filter `action` enum (TagAction {Exclude, Include}) lives at
// properties.logRules.filteringTags[*].action and
// properties.metricRules.filteringTags[*].action — array-element paths that azwise
// StringRules cannot express (no [*] support). Left unmapped; the filter name/value/action
// fields are also Required only within array elements, so they are not emitted as
// RequiredFields either.
type DatadogMonitorTagRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DatadogMonitorTagRule)(nil)

// NewDatadogMonitorTagRule returns knowledge for the monitors/tagRules resource.
func NewDatadogMonitorTagRule() *DatadogMonitorTagRule {
	return &DatadogMonitorTagRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Datadog/monitors/tagRules",
			ApiVersions:  []string{"2021-03-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewDatadogMonitorTagRule()) }
