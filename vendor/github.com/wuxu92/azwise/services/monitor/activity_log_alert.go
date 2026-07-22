// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package monitor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ActivityLogAlert provides resource knowledge for Microsoft.Insights/activityLogAlerts.
//
// Sources:
//   - terraform-provider-azurerm internal/services/monitor/monitor_activity_log_alert_resource.go:31-487
//     (azurerm_monitor_activity_log_alert schema: ForceNew, timeouts, defaults; create mapping to
//     activitylogalertsapis.ActivityLogAlertResource / AlertRuleProperties)
//   - terraform-provider-azurerm vendor/.../insights/2020-10-01/activitylogalertsapis/model_alertruleproperties.go:6-12
//
// Intentionally skipped here:
//   - criteria: an AlertRuleAllOfCondition block (properties.condition). Its many enum/conflict
//     validators (category, level(s), status(es), recommendation_*, resource_health, service_health)
//     live on nested and array-element paths, not expressible as declarative body rules.
//   - action: array block (properties.actions.actionGroups[*]); per-element validators are
//     array-element paths.
//   - location: CustomizeDiff restricts to [global, westeurope, northeurope, eastus2euap]; envelope
//     field with a runtime check, not a static declarative rule.
type ActivityLogAlert struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ActivityLogAlert)(nil)

func NewActivityLogAlert() *ActivityLogAlert {
	return &ActivityLogAlert{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Insights/activityLogAlerts",
			ApiVersions:  []string{"2020-10-01"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					MinLength: 1,
					Message:   "Monitor activity log alert name must not be empty",
				},
			},
			// Scopes, Condition and Actions are non-pointer (required) in AlertRuleProperties.
			RequiredFields: []string{
				"properties.scopes",
				"properties.condition",
				"properties.actions",
			},
			DefaultValues: []azwise.DefaultValue{
				// enabled Default:true -> properties.enabled true.
				{PropertyPath: "properties.enabled", Value: true},
			},
		},
	}
}

func init() { azwise.Register(NewActivityLogAlert()) }
