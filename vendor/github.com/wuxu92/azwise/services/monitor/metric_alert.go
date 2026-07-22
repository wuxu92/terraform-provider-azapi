// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package monitor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MetricAlert provides resource knowledge for Microsoft.Insights/metricAlerts.
//
// Sources:
//   - terraform-provider-azurerm internal/services/monitor/monitor_metric_alert_resource.go:35-500
//     (azurerm_monitor_metric_alert schema: ForceNew, timeouts, validators, defaults; create mapping
//     to metricalerts.MetricAlertResource / MetricAlertProperties)
//   - terraform-provider-azurerm vendor/.../insights/2018-03-01/metricalerts/model_metricalertproperties.go:14-28
//
// Intentionally skipped here:
//   - criteria / dynamic_criteria / application_insights_web_test_location_availability_criteria:
//     a discriminated MetricAlertCriteria union (properties.criteria). ExactlyOneOf across the three
//     blocks and their nested enum validators (aggregation, operator, alert_sensitivity, dimension
//     operator) are nested/array-element paths, not expressible declaratively.
//   - action: array block (properties.actions[*]); per-element validators are array-element paths.
//   - location: hardcoded to "Global" by the resource (envelope field), not a properties.* path.
type MetricAlert struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MetricAlert)(nil)

func NewMetricAlert() *MetricAlert {
	return &MetricAlert{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Insights/metricAlerts",
			ApiVersions:  []string{"2018-03-01"},
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
					Message:   "Monitor metric alert name must not be empty",
				},
				{
					PropertyPath:  "properties.windowSize",
					AllowedValues: []string{"PT1M", "PT5M", "PT15M", "PT30M", "PT1H", "PT6H", "PT12H", "P1D"},
					Message:       "window_size must be one of PT1M, PT5M, PT15M, PT30M, PT1H, PT6H, PT12H or P1D",
				},
			},
			IntRules: []azwise.IntRule{
				// severity IntBetween(0, 4).
				{PropertyPath: "properties.severity", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(4))},
			},
			// Criteria and Scopes are non-pointer (required) in MetricAlertProperties. Severity,
			// EvaluationFrequency, WindowSize and Enabled are non-pointer but carry schema Defaults,
			// so they are not user-required.
			RequiredFields: []string{
				"properties.scopes",
				"properties.criteria",
			},
			DefaultValues: []azwise.DefaultValue{
				// auto_mitigate Default:true -> properties.autoMitigate true.
				{PropertyPath: "properties.autoMitigate", Value: true},
				// enabled Default:true -> properties.enabled true.
				{PropertyPath: "properties.enabled", Value: true},
				// frequency Default:"PT1M" -> properties.evaluationFrequency.
				{PropertyPath: "properties.evaluationFrequency", Value: "PT1M"},
				// severity Default:3 -> properties.severity.
				{PropertyPath: "properties.severity", Value: float64(3)},
				// window_size Default:"PT5M" -> properties.windowSize.
				{PropertyPath: "properties.windowSize", Value: "PT5M"},
			},
		},
	}
}

func init() { azwise.Register(NewMetricAlert()) }
