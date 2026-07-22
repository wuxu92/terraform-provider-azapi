package mssql

import (
	"strings"
	"time"

	"github.com/wuxu92/azwise"
)

// Database provides resource knowledge for Microsoft.Sql/servers/databases.
//
// Contributing Terraform resource: azurerm_mssql_database.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mssql/mssql_database_resource.go
//     (schema resourceMsSqlDatabaseSchema L1550-1973, Create body L336-488,
//     CustomizeDiff conditional ForceNew L77-142)
//   - terraform-provider-azurerm internal/services/mssql/validate/mssql.go
//     (ValidateMsSqlDatabaseName L22-25), database_collation.go (DatabaseCollation L13-16),
//     database_auto_pause_delay.go (DatabaseAutoPauseDelay L8-22)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/databases:
//     model_databaseproperties.go, constants.go
//
// Notes:
//   - name/server_id are envelope/parent references; not emitted as body rules.
//   - max_size_gb validates 0.1-4096 GB but maps to properties.maxSizeBytes (bytes); the
//     GB range cannot be applied to the bytes field, so no FloatRule is emitted.
//   - min_capacity uses validation.FloatInSlice (a discrete set, not a contiguous range),
//     which azwise FloatRule (min/max only) cannot represent; omitted.
//   - auto_pause_delay_in_minutes accepts 15-10080 OR the special sentinel -1; the
//     discontinuous domain cannot be expressed as an IntRule min/max, so it is omitted.
//   - transparent_data_encryption_enabled, geo_backup_enabled, long/short_term_retention_policy
//     and threat_detection_policy are written through separate SQL sub-APIs
//     (transparentDataEncryption, geoBackupPolicies, backupLongTermRetentionPolicies,
//     backupShortTermRetentionPolicies, securityAlertPolicies), not the databases body;
//     their rules belong to those resource types and are not emitted here.
//   - sku_name uses a custom regex (validate.DatabaseSkuName) over many SKU families; too
//     broad to encode reliably as a declarative rule, so no StringRule for sku.name.
type Database struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Database)(nil)

// NewDatabase returns knowledge for the databases resource.
func NewDatabase() *Database {
	return &Database{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/servers/databases",
			ApiVersions:  []string{"2023-08-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.createMode"},
				{PropertyPath: "properties.collation"},
				{PropertyPath: "properties.isLedgerOn"},
				{PropertyPath: "properties.sourceDatabaseId"},
				{PropertyPath: "properties.secondaryType"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[^<>*%&:\\\/?]{0,127}[^\s.<>*%&:\\\/?]$`,
					Message:      "database name can't end with '.' or ' ', can't contain '<,>,*,%,&,:,\\,/,?' or control characters, and can't have more than 128 characters",
				},
				{
					PropertyPath: "properties.collation",
					Regex:        `(^[A-Z]+)([A-Za-z0-9]+_)+((BIN|BIN2|CI_AI|CI_AI_KS|CI_AI_KS_WS|CI_AI_WS|CI_AS|CI_AS_KS|CI_AS_KS_WS|CS_AI|CS_AI_KS|CS_AI_KS_WS|CS_AI_WS|CS_AS|CS_AS_KS|CS_AS_KS_WS)+)((_[A-Za-z0-9]+)+$)*`,
					Message:      "collation must be a valid SQL Server collation identifier",
				},
				{
					PropertyPath:  "properties.createMode",
					AllowedValues: []string{"Copy", "Default", "OnlineSecondary", "PointInTimeRestore", "Recovery", "Restore", "RestoreExternalBackup", "RestoreExternalBackupSecondary", "RestoreLongTermRetentionBackup", "Secondary"},
					Message:       "create_mode is not a valid CreateMode value",
				},
				{
					PropertyPath:  "properties.preferredEnclaveType",
					AllowedValues: []string{"Default", "VBS"},
					Message:       "enclave_type must be Default or VBS",
				},
				{
					PropertyPath:  "properties.licenseType",
					AllowedValues: []string{"BasePrice", "LicenseIncluded"},
					Message:       "license_type must be BasePrice or LicenseIncluded",
				},
				{
					PropertyPath:  "properties.sampleName",
					AllowedValues: []string{"AdventureWorksLT", "WideWorldImportersFull", "WideWorldImportersStd"},
					Message:       "sample_name is not a valid SampleName value",
				},
				{
					PropertyPath:  "properties.requestedBackupStorageRedundancy",
					AllowedValues: []string{"Geo", "GeoZone", "Local", "Zone"},
					Message:       "storage_account_type must be Geo, GeoZone, Local or Zone",
				},
				{
					PropertyPath:  "properties.secondaryType",
					AllowedValues: []string{"Geo", "Named", "Standby"},
					Message:       "secondary_type must be Geo, Named or Standby",
				},
				{
					PropertyPath:  "properties.readScale",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "read_scale must be Enabled or Disabled",
				},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.highAvailabilityReplicaCount", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(4))},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.createMode", Value: "Default"},
				{PropertyPath: "properties.requestedBackupStorageRedundancy", Value: "Geo"},
				{PropertyPath: "properties.encryptionProtectorAutoRotation", Value: false},
			},
			ComputedFields: []string{
				"properties.databaseId",
				"properties.status",
				"properties.creationDate",
				"properties.currentServiceObjectiveName",
				"properties.defaultSecondaryLocation",
				"properties.earliestRestoreDate",
				"properties.failoverGroupId",
				"properties.currentBackupStorageRedundancy",
				"properties.currentSku",
			},
		},
	}
}

// CheckForceNew extends the declarative ForceNew rules with the resource's conditional
// CustomizeDiff logic (mssql_database_resource.go L77-92):
//   - changing sku.name away from a Hyperscale ("HS") SKU forces replacement;
//   - removing enclave_type (properties.preferredEnclaveType) once set forces replacement.
func (s *Database) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
	if s.BaseKnowledge.CheckForceNew(oldBody, newBody) {
		return true
	}

	oldSku := azwise.ExtractStringValue(oldBody, "sku.name")
	newSku := azwise.ExtractStringValue(newBody, "sku.name")
	if strings.HasPrefix(oldSku, "HS") && newSku != "" && !strings.HasPrefix(newSku, "HS") {
		return true
	}

	oldEnclave := azwise.ExtractStringValue(oldBody, "properties.preferredEnclaveType")
	newEnclave := azwise.ExtractStringValue(newBody, "properties.preferredEnclaveType")
	if oldEnclave != "" && newEnclave == "" {
		return true
	}

	return false
}

func init() { azwise.Register(NewDatabase()) }
