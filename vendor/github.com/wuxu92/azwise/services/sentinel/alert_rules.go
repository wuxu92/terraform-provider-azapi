// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package sentinel

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AlertRules provides resource knowledge for
// Microsoft.SecurityInsights/alertRules (an extension resource on a
// Microsoft.OperationalInsights/workspaces scope).
//
// Contributing Terraform resources (all map to the alertRules ARM type,
// discriminated by the top-level `kind`):
//   - azurerm_sentinel_alert_rule_fusion                             (kind Fusion)
//   - azurerm_sentinel_alert_rule_machine_learning_behavior_analytics (kind MLBehaviorAnalytics)
//   - azurerm_sentinel_alert_rule_ms_security_incident              (kind MicrosoftSecurityIncidentCreation)
//   - azurerm_sentinel_alert_rule_nrt                               (kind NRT)
//   - azurerm_sentinel_alert_rule_scheduled                        (kind Scheduled)
//   - azurerm_sentinel_alert_rule_threat_intelligence              (kind ThreatIntelligence)
//
// NOTE: azurerm_sentinel_alert_rule_anomaly_built_in and _anomaly_duplicate are
// NOT alertRules — they create Microsoft.SecurityInsights/securityMLAnalyticsSettings
// (kind Anomaly) and live in security_ml_analytics_settings.go.
//
// Merge policy: only universal knowledge is unioned. kind-specific value
// constraints (severity, triggerOperator, productFilter, triggerThreshold) are
// still emitted because a PropertyPath rule fires only when that path is present
// in the body — safe across the discriminated union. Non-universal ForceNew and
// per-kind DefaultValues are intentionally omitted so they cannot corrupt other
// kinds (e.g. scheduled-only query_frequency/trigger_operator defaults).
//
// Sources:
//   - terraform-provider-azurerm internal/services/sentinel/sentinel_alert_rule_scheduled_resource.go
//     (schema L45-348, body L405-449)
//   - internal/services/sentinel/sentinel_alert_rule_nrt_resource.go (schema L43-320)
//   - internal/services/sentinel/sentinel_alert_rule_ms_security_incident_resource.go (schema L41-129)
//   - internal/services/sentinel/sentinel_alert_rule_fusion_resource.go (schema L42-131)
//   - internal/services/sentinel/sentinel_alert_rule_threat_intelligence_resource.go (schema L53-82)
//   - internal/services/sentinel/sentinel_alert_rule_machine_learning_behavior_analytics_resource.go
//   - go-azure-sdk resource-manager/securityinsights/2023-12-01-preview/alertrules:
//     id_alertrule.go (Microsoft.SecurityInsights/alertRules), constants.go
//     (AlertRuleKind, AlertSeverity, TriggerOperator, MicrosoftSecurityProductName),
//     model_scheduledalertruleproperties.go, model_microsoftsecurityincidentcreationalertruleproperties.go
type AlertRules struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AlertRules)(nil)

// NewAlertRules returns knowledge for the alertRules resource type.
func NewAlertRules() *AlertRules {
	return &AlertRules{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SecurityInsights/alertRules",
			ApiVersions:  []string{"2023-12-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				// alert_rule_template_guid is ForceNew for every kind that exposes it.
				// Fires only when the property is present in both bodies.
				{PropertyPath: "properties.alertRuleTemplateName"},
			},
			// "kind" is the discriminator and is required for every alertRules body.
			RequiredFields: []string{"kind"},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					MinLength:    1,
					Message:      "alert rule name must not be empty",
				},
				{
					PropertyPath: "kind",
					// Full ARM SDK AlertRuleKind set.
					AllowedValues: []string{
						"Fusion", "MLBehaviorAnalytics", "MicrosoftSecurityIncidentCreation",
						"NRT", "Scheduled", "ThreatIntelligence",
					},
					Message: "kind must be one of Fusion, MLBehaviorAnalytics, MicrosoftSecurityIncidentCreation, NRT, Scheduled, ThreatIntelligence",
				},
				{
					// Present for Scheduled/NRT/ThreatIntelligence kinds.
					PropertyPath:  "properties.severity",
					AllowedValues: []string{"High", "Informational", "Low", "Medium"},
					Message:       "severity must be one of High, Informational, Low, Medium",
				},
				{
					// Scheduled kind only.
					PropertyPath:  "properties.triggerOperator",
					AllowedValues: []string{"Equal", "GreaterThan", "LessThan", "NotEqual"},
					Message:       "trigger_operator must be one of Equal, GreaterThan, LessThan, NotEqual",
				},
				{
					// MicrosoftSecurityIncidentCreation kind only.
					PropertyPath: "properties.productFilter",
					AllowedValues: []string{
						"Azure Active Directory Identity Protection",
						"Azure Advanced Threat Protection",
						"Azure Security Center",
						"Azure Security Center for IoT",
						"Microsoft Cloud App Security",
						"Microsoft Defender Advanced Threat Protection",
						"Office 365 Advanced Threat Protection",
					},
					Message: "product_filter must be a supported Microsoft security product name",
				},
				{
					// alert_rule_template_guid → properties.alertRuleTemplateName (validation.IsUUID).
					PropertyPath: "properties.alertRuleTemplateName",
					Regex:        `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
					Message:      "alert_rule_template_guid must be a valid UUID",
				},
			},
			IntRules: []azwise.IntRule{
				{
					// Scheduled kind only (validation.IntAtLeast(0)).
					PropertyPath: "properties.triggerThreshold",
					MinValue:     azwise.Ptr(int64(0)),
					Message:      "trigger_threshold must be at least 0",
				},
			},
		},
	}
}

func init() { azwise.Register(NewAlertRules()) }
