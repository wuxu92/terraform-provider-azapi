// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package monitor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutoscaleSetting provides resource knowledge for Microsoft.Insights/autoscaleSettings.
//
// Sources:
//   - terraform-provider-azurerm internal/services/monitor/monitor_autoscale_setting_resource.go:30-496
//     (azurerm_monitor_autoscale_setting schema: ForceNew, timeouts, defaults; create mapping to
//     autoscalesettings.AutoscaleSettingResource / AutoscaleSetting)
//   - terraform-provider-azurerm vendor/.../insights/2022-10-01/autoscalesettings/model_autoscalesetting.go:6-14
//   - terraform-provider-azurerm vendor/.../insights/2022-10-01/autoscalesettings/model_predictiveautoscalepolicy.go:6-9
//   - terraform-provider-azurerm vendor/.../insights/2022-10-01/autoscalesettings/constants.go:155-159
//     (PredictiveAutoscalePolicyScaleMode)
//
// Intentionally skipped here:
//   - profile: array block (properties.profiles[*]) with nested capacity/rule/fixed_date/recurrence
//     sub-blocks; per-element validators are array-element paths, not expressible declaratively.
//   - notification: array/nested block (properties.notifications[*]); per-element validators are
//     array-element paths.
//   - predictive.look_ahead_time ISO8601DurationBetween("PT1M","PT1H"): a nested single-block path
//     (properties.predictiveAutoscalePolicy.scaleLookAheadTime), not a top-level body field.
type AutoscaleSetting struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutoscaleSetting)(nil)

func NewAutoscaleSetting() *AutoscaleSetting {
	return &AutoscaleSetting{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Insights/autoscaleSettings",
			ApiVersions:  []string{"2022-10-01"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "properties.targetResourceUri"},
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
					Message:   "Monitor autoscale setting name must not be empty",
				},
				{
					// predictive.scale_mode StringInSlice([Enabled, ForecastOnly]) — the schema omits
					// Disabled (block omission means disabled). azwise uses the full ARM enum set.
					PropertyPath:  "properties.predictiveAutoscalePolicy.scaleMode",
					AllowedValues: []string{"Disabled", "Enabled", "ForecastOnly"},
					Message:       "predictive.scale_mode must be one of Disabled, Enabled or ForecastOnly",
				},
			},
			// Profiles and TargetResourceUri are required (Profiles is non-pointer in the model,
			// target_resource_id is Required in schema). Enabled has a Default, so not user-required.
			RequiredFields: []string{
				"properties.profiles",
				"properties.targetResourceUri",
			},
			DefaultValues: []azwise.DefaultValue{
				// enabled Default:true -> properties.enabled true.
				{PropertyPath: "properties.enabled", Value: true},
			},
		},
	}
}

func init() { azwise.Register(NewAutoscaleSetting()) }
