package netapp

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetAppBackupPolicy provides resource knowledge for
// Microsoft.NetApp/netAppAccounts/backupPolicies.
//
// Mirrors azurerm_netapp_backup_policy.
//
// Sources:
//   - internal/services/netapp/netapp_backup_policy_resource.go
//     (schema 41-92: name ForceNew (VolumeQuotaRuleName validator); daily_backups_to_keep
//     IntBetween(2,1019) default 2; weekly/monthly_backups_to_keep IntBetween(0,1019)
//     default 1; enabled default true; timeouts Create 90m Update/Delete 120m Read 5m;
//     create 98-150 → BackupPolicyProperties{Daily/Weekly/MonthlyBackupsToKeep, Enabled}).
//   - internal/services/netapp/validate/volume_quota_rule_name.go (regex ^[a-zA-Z][-_\da-zA-Z]{0,63}$).
//   - go-azure-sdk resource-manager/netapp/2026-01-01/backuppolicies:
//     model_backuppolicyproperties.go (dailyBackupsToKeep/weeklyBackupsToKeep/
//     monthlyBackupsToKeep/enabled settable; backupPolicyId/provisioningState/
//     volumeBackups/volumesAssigned read-only),
//     id_backuppolicy.go (type segment casing "backupPolicies").
type NetAppBackupPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetAppBackupPolicy)(nil)

// NewNetAppBackupPolicy returns knowledge for the backupPolicies resource.
func NewNetAppBackupPolicy() *NetAppBackupPolicy {
	return &NetAppBackupPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.NetApp/netAppAccounts/backupPolicies",
			ApiVersions:  []string{"2026-01-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 120 * time.Minute,
				Delete: 120 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			StringRules: []azwise.StringRule{
				{
					// AzureRM uses VolumeQuotaRuleName validator: 1-64 chars, start with a letter.
					Regex:     `^[a-zA-Z][-_\da-zA-Z]{0,63}$`,
					MinLength: 1,
					MaxLength: 64,
					Message:   "must be 1-64 characters, start with a letter, and contain only letters, numbers, underscores and hyphens",
				},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.dailyBackupsToKeep", MinValue: azwise.Ptr(int64(2)), MaxValue: azwise.Ptr(int64(1019))},
				{PropertyPath: "properties.weeklyBackupsToKeep", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(1019))},
				{PropertyPath: "properties.monthlyBackupsToKeep", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(1019))},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.dailyBackupsToKeep", Value: int64(2)},
				{PropertyPath: "properties.weeklyBackupsToKeep", Value: int64(1)},
				{PropertyPath: "properties.monthlyBackupsToKeep", Value: int64(1)},
				{PropertyPath: "properties.enabled", Value: true},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.backupPolicyId",
				"properties.provisioningState",
				"properties.volumeBackups",
				"properties.volumesAssigned",
			},
			// NOTE: ValidateNetAppBackupPolicyCombinedRetention enforces a cross-field
			// invariant over daily/weekly/monthly retention counts that cannot be expressed
			// as independent declarative rules.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewNetAppBackupPolicy()) }
