package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MsSqlServerSecurityAlertPolicy provides resource knowledge for
// Microsoft.Sql/servers/securityAlertPolicies
// (azurerm_mssql_server_security_alert_policy).
//
// Sources:
//   - AzureRM internal/services/mssql/mssql_server_security_alert_policy_resource.go
//     :37-42   (timeouts: create/update/delete 30m, read 5m)
//     :47-52   (server_name ForceNew — envelope)
//     :70-112  (email_account_admins_enabled default false; retention_days IntAtLeast(0);
//     state Required StringInSlice(SecurityAlertsPolicyState); storage_account_access_key Sensitive)
//     :151-199 (create mapping: properties.state / disabledAlerts / emailAddresses /
//     emailAccountAdmins / retentionDays / storageAccountAccessKey / storageEndpoint)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/serversecurityalertpolicies:
//     model_securityalertspolicyproperties.go, constants.go (SecurityAlertsPolicyState)
//
// NOTE: disabled_alerts is an array of enum strings (properties.disabledAlerts[*]); the
// per-element enum check is not expressible as a declarative StringRule and is skipped.
type MsSqlServerSecurityAlertPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MsSqlServerSecurityAlertPolicy)(nil)

// NewMsSqlServerSecurityAlertPolicy returns knowledge for the
// servers/securityAlertPolicies resource.
func NewMsSqlServerSecurityAlertPolicy() *MsSqlServerSecurityAlertPolicy {
	return &MsSqlServerSecurityAlertPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/servers/securityAlertPolicies",
			ApiVersions:  []string{"2023-08-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.state",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "state must be one of Disabled or Enabled",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.retentionDays",
					MinValue:     azwise.Ptr(int64(0)),
					Message:      "retention_days must be at least 0",
				},
			},
			SensitiveFields: []string{
				"properties.storageAccountAccessKey",
			},
			RequiredFields: []string{
				"properties.state",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.emailAccountAdmins", Value: false},
				{PropertyPath: "properties.retentionDays", Value: float64(0)},
			},
		},
	}
}

func init() { azwise.Register(NewMsSqlServerSecurityAlertPolicy()) }
