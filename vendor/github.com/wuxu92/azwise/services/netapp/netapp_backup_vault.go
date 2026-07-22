package netapp

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetAppBackupVault provides resource knowledge for
// Microsoft.NetApp/netAppAccounts/backupVaults.
//
// Mirrors azurerm_netapp_backup_vault.
//
// Sources:
//   - internal/services/netapp/netapp_backup_vault_resource.go
//     (schema 42-64: name ForceNew (VolumeQuotaRuleName validator); location/account_name/
//     tags only; no user-settable body properties; timeouts Create 90m Update/Delete 120m
//     Read 5m; create 70-111 → BackupVault{Location, Tags}).
//   - internal/services/netapp/validate/volume_quota_rule_name.go (regex ^[a-zA-Z][-_\da-zA-Z]{0,63}$).
//   - go-azure-sdk resource-manager/netapp/2026-01-01/backupvaults:
//     model_backupvaultproperties.go (provisioningState read-only, no settable properties),
//     id_backupvault.go (type segment casing "backupVaults").
type NetAppBackupVault struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetAppBackupVault)(nil)

// NewNetAppBackupVault returns knowledge for the backupVaults resource.
func NewNetAppBackupVault() *NetAppBackupVault {
	return &NetAppBackupVault{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.NetApp/netAppAccounts/backupVaults",
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
			// Server-populated, read-only ARM property.
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewNetAppBackupVault()) }
