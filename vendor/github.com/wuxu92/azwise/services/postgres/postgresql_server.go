package postgres

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PostgresqlServer provides resource knowledge for
// Microsoft.DBforPostgreSQL/servers (azurerm_postgresql_server).
//
// This is the deprecated PostgreSQL Single Server (retired 2025-03-28); it is a
// DISTINCT ARM type from Microsoft.DBforPostgreSQL/flexibleServers.
//
// Sources:
//   - AzureRM internal/services/postgres/postgresql_server_resource.go
//     :40-61   (skuList — the 20 allowed sku_name values)
//     :63-382  (schema: name ServerName regex, sku_name enum, version enum ForceNew,
//     administrator_login ForceNew, storage_mb IntBetween(5120,16777216) DivBy 1024,
//     backup_retention_days IntBetween(7,35), geo_redundant_backup_enabled ForceNew,
//     infrastructure_encryption_enabled ForceNew, ssl_minimal_tls_version_enforced enum,
//     ssl_enforcement_enabled Required, create_mode enum default Default)
//     :107-112 (timeouts: create/update/delete 60m, read 5m)
//     :357-380 (CustomizeDiff: conditional ForceNew — see "Not encoded")
//     :435-520 (create mapping to servers.ServerForCreate / ServerPropertiesForCreate variants)
//   - go-azure-sdk resource-manager/postgresql/2017-12-01/servers:
//     model_serverforcreate.go (properties/sku/identity/tags/location envelope),
//     model_serverpropertiesfordefaultcreate.go, model_storageprofile.go, model_sku.go,
//     constants.go (enum values), id_server.go:117 (staticServers)
//
// Not encoded (deliberate):
//   - storage_mb additionally requires divisibility by 1024 (validation.IntDivisibleBy);
//     only the 5120..16777216 range is expressible via IntRule.
//   - Conditional ForceNew from CustomizeDiff (omitted to avoid false replacement):
//     sku_name when the tier changes to/from Basic (properties are top-level sku.name),
//     create_mode when Default->Replica (properties.createMode).
//   - threat_detection_policy is a SEPARATE ARM resource
//     (Microsoft.DBforPostgreSQL/servers/securityAlertPolicies) written via a distinct
//     client; its rules do not belong in this servers knowledge file.
//   - administrator_login / administrator_login_password are only Required when
//     create_mode == Default; not universal RequiredFields.
//   - creation_source_server_id maps to properties.sourceServerId only in the
//     Restore/Replica/GeoRestore create variants (discriminated union), never Default.
type PostgresqlServer struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PostgresqlServer)(nil)

// NewPostgresqlServer returns knowledge for the servers resource.
func NewPostgresqlServer() *PostgresqlServer {
	return &PostgresqlServer{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforPostgreSQL/servers",
			ApiVersions:  []string{"2017-12-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.version"},
				{PropertyPath: "properties.administratorLogin"},
				{PropertyPath: "properties.storageProfile.geoRedundantBackup"},
				{PropertyPath: "properties.infrastructureEncryption"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			SensitiveFields: []string{
				"properties.administratorLoginPassword",
			},
			ComputedFields: []string{
				"properties.fullyQualifiedDomainName",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					MinLength:    3,
					MaxLength:    63,
					Regex:        `^[0-9a-z][-0-9a-z]{1,61}[0-9a-z]$`,
					Message:      "server name must be 3-63 characters: lowercase letters, numbers and '-', not starting or ending with '-'",
				},
				{
					PropertyPath: "sku.name",
					AllowedValues: []string{
						"B_Gen4_1", "B_Gen4_2", "B_Gen5_1", "B_Gen5_2",
						"GP_Gen4_2", "GP_Gen4_4", "GP_Gen4_8", "GP_Gen4_16", "GP_Gen4_32",
						"GP_Gen5_2", "GP_Gen5_4", "GP_Gen5_8", "GP_Gen5_16", "GP_Gen5_32", "GP_Gen5_64",
						"MO_Gen5_2", "MO_Gen5_4", "MO_Gen5_8", "MO_Gen5_16", "MO_Gen5_32",
					},
					Message: "sku_name must be a supported Single Server SKU",
				},
				{
					PropertyPath:  "properties.version",
					AllowedValues: []string{"9.5", "9.6", "10", "10.0", "10.2", "11"},
					Message:       "version must be a valid ServerVersion",
				},
				{
					PropertyPath:  "properties.createMode",
					AllowedValues: []string{"Default", "GeoRestore", "PointInTimeRestore", "Replica"},
					Message:       "create_mode must be a valid CreateMode",
				},
				{
					PropertyPath:  "properties.minimalTlsVersion",
					AllowedValues: []string{"TLSEnforcementDisabled", "TLS1_0", "TLS1_1", "TLS1_2"},
					Message:       "ssl_minimal_tls_version_enforced must be a valid MinimalTlsVersionEnum",
				},
				{
					PropertyPath:  "properties.sslEnforcement",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "ssl_enforcement must be Disabled or Enabled",
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "public_network_access must be Disabled or Enabled",
				},
				{
					PropertyPath:  "properties.infrastructureEncryption",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "infrastructure_encryption must be Disabled or Enabled",
				},
				{
					PropertyPath:  "properties.storageProfile.storageAutogrow",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "storage autogrow must be Disabled or Enabled",
				},
				{
					PropertyPath:  "properties.storageProfile.geoRedundantBackup",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "geo_redundant_backup must be Disabled or Enabled",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.storageProfile.backupRetentionDays",
					MinValue:     azwise.Ptr(int64(7)),
					MaxValue:     azwise.Ptr(int64(35)),
					Message:      "backup_retention_days must be between 7 and 35",
				},
				{
					PropertyPath: "properties.storageProfile.storageMB",
					MinValue:     azwise.Ptr(int64(5120)),
					MaxValue:     azwise.Ptr(int64(16777216)),
					Message:      "storage_mb must be between 5120 and 16777216 (and divisible by 1024)",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.createMode", Value: "Default"},
				{PropertyPath: "properties.minimalTlsVersion", Value: "TLS1_2"},
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				{PropertyPath: "properties.storageProfile.storageAutogrow", Value: "Enabled"},
				{PropertyPath: "properties.storageProfile.geoRedundantBackup", Value: "Disabled"},
			},
			RequiredFields: []string{
				"sku.name",
				"properties.version",
				"properties.sslEnforcement",
			},
		},
	}
}

func init() { azwise.Register(NewPostgresqlServer()) }
