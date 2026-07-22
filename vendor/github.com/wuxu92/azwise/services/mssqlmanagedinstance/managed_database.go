// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package mssqlmanagedinstance

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ManagedDatabase provides resource knowledge for Microsoft.Sql/managedInstances/databases
// (Terraform azurerm_mssql_managed_database).
//
// The ARM body is {location, tags, properties:{...}} against the go-azure-sdk
// manageddatabases.ManagedDatabase model. location is taken from the parent
// managed instance (not user input) and tags/name are envelope-owned.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mssqlmanagedinstance/mssql_managed_database_resource.go:70-227
//     (Arguments: name/point_in_time_restore ForceNew, LTR/STR sub-service blocks)
//   - .../mssql_managed_database_resource.go:233-318 (Create → ManagedDatabaseProperties: createMode,
//     restorePointInTime, sourceDatabaseId/restorableDroppedDatabaseId)
//   - .../validate/mssql.go:22-25 (ValidateMsSqlManagedInstanceDatabaseName regex)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/manageddatabases:
//     model_manageddatabase.go, model_manageddatabaseproperties.go:12-35
//
// Intentionally not encoded (documented, not emitted):
//   - long_term_retention_policy and short_term_retention_days are SEPARATE ARM
//     sub-service resources — Microsoft.Sql/managedInstances/databases/backupLongTermRetentionPolicies
//     and .../backupShortTermRetentionPolicies (SDK packages
//     managedinstancelongtermretentionpolicies / managedbackupshorttermretentionpolicies).
//     Their rules (ISO8601 durations, week_of_year 0..52, retention 1..35) belong
//     in those sub-service knowledge files, NOT in the parent database body.
//   - point_in_time_restore.source_database_id maps to EITHER
//     properties.sourceDatabaseId or properties.restorableDroppedDatabaseId
//     depending on which id parser matches (a one-field→two-path branch), so no
//     single StringRule/ForceNew path fully models it; both target paths are
//     listed as ForceNew below since both are create-only PITR inputs.
type ManagedDatabase struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ManagedDatabase)(nil)

// NewManagedDatabase returns knowledge for Microsoft.Sql/managedInstances/databases.
func NewManagedDatabase() *ManagedDatabase {
	return &ManagedDatabase{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/managedInstances/databases",
			ApiVersions:  []string{"2023-08-01-preview"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// point_in_time_restore block + its members are ForceNew:true. name and
			// managed_instance_id (parent) are envelope-owned RequiresReplace.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.createMode"},
				{PropertyPath: "properties.restorePointInTime"},
				{PropertyPath: "properties.sourceDatabaseId"},
				{PropertyPath: "properties.restorableDroppedDatabaseId"},
			},
			StringRules: []azwise.StringRule{
				// name (PropertyPath == "") — ValidateMsSqlManagedInstanceDatabaseName:
				// can't end with '.'/' ', can't contain <>*%&:\/? or control chars, ≤128 chars.
				{
					PropertyPath: "",
					Regex:        `^[^<>*%&:\\\/?]{0,127}[^\s.<>*%&:\\\/?]$`,
					Message:      `must not end with '.' or ' ', must not contain '<>*%&:\/?' or control characters, and be at most 128 characters`,
				},
				// restore_point_in_time → properties.restorePointInTime (IsRFC3339Time).
				{
					PropertyPath: "properties.restorePointInTime",
					Regex:        `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`,
					Message:      "restore_point_in_time must be a valid RFC3339 timestamp",
				},
			},
			IntRules:        []azwise.IntRule{},
			FloatRules:      []azwise.FloatRule{},
			ArrayRules:      []azwise.ArrayRule{},
			SensitiveFields: []string{},
			ComputedFields:  []string{},
			DefaultValues:   []azwise.DefaultValue{},
			RequiredFields:  []string{},
		},
	}
}

func init() { azwise.Register(NewManagedDatabase()) }
