package recoveryservices

import (
	"time"

	"github.com/wuxu92/azwise"
)

// BackupProtectedItem provides resource knowledge for
// Microsoft.RecoveryServices/vaults/backupFabrics/protectionContainers/protectedItems.
//
// This single ARM type is exposed by AzureRM as TWO typed Terraform resources, each a
// variant of the same protectedItems body — a discriminated union (BaseProtectedItem
// interface) keyed by `protectedItemType`. They are merged here; only knowledge
// universal to every variant is encoded. Contributing TF resources:
//   - azurerm_backup_protected_vm          (AzureIaaSComputeVMProtectedItem)
//   - azurerm_backup_protected_file_share  (AzureFileshareProtectedItem)
//
// Sources:
//   - AzureRM internal/services/recoveryservices/backup_protected_vm_resource.go
//     :46-50 (timeouts create 120m / read 5m / update 120m / delete 80m),
//     :57-59 (source_vm_id conditional ForceNew), :425-441 (source_vm_id / backup_policy_id),
//     :468-474 (protection_state enum, Optional+Computed).
//   - internal/services/recoveryservices/backup_protected_file_share_resource.go
//     :46-50 (timeouts create/update/delete 80m, read 5m), :59-80 (recovery_vault_name /
//     source_storage_account_id / source_file_share_name ForceNew, backup_policy_id).
//   - go-azure-sdk resource-manager/recoveryservicesbackup/2023-02-01/protecteditems:
//     id_protecteditem.go:121-143 (ARM path/segment casing
//     .../vaults/{vaultName}/backupFabrics/{fabric}/protectionContainers/{container}/protectedItems/{name}),
//     model_protecteditem.go:12-37 (BaseProtectedItem discriminated union keyed by
//     protectedItemType; common base fields policyId, protectedItemType).
//
// Not encoded (deliberate):
//   - properties is a discriminated-union interface (BaseProtectedItem) whose concrete
//     shape differs per protectedItemType. Paths under it cannot be resolved through the
//     interface, so no body-path ForceNew/enum/range rules are emitted.
//   - The resource name is a computed composite identifier (e.g.
//     "vm;iaasvmcontainerv2;rg;vmName"), not a user-supplied value, so no name StringRule
//     applies.
//   - source_vm_id / source_storage_account_id / source_file_share_name (schema ForceNew)
//     identify the parent container / protected item and compose the resource ID rather
//     than mapping to settable body properties; source_vm_id is additionally only
//     conditionally ForceNew (replaces only when set to a new non-empty value).
//   - Timeouts differ between the two variants (VM create/update 120m, file share 80m).
//     The larger bound is used so neither variant is starved; read 5m and delete 80m are
//     identical across both.
type BackupProtectedItem struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*BackupProtectedItem)(nil)

// NewBackupProtectedItem returns knowledge for the protectedItems resource.
func NewBackupProtectedItem() *BackupProtectedItem {
	return &BackupProtectedItem{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.RecoveryServices/vaults/backupFabrics/protectionContainers/protectedItems",
			ApiVersions:  []string{"2023-02-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 120 * time.Minute,
				Read:   5 * time.Minute,
				Update: 120 * time.Minute,
				Delete: 80 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewBackupProtectedItem()) }
