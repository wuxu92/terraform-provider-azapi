package monitor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DiagnosticSetting provides resource knowledge for
// Microsoft.Insights/diagnosticSettings.
//
// Mirrors azurerm_monitor_diagnostic_setting. This is a scoped (extension) resource
// attached to a target resource; its ID has the form
// /{scope}/providers/Microsoft.Insights/diagnosticSettings/{name}.
//
// Sources:
//   - terraform-provider-azurerm internal/services/monitor/monitor_diagnostic_setting_resource.go
//     Schema (50-149): name Required+ForceNew (MonitorDiagnosticSettingName);
//     target_resource_id Required+ForceNew (scope); eventhub_name → properties.eventHubName;
//     eventhub_authorization_rule_id → properties.eventHubAuthorizationRuleId;
//     log_analytics_workspace_id → properties.workspaceId; storage_account_id →
//     properties.storageAccountId; partner_solution_id → properties.marketplacePartnerId
//     (AtLeastOneOf across the four destinations); log_analytics_destination_type
//     Optional+Computed enum Dedicated/AzureDiagnostics → properties.logAnalyticsDestinationType;
//     enabled_log → properties.logs[]; enabled_metric / metric → properties.metrics[].
//     Timeouts Create/Update 30m, Read 5m, Delete 60m.
//   - internal/services/monitor/validate/monitor_diagnostic_setting.go: name disallows
//     < > * %% & : \ ? + / and must be 1-260 characters.
//   - go-azure-sdk resource-manager/insights/2021-05-01-preview/diagnosticsettings:
//     DiagnosticSettings (model_diagnosticsettings.go) property names;
//     id_scopeddiagnosticsetting.go type Microsoft.Insights/diagnosticSettings.
//
// Note: at least one destination (eventhub_authorization_rule_id / log_analytics_workspace_id
// / storage_account_id / partner_solution_id) and at least one of logs/metrics must be set;
// these AtLeastOneOf relational constraints are not expressible declaratively in azwise.
type DiagnosticSetting struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DiagnosticSetting)(nil)

// NewDiagnosticSetting returns knowledge for the diagnosticSettings resource.
func NewDiagnosticSetting() *DiagnosticSetting {
	return &DiagnosticSetting{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Insights/diagnosticSettings",
			ApiVersions:  []string{"2021-05-01-preview"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name: no < > * %% & : \ ? + / ; length 1-260.
					Regex:     `^[^<>*%&:\\?+/]+$`,
					MinLength: 1,
					MaxLength: 260,
					Message:   `name must be 1-260 characters and must not contain < > * % & : \ ? + /`,
				},
				{
					PropertyPath:  "properties.logAnalyticsDestinationType",
					AllowedValues: []string{"Dedicated", "AzureDiagnostics"},
					Message:       "log_analytics_destination_type must be one of Dedicated, AzureDiagnostics",
				},
			},
		},
	}
}

func init() { azwise.Register(NewDiagnosticSetting()) }
