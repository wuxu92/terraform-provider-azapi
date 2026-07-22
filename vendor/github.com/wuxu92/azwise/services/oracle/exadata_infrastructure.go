// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oracle

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ExadataInfrastructure provides resource knowledge for
// Oracle.Database/cloudExadataInfrastructures.
//
// Contributing TF resource:
//   - azurerm_oracle_exadata_infrastructure (oracle_exadata_infrastructure_resource.go)
//
// Sources:
//   - AzureRM oracle_exadata_infrastructure_resource.go (schema + Create mapping)
//   - AzureRM validate/exadata_infrastructure.go (compute/storage count, maintenance window)
//   - Azure SDK oracledatabase/2025-09-01/cloudexadatainfrastructures
//     model_cloudexadatainfrastructureproperties.go + constants.go
type ExadataInfrastructure struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ExadataInfrastructure)(nil)

// NewExadataInfrastructure returns an ExadataInfrastructure knowledge instance.
func NewExadataInfrastructure() *ExadataInfrastructure {
	return &ExadataInfrastructure{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Oracle.Database/cloudExadataInfrastructures",
			ApiVersions:  []string{"2025-09-01"},
			// Every non-tag property (including zones and maintenance_window) is ForceNew.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "zones"},
				{PropertyPath: "properties.computeCount"},
				{PropertyPath: "properties.databaseServerType"},
				{PropertyPath: "properties.displayName"},
				{PropertyPath: "properties.shape"},
				{PropertyPath: "properties.storageCount"},
				{PropertyPath: "properties.storageServerType"},
				{PropertyPath: "properties.customerContacts"},
				{PropertyPath: "properties.maintenanceWindow"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 120 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 60 * time.Minute,
			},
			// Non-pointer required fields in CloudExadataInfrastructureProperties, plus
			// compute_count / storage_count (Required in schema).
			RequiredFields: []string{
				"properties.computeCount",
				"properties.displayName",
				"properties.shape",
				"properties.storageCount",
			},
			StringRules: []azwise.StringRule{
				// ── Resource name (validate.ExadataName) ── must not be empty.
				{
					MinLength: 1,
					Message:   "must not be an empty string",
				},
				// ── properties.displayName (validate.CloudVMClusterName) ──
				// 1-255 chars, start with letter/underscore. The "no consecutive
				// hyphens" check needs negative lookahead (unsupported by RE2), so
				// only the start-char rule is encoded.
				{
					PropertyPath: "properties.displayName",
					Regex:        `^[a-zA-Z_]`,
					MinLength:    1,
					MaxLength:    255,
					Message:      "must be 1-255 characters and start with a letter or underscore",
				},
				// ── properties.databaseServerType (StringLenBetween 1-255) ──
				{
					PropertyPath: "properties.databaseServerType",
					MinLength:    1,
					MaxLength:    255,
				},
				// ── properties.storageServerType (StringLenBetween 1-255) ──
				{
					PropertyPath: "properties.storageServerType",
					MinLength:    1,
					MaxLength:    255,
				},
				// ── properties.maintenanceWindow.patchingMode (PatchingMode) ──
				{
					PropertyPath:  "properties.maintenanceWindow.patchingMode",
					AllowedValues: []string{"NonRolling", "Rolling"},
				},
				// ── properties.maintenanceWindow.preference (Preference) ──
				{
					PropertyPath:  "properties.maintenanceWindow.preference",
					AllowedValues: []string{"CustomPreference", "NoPreference"},
				},
			},
			IntRules: []azwise.IntRule{
				// compute_count — validate.ComputeCount: 2..32
				{PropertyPath: "properties.computeCount", MinValue: azwise.Ptr(int64(2)), MaxValue: azwise.Ptr(int64(32))},
				// storage_count — validate.StorageCount: 3..64
				{PropertyPath: "properties.storageCount", MinValue: azwise.Ptr(int64(3)), MaxValue: azwise.Ptr(int64(64))},
				// maintenance_window.lead_time_in_weeks — validate.LeadTimeInWeeks: 1..4
				{PropertyPath: "properties.maintenanceWindow.leadTimeInWeeks", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(4))},
				// NOTE: maintenance_window days_of_week / months / hours_of_day /
				// weeks_of_month are array-element enum/range validators; azwise skips
				// array-element paths, so those are not represented here.
			},
			// Server-computed read-only properties (SDK CloudExadataInfrastructureProperties).
			ComputedFields: []string{
				"properties.activatedStorageCount",
				"properties.additionalStorageCount",
				"properties.availableStorageSizeInGbs",
				"properties.computeModel",
				"properties.cpuCount",
				"properties.dataStorageSizeInTbs",
				"properties.dbNodeStorageSizeInGbs",
				"properties.dbServerVersion",
				"properties.definedFileSystemConfiguration",
				"properties.estimatedPatchingTime",
				"properties.lastMaintenanceRunId",
				"properties.lifecycleDetails",
				"properties.lifecycleState",
				"properties.maxCpuCount",
				"properties.maxDataStorageInTbs",
				"properties.maxDbNodeStorageSizeInGbs",
				"properties.maxMemoryInGbs",
				"properties.memorySizeInGbs",
				"properties.monthlyDbServerVersion",
				"properties.monthlyStorageServerVersion",
				"properties.nextMaintenanceRunId",
				"properties.ociUrl",
				"properties.ocid",
				"properties.provisioningState",
				"properties.storageServerVersion",
				"properties.timeCreated",
				"properties.totalStorageSizeInGbs",
			},
		},
	}
}

func init() { azwise.Register(NewExadataInfrastructure()) }
