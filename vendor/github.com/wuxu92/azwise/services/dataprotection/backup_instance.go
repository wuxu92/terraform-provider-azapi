package dataprotection

import (
	"time"

	"github.com/wuxu92/azwise"
)

// BackupInstance provides resource knowledge for
// Microsoft.DataProtection/backupVaults/backupInstances.
//
// This ARM type is exposed by AzureRM as SEVEN typed Terraform resources, each a
// datasource-type variant of the same backupInstanceResources body. They are
// merged here; only knowledge universal to every variant is encoded. Contributing
// TF resources:
//   - azurerm_data_protection_backup_instance_blob_storage
//   - azurerm_data_protection_backup_instance_data_lake_storage
//   - azurerm_data_protection_backup_instance_disk
//   - azurerm_data_protection_backup_instance_kubernetes_cluster
//   - azurerm_data_protection_backup_instance_mysql_flexible_server
//   - azurerm_data_protection_backup_instance_postgresql_flexible_server
//   - azurerm_data_protection_backup_instance_postgresql (deprecated, removed in v5.0)
//
// Sources:
//   - AzureRM internal/services/dataprotection/data_protection_backup_instance_*_resource.go
//     (all seven) — timeouts create/update/delete 30m, read 5m; name/vault_id/
//     datasource-id ForceNew; backup_policy_id -> properties.policyInfo.policyId;
//     protection_state (Computed) -> properties.currentProtectionState.
//     disk_resource.go:154-178 (expand: dataSourceInfo.resourceID, friendlyName,
//     policyInfo.policyId), :106-109 (protection_state computed).
//     mysql/postgresql_flexible_server_resource.go:120-135 (dataSourceInfo +
//     dataSourceSetInfo both -> resourceID).
//   - go-azure-sdk resource-manager/dataprotection/2025-07-01/backupinstanceresources:
//     model_backupinstance.go, model_backupinstanceresource.go, model_datasource.go,
//     model_policyinfo.go, constants.go (CurrentProtectionState, ValidationType).
//
// Not encoded (deliberate):
//   - backup_policy_id (properties.policyInfo.policyId) is ForceNew ONLY for the
//     kubernetes_cluster variant; it is updatable for the others. Non-universal, so
//     it is left out of the shared ForceNew list.
//   - Datasource-specific ForceNew fields (disk_id, storage_account_id,
//     kubernetes_cluster_id, database_id, server_id, snapshot_subscription_id) all
//     expand to properties.dataSourceInfo.resourceID, which IS the universal ForceNew
//     encoded below.
//   - backup_datasource_parameters (kubernetes only) is a variant-specific ForceNew
//     block; not universal.
//   - The datasource-specific resource-ID validators (managed disk / storage account
//     / AKS / flexible server / database IDs) are semantic ARM-ID validators, not
//     declarative rules; they are variant-specific and left to residual customizers.
type BackupInstance struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*BackupInstance)(nil)

// NewBackupInstance returns knowledge for the backupInstances resource.
func NewBackupInstance() *BackupInstance {
	return &BackupInstance{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DataProtection/backupVaults/backupInstances",
			ApiVersions:  []string{"2025-07-01"},
			SoftDelete:   false,
			// name is envelope-owned. The datasource being protected
			// (properties.dataSourceInfo.resourceID) is ForceNew across every
			// variant — the protected resource cannot be swapped in place.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.dataSourceInfo.resourceID"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.validationType",
					AllowedValues: []string{"DeepValidation", "ShallowValidation"},
					Message:       "must be DeepValidation or ShallowValidation",
				},
			},
			// Response-only properties Azure populates in the GET body.
			ComputedFields: []string{
				"properties.currentProtectionState",
				"properties.protectionStatus",
				"properties.protectionErrorDetails",
				"properties.provisioningState",
			},
			// Universal required body paths (present for every variant): the
			// discriminator, the protected datasource, and the associated policy.
			RequiredFields: []string{
				"properties.objectType",
				"properties.dataSourceInfo.resourceID",
				"properties.policyInfo.policyId",
			},
		},
	}
}

func init() { azwise.Register(NewBackupInstance()) }
