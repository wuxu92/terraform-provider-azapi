// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package sentinel

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SecurityMLAnalyticsSettings provides resource knowledge for
// Microsoft.SecurityInsights/securityMLAnalyticsSettings (an extension resource
// on a Microsoft.OperationalInsights/workspaces scope). The only ARM kind is
// "Anomaly".
//
// Contributing Terraform resources (both map to the securityMLAnalyticsSettings
// ARM type — verified via securitymlanalyticssettings.ValidateSecurityMLAnalyticsSettingID):
//   - azurerm_sentinel_alert_rule_anomaly_built_in
//   - azurerm_sentinel_alert_rule_anomaly_duplicate
//
// Merge policy: only universal knowledge is unioned. display_name is ForceNew on
// built_in but not on duplicate, so it is NOT emitted as a ForceNew rule.
//
// Sources:
//   - terraform-provider-azurerm internal/services/sentinel/sentinel_alert_rule_anomaly_built_in_resource.go
//     (schema L58-99, mode=SettingsStatus)
//   - internal/services/sentinel/sentinel_alert_rule_anomaly_duplicate_resource.go (schema L61-97)
//   - go-azure-sdk resource-manager/securityinsights/2022-10-01-preview/securitymlanalyticssettings:
//     id_securitymlanalyticssetting.go (Microsoft.SecurityInsights/securityMLAnalyticsSettings),
//     constants.go (SecurityMLAnalyticsSettingsKind=Anomaly, SettingsStatus),
//     model_anomalysecuritymlanalyticssettingsproperties.go (settingsStatus, displayName, enabled)
type SecurityMLAnalyticsSettings struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SecurityMLAnalyticsSettings)(nil)

// NewSecurityMLAnalyticsSettings returns knowledge for the securityMLAnalyticsSettings resource type.
func NewSecurityMLAnalyticsSettings() *SecurityMLAnalyticsSettings {
	return &SecurityMLAnalyticsSettings{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SecurityInsights/securityMLAnalyticsSettings",
			ApiVersions:  []string{"2022-10-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// "kind" is the discriminator (always Anomaly); settingsStatus is set by
			// both contributing resources.
			RequiredFields: []string{"kind", "properties.settingsStatus"},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					MinLength:    1,
					Message:      "security ML analytics setting name must not be empty",
				},
				{
					PropertyPath:  "kind",
					AllowedValues: []string{"Anomaly"},
					Message:       "kind must be Anomaly",
				},
				{
					// mode → properties.settingsStatus.
					PropertyPath:  "properties.settingsStatus",
					AllowedValues: []string{"Flighting", "Production"},
					Message:       "mode must be one of Flighting, Production",
				},
			},
		},
	}
}

func init() { azwise.Register(NewSecurityMLAnalyticsSettings()) }
