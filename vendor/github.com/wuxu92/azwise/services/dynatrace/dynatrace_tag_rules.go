package dynatrace

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DynatraceTagRules provides resource knowledge for Microsoft.Dynatrace/monitors/tagRules.
//
// Mirrors azurerm_dynatrace_tag_rules.
//
// Sources:
//   - terraform-provider-azurerm internal/services/dynatrace/dynatrace_tag_rules_resource.go
//     schema Arguments (lines 49-164: name/monitor_id ForceNew; log_rule/metric_rule optional
//     single blocks; filtering_tag.action StringInSlice Include/Exclude), Create (178-225),
//     timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/dynatrace/2023-04-27/tagrules
//     MonitoringTagRulesProperties (logRules/metricRules settable; provisioningState read-only),
//     LogRules/MetricRules (filteringTags array), constants.go PossibleValuesForTagAction
//     (Include, Exclude).
//
// Note: filtering_tag.action lives inside the filteringTags array
// (properties.logRules.filteringTags[*].action / properties.metricRules.filteringTags[*].action).
// azwise/azwise_validate cannot resolve array-element paths, so its enum (Include/Exclude) is
// documented here but not emitted as a StringRule.
type DynatraceTagRules struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DynatraceTagRules)(nil)

// NewDynatraceTagRules returns knowledge for the Dynatrace tagRules resource.
func NewDynatraceTagRules() *DynatraceTagRules {
	return &DynatraceTagRules{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Dynatrace/monitors/tagRules",
			ApiVersions:  []string{"2023-04-27"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			// Server-populated, read-only ARM property.
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewDynatraceTagRules()) }
