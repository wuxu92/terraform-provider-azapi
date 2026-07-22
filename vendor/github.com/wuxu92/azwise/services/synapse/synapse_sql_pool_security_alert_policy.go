package synapse

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SynapseSqlPoolSecurityAlertPolicy provides resource knowledge for
// Microsoft.Synapse/workspaces/sqlPools/securityAlertPolicies.
//
// Mirrors azurerm_synapse_sql_pool_security_alert_policy (ID
// .../sqlPools/{pool}/securityAlertPolicies/Default).
//
// Sources:
//   - internal/services/synapse/synapse_sql_pool_security_alert_policy_resource.go
//     (schema 41-109: policy_state Required enum (Disabled/Enabled/New);
//     disabled_alerts set of enum strings; email_addresses set; email_account_admins_enabled
//     Default false; retention_days IntAtLeast(0) Default 0; storage_account_access_key
//     Sensitive; timeouts Create/Update/Delete 30m Read 5m).
//   - go-azure-sdk resource-manager (track1) synapse model
//     ServerSecurityAlertPolicyProperties (state/disabledAlerts/emailAddresses/
//     emailAccountAdmins/storageEndpoint/storageAccountAccessKey/retentionDays json tags),
//     enums.go SecurityAlertPolicyState (Disabled/Enabled/New), resourceids.go
//     SqlPoolSecurityAlertPolicy (segment casing "securityAlertPolicies").
//
// Notes:
//   - disabled_alerts is an array of enum strings; per-element enum validation is not
//     expressible as an array-element path — documented, no rule emitted.
type SynapseSqlPoolSecurityAlertPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SynapseSqlPoolSecurityAlertPolicy)(nil)

func NewSynapseSqlPoolSecurityAlertPolicy() *SynapseSqlPoolSecurityAlertPolicy {
	return &SynapseSqlPoolSecurityAlertPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Synapse/workspaces/sqlPools/securityAlertPolicies",
			ApiVersions:  []string{"2021-06-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.state",
					AllowedValues: []string{"New", "Enabled", "Disabled"},
					Message:       "must be one of New, Enabled or Disabled",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.retentionDays",
					MinValue:     azwise.Ptr(int64(0)),
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
				{PropertyPath: "properties.retentionDays", Value: int64(0)},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewSynapseSqlPoolSecurityAlertPolicy()) }
