package monitor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AadDiagnosticSetting provides resource knowledge for the Azure Active Directory
// diagnostic settings resource (ARM provider Microsoft.Aadiam).
//
// Mirrors azurerm_monitor_aad_diagnostic_setting. This is a tenant-level resource; its
// ID has the form /providers/Microsoft.AADIAM/diagnosticSettings/{name} (the go-azure-sdk
// id parser renders the provider segment as "Microsoft.AADIAM"; azwise lookup is
// case-insensitive).
//
// Sources:
//   - terraform-provider-azurerm internal/services/monitor/monitor_aad_diagnostic_setting_resource.go
//     Schema (48-102): name Required+ForceNew (MonitorDiagnosticSettingName);
//     eventhub_name Optional+ForceNew (StringMatch name regex) → properties.eventHubName;
//     eventhub_authorization_rule_id Optional+ForceNew → properties.eventHubAuthorizationRuleId;
//     log_analytics_workspace_id Optional → properties.workspaceId; storage_account_id
//     Optional+ForceNew → properties.storageAccountId (AtLeastOneOf across the three
//     destinations); enabled_log → properties.logs[]. Timeouts Create/Update/Delete 30m,
//     Read 5m.
//   - internal/services/monitor/validate/monitor_diagnostic_setting.go: name disallows
//     < > * %% & : \ ? + / and must be 1-260 characters.
//   - go-azure-sdk resource-manager/azureactivedirectory/2017-04-01/diagnosticsettings:
//     DiagnosticSettings (model_diagnosticsettings.go) property names;
//     id_diagnosticsetting.go provider Microsoft.AADIAM.
//
// Note: at least one destination must be set (AtLeastOneOf) and at least one enabled_log
// entry is required; these relational constraints are not expressible declaratively in
// azwise. logs[*].category is an array-element enum and is intentionally not emitted.
type AadDiagnosticSetting struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AadDiagnosticSetting)(nil)

// NewAadDiagnosticSetting returns knowledge for the AADIAM diagnosticSettings resource.
func NewAadDiagnosticSetting() *AadDiagnosticSetting {
	return &AadDiagnosticSetting{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Aadiam/diagnosticSettings",
			ApiVersions:  []string{"2017-04-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.eventHubName"},
				{PropertyPath: "properties.eventHubAuthorizationRuleId"},
				{PropertyPath: "properties.storageAccountId"},
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
					PropertyPath: "properties.eventHubName",
					Regex:        `^[a-zA-Z0-9]([-._a-zA-Z0-9]{0,48}[a-zA-Z0-9])?$`,
					Message:      "eventHubName may contain only letters, numbers, periods, hyphens and underscores, up to 50 characters, beginning and ending with a letter or number",
				},
			},
		},
	}
}

func init() { azwise.Register(NewAadDiagnosticSetting()) }
