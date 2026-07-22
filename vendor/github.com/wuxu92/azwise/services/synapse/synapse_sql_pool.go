package synapse

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SynapseSqlPool provides resource knowledge for
// Microsoft.Synapse/workspaces/sqlPools.
//
// Mirrors azurerm_synapse_sql_pool.
//
// Sources:
//   - internal/services/synapse/synapse_sql_pool_resource.go
//     (schema 61-183: name ForceNew (SqlPoolName validator); sku_name Required enum;
//     storage_account_type Required enum (LRS/GRS) ForceNew; create_mode enum
//     (Default/Recovery/PointInTimeRestore) Default Default ForceNew; collation O+C ForceNew;
//     geo_backup_policy_enabled Default true; recovery_database_id ConflictsWith restore;
//     restore ForceNew; timeouts Create/Update/Delete 30m Read 5m; create 239-278 →
//     SQLPool{Sku, SQLPoolResourceProperties}).
//   - go-azure-sdk resource-manager (track1) synapse models SQLPool / Sku /
//     SQLPoolResourceProperties (sku.name/collation/createMode/storageAccountType/
//     recoverableDatabaseId/sourceDatabaseId json tags), enums.go StorageAccountType
//     (GRS/LRS) + CreateMode (Default/PointInTimeRestore/Recovery/Restore);
//     resourceids.go SqlPool (segment casing "sqlPools").
//
// Notes:
//   - data_encrypted is applied via the separate transparentDataEncryption sub-API and
//     geo_backup_policy_enabled via the geoBackupPolicies sub-API — neither is part of the
//     sqlPools body, so both are omitted here.
//   - recovery_database_id / restore ConflictsWith: recovery_database_id maps to
//     properties.recoverableDatabaseId and restore maps to properties.sourceDatabaseId +
//     properties.restorePointInTime (create_mode discriminated), so the conflict is
//     expressible; emitted as a ConflictsWith relational rule.
//   - StorageAccountType enum uses the full ARM SDK set (GRS/LRS); AzureRM restricts to the
//     same two.
//   - collation is ForceNew and optional+computed (server/user-configurable default).
type SynapseSqlPool struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SynapseSqlPool)(nil)

func NewSynapseSqlPool() *SynapseSqlPool {
	return &SynapseSqlPool{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Synapse/workspaces/sqlPools",
			ApiVersions:  []string{"2021-06-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.storageAccountType"},
				{PropertyPath: "properties.createMode"},
				{PropertyPath: "properties.collation"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "sku.name",
					AllowedValues: []string{
						"DW100c", "DW200c", "DW300c", "DW400c", "DW500c", "DW1000c",
						"DW1500c", "DW2000c", "DW2500c", "DW3000c", "DW5000c", "DW6000c",
						"DW7500c", "DW10000c", "DW15000c", "DW30000c",
					},
					Message: "must be a valid dedicated SQL pool SKU (e.g. DW100c)",
				},
				{
					PropertyPath:  "properties.storageAccountType",
					AllowedValues: []string{"GRS", "LRS"},
					Message:       "must be one of GRS or LRS",
				},
				{
					PropertyPath:  "properties.createMode",
					AllowedValues: []string{"Default", "PointInTimeRestore", "Recovery", "Restore"},
					Message:       "must be one of Default, PointInTimeRestore, Recovery or Restore",
				},
			},
			ConflictsWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.recoverableDatabaseId", "properties.sourceDatabaseId"},
					Message: "recovery (recoverableDatabaseId) and point-in-time restore (sourceDatabaseId) cannot be combined",
				},
			},
			RequiredFields: []string{
				"sku.name",
				"properties.storageAccountType",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.createMode", Value: "Default"},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewSynapseSqlPool()) }
