// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oracle

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ExascaleDbStorageVault provides resource knowledge for
// Oracle.Database/exascaleDbStorageVaults.
//
// Contributing TF resource:
//   - azurerm_oracle_exascale_database_storage_vault
//     (oracle_exascale_database_storage_vault_resource.go)
//
// Sources:
//   - AzureRM oracle_exascale_database_storage_vault_resource.go (schema + Create mapping)
//   - AzureRM validate/exascale_database_storage_vault.go (name regex)
//   - Azure SDK oracledatabase/2025-09-01/exascaledbstoragevaults
//     model_exascaledbstoragevaultproperties.go
type ExascaleDbStorageVault struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ExascaleDbStorageVault)(nil)

// NewExascaleDbStorageVault returns an ExascaleDbStorageVault knowledge instance.
func NewExascaleDbStorageVault() *ExascaleDbStorageVault {
	return &ExascaleDbStorageVault{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Oracle.Database/exascaleDbStorageVaults",
			ApiVersions:  []string{"2025-09-01"},
			// Every non-tag property is ForceNew.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "zones"},
				{PropertyPath: "properties.additionalFlashCacheInPercent"},
				{PropertyPath: "properties.description"},
				{PropertyPath: "properties.displayName"},
				{PropertyPath: "properties.highCapacityDatabaseStorageInput"},
				{PropertyPath: "properties.timeZone"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 120 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// Non-pointer required fields in the SDK properties model.
			RequiredFields: []string{
				"properties.displayName",
				"properties.highCapacityDatabaseStorageInput.totalSizeInGbs",
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validation.All(StringLenBetween(1,255),
				// validate.ExascaleDatabaseStorageVaultName): start with letter or
				// underscore, letters/numbers/underscores/hyphens, no consecutive "--".
				{
					Regex:     `^[a-zA-Z_](?:[a-zA-Z0-9_]*(?:-[a-zA-Z0-9_]+)*-?)?$`,
					MinLength: 1,
					MaxLength: 255,
					Message:   "must be 1-255 characters, begin with a letter or underscore, and contain only letters, numbers, underscores and non-consecutive hyphens",
				},
				// ── properties.displayName ── same validator as name.
				{
					PropertyPath: "properties.displayName",
					Regex:        `^[a-zA-Z_](?:[a-zA-Z0-9_]*(?:-[a-zA-Z0-9_]+)*-?)?$`,
					MinLength:    1,
					MaxLength:    255,
					Message:      "must be 1-255 characters, begin with a letter or underscore, and contain only letters, numbers, underscores and non-consecutive hyphens",
				},
			},
			IntRules: []azwise.IntRule{
				// additional_flash_cache_percentage — IntBetween(0, 100)
				{PropertyPath: "properties.additionalFlashCacheInPercent", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(100))},
			},
			DefaultValues: []azwise.DefaultValue{
				// time_zone Default "UTC"
				{PropertyPath: "properties.timeZone", Value: "UTC"},
			},
			// Server-computed read-only properties (SDK ExascaleDbStorageVaultProperties).
			ComputedFields: []string{
				"properties.attachedShapeAttributes",
				"properties.exadataInfrastructureId",
				"properties.highCapacityDatabaseStorage",
				"properties.lifecycleDetails",
				"properties.lifecycleState",
				"properties.ociUrl",
				"properties.ocid",
				"properties.provisioningState",
				"properties.vmClusterCount",
			},
		},
	}
}

func init() { azwise.Register(NewExascaleDbStorageVault()) }
