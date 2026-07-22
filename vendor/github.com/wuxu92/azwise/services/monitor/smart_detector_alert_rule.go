// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package monitor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SmartDetectorAlertRule provides resource knowledge for
// Microsoft.AlertsManagement/smartDetectorAlertRules.
//
// Sources:
//   - terraform-provider-azurerm internal/services/monitor/monitor_smart_detector_alert_rule_resource.go:31-217
//     (azurerm_monitor_smart_detector_alert_rule schema: ForceNew, timeouts, validators, defaults;
//     create mapping to smartdetectoralertrules.AlertRule / AlertRuleProperties)
//   - terraform-provider-azurerm vendor/.../alertsmanagement/2019-06-01/smartdetectoralertrules/model_alertruleproperties.go:6-15
//   - terraform-provider-azurerm vendor/.../alertsmanagement/2019-06-01/smartdetectoralertrules/constants.go:14-17,55-61
//     (AlertRuleState, Severity)
//
// Intentionally skipped here:
//   - action_group: single block (properties.actionGroups) whose ids/email_subject/webhook_payload
//     live on nested and array-element paths, not expressible as top-level body rules.
//   - frequency ISO8601Duration and throttling_duration ISO8601Duration: duration-format checks with
//     no StringInSlice/regex extracted; ARM validates the duration server-side.
//   - location: hardcoded to "Global" by the resource (envelope field).
type SmartDetectorAlertRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SmartDetectorAlertRule)(nil)

func NewSmartDetectorAlertRule() *SmartDetectorAlertRule {
	return &SmartDetectorAlertRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AlertsManagement/smartDetectorAlertRules",
			ApiVersions:  []string{"2019-06-01"},
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
					Message:   "Monitor smart detector alert rule name must not be empty",
				},
				{
					// detector_type StringInSlice(...) -> properties.detector.id.
					PropertyPath: "properties.detector.id",
					AllowedValues: []string{
						"FailureAnomaliesDetector",
						"RequestPerformanceDegradationDetector",
						"DependencyPerformanceDegradationDetector",
						"ExceptionVolumeChangedDetector",
						"TraceSeverityDetector",
						"MemoryLeakDetector",
					},
					Message: "detector_type must be one of the supported smart detector types",
				},
				{
					// severity StringInSlice([Sev0..Sev4]) -> properties.severity.
					PropertyPath:  "properties.severity",
					AllowedValues: []string{"Sev0", "Sev1", "Sev2", "Sev3", "Sev4"},
					Message:       "severity must be one of Sev0, Sev1, Sev2, Sev3 or Sev4",
				},
				{
					// enabled maps to properties.state (Enabled/Disabled); full ARM enum set.
					PropertyPath:  "properties.state",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "state must be either Enabled or Disabled",
				},
			},
			// Detector, Frequency, Scope, Severity, State and ActionGroups are non-pointer (required)
			// in AlertRuleProperties; state carries a schema Default so it is not user-required.
			RequiredFields: []string{
				"properties.detector",
				"properties.frequency",
				"properties.scope",
				"properties.severity",
				"properties.actionGroups",
			},
			DefaultValues: []azwise.DefaultValue{
				// enabled Default:true -> properties.state Enabled.
				{PropertyPath: "properties.state", Value: "Enabled"},
			},
		},
	}
}

func init() { azwise.Register(NewSmartDetectorAlertRule()) }
