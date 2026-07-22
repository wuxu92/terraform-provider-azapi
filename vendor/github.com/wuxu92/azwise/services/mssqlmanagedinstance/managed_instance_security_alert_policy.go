// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package mssqlmanagedinstance

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ManagedInstanceSecurityAlertPolicy provides resource knowledge for
// Microsoft.Sql/managedInstances/securityAlertPolicies
// (Terraform azurerm_mssql_managed_instance_security_alert_policy).
//
// The ARM resource name is the constant "Default". Body is
// {properties:{state, disabledAlerts[], emailAddresses[], emailAccountAdmins,
// retentionDays, storageAccountAccessKey, storageEndpoint}} against the go-azure-sdk
// managedserversecurityalertpolicies.ManagedServerSecurityAlertPolicy model.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mssqlmanagedinstance/mssql_managed_instance_security_alert_policy_resource.go:43-109
//     (Schema: disabled_alerts enum, retention_days IntAtLeast(0), sensitive key)
//   - .../mssql_managed_instance_security_alert_policy_resource.go:343-391 (expand → properties)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/managedserversecurityalertpolicies:
//     model_securityalertspolicyproperties.go:12-21, constants.go (SecurityAlertsPolicyState)
//
// Intentionally not encoded (documented, not emitted):
//   - disabled_alerts is a SET of enum strings mapping to
//     properties.disabledAlerts[*]; array-element enum constraints are not
//     expressible as declarative StringRules, so the allowed set
//     (Sql_Injection, Sql_Injection_Vulnerability, Access_Anomaly,
//     Data_Exfiltration, Unsafe_Action, Brute_Force) is documented here only.
//   - managed_instance_name (parent) is envelope-owned RequiresReplace.
type ManagedInstanceSecurityAlertPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ManagedInstanceSecurityAlertPolicy)(nil)

// NewManagedInstanceSecurityAlertPolicy returns knowledge for
// Microsoft.Sql/managedInstances/securityAlertPolicies.
func NewManagedInstanceSecurityAlertPolicy() *ManagedInstanceSecurityAlertPolicy {
	return &ManagedInstanceSecurityAlertPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/managedInstances/securityAlertPolicies",
			ApiVersions:  []string{"2023-08-01-preview"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{},
			StringRules: []azwise.StringRule{
				// storage_account_access_key → properties.storageAccountAccessKey (StringIsNotEmpty).
				{PropertyPath: "properties.storageAccountAccessKey", MinLength: 1, Message: "storage_account_access_key must not be empty"},
				// storage_endpoint → properties.storageEndpoint (StringIsNotEmpty).
				{PropertyPath: "properties.storageEndpoint", MinLength: 1, Message: "storage_endpoint must not be empty"},
			},
			IntRules: []azwise.IntRule{
				// retention_days → properties.retentionDays (IntAtLeast 0).
				{PropertyPath: "properties.retentionDays", MinValue: azwise.Ptr(int64(0)), Message: "retention_days must be at least 0"},
			},
			FloatRules:      []azwise.FloatRule{},
			ArrayRules:      []azwise.ArrayRule{},
			SensitiveFields: []string{"properties.storageAccountAccessKey"},
			ComputedFields:  []string{},
			DefaultValues: []azwise.DefaultValue{
				// email_account_admins_enabled default false; retention_days default 0.
				{PropertyPath: "properties.emailAccountAdmins", Value: false},
				{PropertyPath: "properties.retentionDays", Value: float64(0)},
			},
			// state is Required (json:"state" without omitempty); derived from `enabled`.
			RequiredFields: []string{"properties.state"},
		},
	}
}

func init() { azwise.Register(NewManagedInstanceSecurityAlertPolicy()) }
