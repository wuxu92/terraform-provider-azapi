package dataprotection

import (
	"time"

	"github.com/wuxu92/azwise"
)

// BackupPolicy provides resource knowledge for
// Microsoft.DataProtection/backupVaults/backupPolicies.
//
// This ARM type is exposed by AzureRM as SEVEN typed Terraform resources, each a
// variant of the same baseBackupPolicyResources body (a discriminated union whose
// only concrete objectType today is "BackupPolicy"). They are merged here; only
// knowledge universal to every variant is encoded. Contributing TF resources:
//   - azurerm_data_protection_backup_policy_blob_storage
//   - azurerm_data_protection_backup_policy_data_lake_storage
//   - azurerm_data_protection_backup_policy_disk
//   - azurerm_data_protection_backup_policy_kubernetes_cluster
//   - azurerm_data_protection_backup_policy_mysql_flexible_server
//   - azurerm_data_protection_backup_policy_postgresql_flexible_server
//   - azurerm_data_protection_backup_policy_postgresql (deprecated, removed in v5.0)
//
// Every variant is fully immutable: AzureRM defines only Create/Read/Delete (no
// Update), so any change replaces the policy. All schema fields are ForceNew.
//
// Sources:
//   - AzureRM internal/services/dataprotection/data_protection_backup_policy_*_resource.go
//     (all seven) — timeouts create/delete 30m, read 5m (no update); name/vault_id
//     and the entire rule tree ForceNew.
//     blob_storage_resource.go:34-38 (timeouts), :46-61 (name 3-150 regex, vault_id).
//     data_lake_storage_resource.go:71-81 (name ^[a-zA-Z][-a-zA-Z0-9]{2,149}$).
//   - go-azure-sdk resource-manager/dataprotection/2025-07-01/basebackuppolicyresources:
//     model_basebackuppolicyresource.go, model_basebackuppolicy.go (BaseBackupPolicy
//     interface / discriminated union), constants.go.
//
// Not encoded (deliberate):
//   - properties is a discriminated-union interface (BaseBackupPolicy) in the SDK
//     model. Paths under it (properties.datasourceTypes, properties.policyRules[*],
//     etc.) cannot be resolved through the interface, and the rule tree lives inside
//     arrays. So no body-path ForceNew/enum/range rules are emitted — the whole body
//     is replace-only regardless.
//   - The name regex differs per variant (most use ^[-a-zA-Z0-9]{3,150}$;
//     data_lake_storage requires a leading letter: ^[a-zA-Z][-a-zA-Z0-9]{2,149}$).
//     Only the shared 3-150 length bound is universal, so a length-only StringRule is
//     emitted instead of a single regex.
type BackupPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*BackupPolicy)(nil)

// NewBackupPolicy returns knowledge for the backupPolicies resource.
func NewBackupPolicy() *BackupPolicy {
	return &BackupPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DataProtection/backupVaults/backupPolicies",
			ApiVersions:  []string{"2025-07-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				// No Update: the resource is immutable (replace-only).
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name: universal 3-150 length bound across all variants.
					MinLength: 3,
					MaxLength: 150,
					Message:   "DataProtection BackupPolicy name must be 3-150 characters long and contain only letters, numbers and hyphens",
				},
			},
		},
	}
}

func init() { azwise.Register(NewBackupPolicy()) }
