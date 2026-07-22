// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package monitor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PrometheusRuleGroup provides resource knowledge for Microsoft.AlertsManagement/prometheusRuleGroups.
//
// Sources:
//   - terraform-provider-azurerm internal/services/monitor/monitor_alert_prometheus_rule_group_resource.go:63-309
//     (azurerm_monitor_alert_prometheus_rule_group schema: ForceNew, timeouts, validators; create
//     mapping to prometheusrulegroups.PrometheusRuleGroupResource / PrometheusRuleGroupProperties)
//   - terraform-provider-azurerm vendor/.../alertsmanagement/2023-03-01/prometheusrulegroupresources/model_prometheusrulegroupproperties.go:6-13
//
// Intentionally skipped here:
//   - rule: array block (properties.rules[*]) with alert/record, expression, severity, action,
//     annotations, labels, alert_resolution sub-fields. The record-vs-alert exclusivity is enforced
//     by CustomizeDiff (runtime) and the per-element validators are array-element paths.
type PrometheusRuleGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PrometheusRuleGroup)(nil)

func NewPrometheusRuleGroup() *PrometheusRuleGroup {
	return &PrometheusRuleGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AlertsManagement/prometheusRuleGroups",
			ApiVersions:  []string{"2023-03-01"},
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
					// name StringLenBetween(1, 260).
					MinLength: 1,
					MaxLength: 260,
					Message:   "Prometheus rule group name must be between 1 and 260 characters",
				},
			},
			// Rules and Scopes are non-pointer (required) in PrometheusRuleGroupProperties.
			RequiredFields: []string{
				"properties.scopes",
				"properties.rules",
			},
		},
	}
}

func init() { azwise.Register(NewPrometheusRuleGroup()) }
