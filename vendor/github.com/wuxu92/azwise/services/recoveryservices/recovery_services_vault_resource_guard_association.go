package recoveryservices

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VaultResourceGuardProxy provides resource knowledge for
// Microsoft.RecoveryServices/vaults/backupResourceGuardProxies.
//
// Contributing TF resource:
//   - azurerm_recovery_services_vault_resource_guard_association
//
// The association is modelled by AzureRM as a backupResourceGuardProxies child of a
// recovery services vault (NOT a DataProtection resource). The proxy body carries a
// single settable property, resourceGuardResourceId, which references the target
// resource guard and is ForceNew (the association is immutable — Create/Read/Delete
// only, no Update).
//
// Sources:
//   - AzureRM internal/services/recoveryservices/recovery_services_vault_resource_guard_association_resource.go
//     :25 (ARM type constant), :52-54 (vault_id + resource_guard_id both RequiredForceNew),
//     :66 (create 30m), :98-100 (properties.resourceGuardResourceId), :115 (read 5m),
//     :154 (delete 30m).
//   - go-azure-sdk resource-manager/recoveryservicesbackup/2023-02-01/resourceguardproxy:
//     id_backupresourceguardproxy.go:109-127 (ARM path/segment casing
//     .../vaults/{vaultName}/backupResourceGuardProxies/{name}),
//     model_resourceguardproxybase.go:10 (resourceGuardResourceId json tag).
//
// Not encoded (deliberate):
//   - ForceNew on vault_id is the parent-reference envelope, not a body path.
type VaultResourceGuardProxy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VaultResourceGuardProxy)(nil)

// NewVaultResourceGuardProxy returns knowledge for the backupResourceGuardProxies resource.
func NewVaultResourceGuardProxy() *VaultResourceGuardProxy {
	return &VaultResourceGuardProxy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.RecoveryServices/vaults/backupResourceGuardProxies",
			ApiVersions:  []string{"2023-02-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				// No Update: the association is immutable (replace-only).
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.resourceGuardResourceId",
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.resourceGuardResourceId"},
			},
		},
	}
}

func init() { azwise.Register(NewVaultResourceGuardProxy()) }
