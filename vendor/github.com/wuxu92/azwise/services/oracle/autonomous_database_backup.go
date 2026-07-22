// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oracle

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutonomousDatabaseBackup provides resource knowledge for the child ARM type
// Oracle.Database/autonomousDatabases/autonomousDatabaseBackups.
//
// The SDK id parser (autonomousdatabasebackups/id_autonomousdatabasebackup.go)
// nests the backup under its parent autonomousDatabases resource:
//   .../Oracle.Database/autonomousDatabases/{db}/autonomousDatabaseBackups/{name}
//
// Contributing TF resource:
//   - azurerm_oracle_autonomous_database_backup (oracle_autonomous_database_backup_resource.go)
//
// Sources:
//   - AzureRM oracle_autonomous_database_backup_resource.go (schema + Create mapping)
//   - Azure SDK oracledatabase/2025-09-01/autonomousdatabasebackups
//     model_autonomousdatabasebackupproperties.go + constants.go
type AutonomousDatabaseBackup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutonomousDatabaseBackup)(nil)

// NewAutonomousDatabaseBackup returns an AutonomousDatabaseBackup knowledge instance.
func NewAutonomousDatabaseBackup() *AutonomousDatabaseBackup {
	return &AutonomousDatabaseBackup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Oracle.Database/autonomousDatabases/autonomousDatabaseBackups",
			ApiVersions:  []string{"2025-09-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				// backup type is ForceNew (schema "type" -> properties.backupType)
				{PropertyPath: "properties.backupType"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.retentionPeriodInDays",
			},
			StringRules: []azwise.StringRule{
				// ── Resource name (validate.AutonomousDatabaseName) ──
				{
					Regex:     `^[a-zA-Z][a-zA-Z0-9]*$`,
					MaxLength: 30,
					Message:   "must start with a letter and contain only letters and numbers, 30 characters max",
				},
				// ── properties.backupType (AutonomousDatabaseBackupType) ──
				// AzureRM only allows "LongTerm" on create; azwise uses the full ARM
				// SDK set since AzAPI sends raw ARM values.
				{
					PropertyPath:  "properties.backupType",
					AllowedValues: []string{"Full", "Incremental", "LongTerm"},
				},
			},
			IntRules: []azwise.IntRule{
				// retention_period_in_days — IntBetween(90, 3650)
				{PropertyPath: "properties.retentionPeriodInDays", MinValue: azwise.Ptr(int64(90)), MaxValue: azwise.Ptr(int64(3650))},
			},
			// Server-computed read-only properties (SDK AutonomousDatabaseBackupProperties).
			ComputedFields: []string{
				"properties.autonomousDatabaseOcid",
				"properties.databaseSizeInTbs",
				"properties.dbVersion",
				"properties.displayName",
				"properties.isAutomatic",
				"properties.isRestorable",
				"properties.lifecycleDetails",
				"properties.lifecycleState",
				"properties.ocid",
				"properties.provisioningState",
				"properties.sizeInTbs",
				"properties.timeAvailableTil",
				"properties.timeEnded",
				"properties.timeStarted",
			},
		},
	}
}

func init() { azwise.Register(NewAutonomousDatabaseBackup()) }
