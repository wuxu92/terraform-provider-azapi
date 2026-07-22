package recoveryservices

import (
	"time"

	"github.com/wuxu92/azwise"
)

// BackupProtectionPolicy provides resource knowledge for
// Microsoft.RecoveryServices/vaults/backupPolicies.
//
// This single ARM type is exposed by AzureRM as THREE typed Terraform resources,
// each a variant of the same protectionPolicies body — a discriminated union
// (BaseProtectionPolicy interface) keyed by `backupManagementType`. They are merged
// here; only knowledge universal to every variant is encoded. Contributing TF
// resources:
//   - azurerm_backup_policy_vm            (backupManagementType "AzureIaasVM")
//   - azurerm_backup_policy_file_share    (backupManagementType "AzureStorage")
//   - azurerm_backup_policy_vm_workload   (backupManagementType "AzureWorkload")
//
// Sources:
//   - AzureRM internal/services/recoveryservices/backup_policy_vm_resource.go
//     :43-47 (timeouts 30m/5m/30m/30m), :1061-1068 (name regex
//     ^[a-zA-Z][-_!a-zA-Z0-9]{2,149}$, ForceNew).
//   - internal/services/recoveryservices/backup_policy_file_share_resource.go
//     :41-45 (timeouts), :695-702 (name regex ^[a-zA-Z][-_!a-zA-Z0-9]{2,149}$, ForceNew).
//   - internal/services/recoveryservices/backup_policy_vm_workload_resource.go
//     :91-93 (TF type), :105-110 (name validate.BackupPolicyName, ForceNew),
//     :419/:468/:517/:557 (timeouts 30m/5m/30m/30m).
//   - internal/services/recoveryservices/validate/backup_policy_name.go:14
//     (workload name regex ^[a-zA-Z][-\da-zA-Z]{2,149}$).
//   - go-azure-sdk resource-manager/recoveryservicesbackup/2024-10-01/protectionpolicies:
//     id_backuppolicy.go:109-127 (ARM path/segment casing
//     .../vaults/{vaultName}/backupPolicies/{name}),
//     model_protectionpolicy.go:12-108 (BaseProtectionPolicy discriminated union,
//     backupManagementType discriminator: AzureIaasVM/AzureSql/AzureStorage/
//     AzureWorkload/GenericProtectionPolicy/MAB).
//
// Not encoded (deliberate):
//   - properties is a discriminated-union interface (BaseProtectionPolicy) whose
//     concrete shape (schedule/retention rule tree, all inside arrays) differs per
//     backupManagementType. Paths under it cannot be resolved through the interface,
//     so no body-path ForceNew/enum/range rules are emitted.
//   - The name regex differs per variant (vm/file_share allow `-_!`; workload allows
//     only `-` plus alphanumerics). Only the shared "start with a letter, 3-150 chars"
//     bound is universal, so a length-only StringRule is emitted rather than a single
//     regex that would over-reject valid names of another variant.
//   - `name`/`policy_type` ForceNew are the resource-name envelope and a kind-specific
//     (VM-only) body discriminator respectively; neither is universal.
type BackupProtectionPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*BackupProtectionPolicy)(nil)

// NewBackupProtectionPolicy returns knowledge for the backupPolicies resource.
func NewBackupProtectionPolicy() *BackupProtectionPolicy {
	return &BackupProtectionPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.RecoveryServices/vaults/backupPolicies",
			ApiVersions:  []string{"2024-10-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name: universal 3-150 length bound across all variants.
					MinLength: 3,
					MaxLength: 150,
					Message:   "Backup Policy name must be 3-150 characters long and start with a letter",
				},
				{
					// Discriminator present in every variant's body.
					PropertyPath:  "properties.backupManagementType",
					AllowedValues: []string{"AzureIaasVM", "AzureSql", "AzureStorage", "AzureWorkload", "GenericProtectionPolicy", "MAB"},
					Message:       "properties.backupManagementType must be one of AzureIaasVM, AzureSql, AzureStorage, AzureWorkload, GenericProtectionPolicy or MAB",
				},
			},
		},
	}
}

func init() { azwise.Register(NewBackupProtectionPolicy()) }
