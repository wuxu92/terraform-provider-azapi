// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oracle

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutonomousDatabase provides resource knowledge for Oracle.Database/autonomousDatabases.
//
// One ARM type served by three AzureRM Terraform resources, merged here. They create
// the SAME autonomousDatabases body via different DataBaseType / Source / CloneType
// discriminators:
//   - azurerm_oracle_autonomous_database                — DataBaseType "Regular"
//     (oracle_autonomous_database_resource.go)
//   - azurerm_oracle_autonomous_database_clone_from_backup   — Source "BackupFromTimestamp",
//     DataBaseType "CloneFromBackupTimestamp"
//     (oracle_autonomous_database_clone_from_backup_resource.go)
//   - azurerm_oracle_autonomous_database_clone_from_database — Source "Database",
//     DataBaseType "Clone" (oracle_autonomous_database_clone_from_database_resource.go)
//
// Only knowledge universally true for every autonomousDatabases body is unioned. The
// discriminator fields (dataBaseType, source, cloneType, sourceId, timestamp,
// refreshableModel) are kind-specific and deliberately NOT declared as ForceNew or
// RequiredFields here. Value constraints on the optional longTermBackupSchedule
// sub-object (only the Regular kind sets it) are safe because they fire only when
// that sub-object is present.
//
// Sources:
//   - AzureRM oracle_autonomous_database_resource.go (schema + Create/Update mapping)
//   - AzureRM oracle_autonomous_database_clone_from_backup_resource.go (schema)
//   - AzureRM oracle_autonomous_database_clone_from_database_resource.go (schema)
//   - AzureRM validate/autonomous_database_regular.go (name/password/compute-model rules)
//   - Azure SDK oracledatabase/2025-09-01/autonomousdatabases models + constants.go
type AutonomousDatabase struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutonomousDatabase)(nil)

// NewAutonomousDatabase returns an AutonomousDatabase knowledge instance.
func NewAutonomousDatabase() *AutonomousDatabase {
	return &AutonomousDatabase{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Oracle.Database/autonomousDatabases",
			ApiVersions:  []string{"2025-09-01"},
			// ForceNew paths that are ForceNew in EVERY contributor. The Regular
			// resource keeps several base props mutable (admin_password,
			// compute_count, data_storage_size, backup_retention, auto_scaling*,
			// allowed_ips) that the clones force-new; those are NOT universal and
			// stay out to avoid over-forcing.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.characterSet"},
				{PropertyPath: "properties.computeModel"},
				{PropertyPath: "properties.dbVersion"},
				{PropertyPath: "properties.dbWorkload"},
				{PropertyPath: "properties.displayName"},
				{PropertyPath: "properties.licenseModel"},
				{PropertyPath: "properties.ncharacterSet"},
				{PropertyPath: "properties.customerContacts"},
				{PropertyPath: "properties.isMtlsConnectionRequired"},
				{PropertyPath: "properties.subnetId"},
				{PropertyPath: "properties.vnetId"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 120 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// Base properties required by all three contributors.
			RequiredFields: []string{
				"properties.adminPassword",
				"properties.backupRetentionPeriodInDays",
				"properties.characterSet",
				"properties.computeCount",
				"properties.computeModel",
				"properties.dataStorageSizeInTbs",
				"properties.dbVersion",
				"properties.dbWorkload",
				"properties.displayName",
				"properties.isAutoScalingEnabled",
				"properties.isAutoScalingForStorageEnabled",
				"properties.isMtlsConnectionRequired",
				"properties.licenseModel",
				"properties.ncharacterSet",
			},
			SensitiveFields: []string{
				"properties.adminPassword",
			},
			StringRules: []azwise.StringRule{
				// ── Resource name (validate.AutonomousDatabaseName) ──
				// must start with a letter, then letters/numbers only, max 30 chars.
				{
					Regex:     `^[a-zA-Z][a-zA-Z0-9]*$`,
					MaxLength: 30,
					Message:   "must start with a letter and contain only letters and numbers, 30 characters max",
				},
				// ── properties.dbWorkload (WorkloadType) ──
				{
					PropertyPath:  "properties.dbWorkload",
					AllowedValues: []string{"AJD", "APEX", "DW", "OLTP"},
				},
				// ── properties.computeModel (ComputeModel) ──
				{
					PropertyPath:  "properties.computeModel",
					AllowedValues: []string{"ECPU", "OCPU"},
				},
				// ── properties.licenseModel (LicenseModel) ──
				{
					PropertyPath:  "properties.licenseModel",
					AllowedValues: []string{"BringYourOwnLicense", "LicenseIncluded"},
				},
				// ── properties.longTermBackupSchedule.repeatCadence (RepeatCadenceType) ──
				// Only the Regular kind sets long_term_backup_schedule; the rule fires
				// only when the sub-object is present, so it is safe on the shared type.
				{
					PropertyPath:  "properties.longTermBackupSchedule.repeatCadence",
					AllowedValues: []string{"Monthly", "OneTime", "Weekly", "Yearly"},
				},
			},
			IntRules: []azwise.IntRule{
				// properties.backupRetentionPeriodInDays — IntBetween(1, 60)
				{PropertyPath: "properties.backupRetentionPeriodInDays", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(60))},
				// properties.dataStorageSizeInTbs — IntBetween(1, 384)
				{PropertyPath: "properties.dataStorageSizeInTbs", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(384))},
				// properties.longTermBackupSchedule.retentionPeriodInDays — IntBetween(90, 2558)
				{PropertyPath: "properties.longTermBackupSchedule.retentionPeriodInDays", MinValue: azwise.Ptr(int64(90)), MaxValue: azwise.Ptr(int64(2558))},
			},
			FloatRules: []azwise.FloatRule{
				// properties.computeCount — FloatBetween(2.0, 512.0)
				{PropertyPath: "properties.computeCount", MinValue: azwise.Ptr(2.0), MaxValue: azwise.Ptr(512.0)},
			},
			// Server-computed read-only properties present in the GET model but never
			// settable via Create/Update (SDK BaseAutonomousDatabaseBasePropertiesImpl).
			ComputedFields: []string{
				"properties.actualUsedDataStorageSizeInTbs",
				"properties.allocatedStorageSizeInTbs",
				"properties.apexDetails",
				"properties.autonomousDatabaseId",
				"properties.availableUpgradeVersions",
				"properties.connectionStrings",
				"properties.connectionUrls",
				"properties.cpuCoreCount",
				"properties.dataStorageSizeInGbs",
				"properties.failedDataRecoveryInSeconds",
				"properties.inMemoryAreaInGbs",
				"properties.lifecycleDetails",
				"properties.lifecycleState",
				"properties.localStandbyDb",
				"properties.memoryPerOracleComputeUnitInGbs",
				"properties.nextLongTermBackupTimeStamp",
				"properties.ociUrl",
				"properties.ocid",
				"properties.privateEndpoint",
				"properties.privateEndpointIp",
				"properties.privateEndpointLabel",
				"properties.provisionableCpus",
				"properties.provisioningState",
				"properties.serviceConsoleUrl",
				"properties.sqlWebDeveloperUrl",
				"properties.supportedRegionsToCloneTo",
				"properties.timeCreated",
				"properties.timeMaintenanceBegin",
				"properties.timeMaintenanceEnd",
				"properties.usedDataStorageSizeInGbs",
				"properties.usedDataStorageSizeInTbs",
			},
		},
	}
}

func init() { azwise.Register(NewAutonomousDatabase()) }
