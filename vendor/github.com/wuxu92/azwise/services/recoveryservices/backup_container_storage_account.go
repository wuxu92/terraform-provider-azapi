package recoveryservices

import (
	"time"

	"github.com/wuxu92/azwise"
)

// BackupProtectionContainer provides resource knowledge for
// Microsoft.RecoveryServices/vaults/backupFabrics/protectionContainers.
//
// Contributing TF resource:
//   - azurerm_backup_container_storage_account (AzureStorageContainer)
//
// Sources:
//   - AzureRM internal/services/recoveryservices/backup_container_storage_account_resource.go
//     :36-39 (timeouts create 30m / read 5m / delete 30m; no update),
//     :45-56 (recovery_vault_name + storage_account_id both Required+ForceNew).
//   - go-azure-sdk resource-manager/recoveryservicesbackup/2023-02-01/protectioncontainers:
//     id_protectioncontainer.go:115-135 (ARM path/segment casing
//     .../vaults/{vaultName}/backupFabrics/{fabric}/protectionContainers/{name}).
//
// Not encoded (deliberate):
//   - properties is a discriminated-union interface (ProtectionContainer) whose concrete
//     shape (AzureStorageContainer) is selected by containerType; paths under it cannot be
//     resolved through the interface, so no body-path rules are emitted.
//   - The resource name is a computed composite identifier
//     ("StorageContainer;Storage;{rg};{account}"), not user-supplied, so no name
//     StringRule applies.
//   - storage_account_id (schema ForceNew) identifies the container source and composes
//     the resource ID rather than mapping to a settable top-level body property.
type BackupProtectionContainer struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*BackupProtectionContainer)(nil)

// NewBackupProtectionContainer returns knowledge for the protectionContainers resource.
func NewBackupProtectionContainer() *BackupProtectionContainer {
	return &BackupProtectionContainer{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.RecoveryServices/vaults/backupFabrics/protectionContainers",
			ApiVersions:  []string{"2023-02-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				// No Update: AzureRM defines only Create/Read/Delete.
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewBackupProtectionContainer()) }
