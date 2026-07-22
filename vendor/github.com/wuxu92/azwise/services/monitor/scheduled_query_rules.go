// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package monitor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ScheduledQueryRules provides resource knowledge for Microsoft.Insights/scheduledQueryRules.
//
// This ARM type is surfaced by three AzureRM resources built on TWO different API versions with
// substantially different body shapes:
//   - azurerm_monitor_scheduled_query_rules_alert_v2 -> insights/2023-03-15-preview
//     (ScheduledQueryRuleProperties: criteria, scopes, severity AlertSeverity 0-4, evaluationFrequency,
//      windowSize, autoMitigate, muteActionsDuration, ...)
//   - azurerm_monitor_scheduled_query_rules_alert    -> insights/2018-04-16
//     (LogSearchRule: source.dataSourceId (ForceNew), action AlertingAction, description len 1-4096)
//   - azurerm_monitor_scheduled_query_rules_log       -> insights/2018-04-16
//     (LogSearchRule: source.dataSourceId (ForceNew), action LogToMetricAction, criteria dimensions)
//
// Because the two API versions disagree on required fields (v2 requires criteria+scopes; v1 requires
// source+action), on ForceNew members (v1's data_source_id is not present in v2), on the severity
// representation (v2 numeric AlertSeverity vs v1 enum), and on defaults, ONLY the knowledge universal
// to all three resources is captured here: name/location ForceNew, timeouts, and a non-empty name.
// Version-specific rules are intentionally omitted to avoid corrupting validation for the other bodies.
//
// Sources:
//   - terraform-provider-azurerm internal/services/monitor/monitor_scheduled_query_rules_alert_v2_resource.go:87-405
//   - terraform-provider-azurerm internal/services/monitor/monitor_scheduled_query_rules_alert_resource.go:32-229
//   - terraform-provider-azurerm internal/services/monitor/monitor_scheduled_query_rules_log_resource.go:26-134
//   - terraform-provider-azurerm vendor/.../insights/2023-03-15-preview/scheduledqueryrules/model_scheduledqueryruleproperties.go:6-26
//   - terraform-provider-azurerm vendor/.../insights/2018-04-16/scheduledqueryrules/model_logsearchrule.go:14-26
//
// Intentionally skipped here:
//   - criteria / action / source / schedule / trigger: version-specific blocks living on nested and
//     array-element paths (properties.criteria[*], properties.source.*, properties.action.*).
//   - data_source_id ForceNew (v1 alert & log only): properties.source.dataSourceId — not universal,
//     absent from the v2 body.
//   - severity, enabled, description defaults/enums: differ by API version; not universal.
type ScheduledQueryRules struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ScheduledQueryRules)(nil)

func NewScheduledQueryRules() *ScheduledQueryRules {
	return &ScheduledQueryRules{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Insights/scheduledQueryRules",
			ApiVersions:  []string{"2023-03-15-preview", "2018-04-16"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// v2: StringIsNotEmpty; v1 alert & log: StringDoesNotContainAny("<>*%&:\?+/").
					// The universal, non-corrupting constraint across all three is a non-empty name.
					MinLength: 1,
					Message:   "Monitor scheduled query rule name must not be empty",
				},
			},
		},
	}
}

func init() { azwise.Register(NewScheduledQueryRules()) }
