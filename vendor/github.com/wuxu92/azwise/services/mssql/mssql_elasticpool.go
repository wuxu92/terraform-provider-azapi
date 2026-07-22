package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MsSqlElasticPool provides resource knowledge for Microsoft.Sql/servers/elasticPools
// (azurerm_mssql_elasticpool).
//
// Sources:
//   - AzureRM internal/services/mssql/mssql_elasticpool_resource.go
//     :48-53   (timeouts: create/update/delete 30m, read 5m)
//     :56-72   (name ForceNew; server_name ForceNew — envelope)
//     :74-133  (sku Required: name/tier/capacity Required StringInSlice/IntAtLeast; family Optional)
//     :135-140 (maintenance_configuration_name default "SQL_Default")
//     :142-161 (per_database_settings Required: min_capacity/max_capacity FloatAtLeast(0))
//     :163-213 (max_size_bytes IntAtLeast(0); enclave_type StringInSlice(Default/VBS);
//     license_type StringInSlice(BasePrice/LicenseIncluded); high_availability_replica_count IntBetween(0,4))
//     :238-246 (enclave_type conditional ForceNew: only when removing a previously-set value)
//     :288-322 (create mapping: sku.* / properties.licenseType / perDatabaseSettings /
//     zoneRedundant / maintenanceConfigurationId / preferredEnclaveType / maxSizeBytes /
//     highAvailabilityReplicaCount)
//   - AzureRM internal/services/mssql/validate/mssql.go:41-47 (elastic pool name regex)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/elasticpools:
//     model_elasticpool.go (sku top-level), model_elasticpoolproperties.go,
//     model_sku.go, model_elasticpoolperdatabasesettings.go, constants.go
//     (ElasticPoolLicenseType, AlwaysEncryptedEnclaveType)
type MsSqlElasticPool struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MsSqlElasticPool)(nil)

// NewMsSqlElasticPool returns knowledge for the servers/elasticPools resource.
func NewMsSqlElasticPool() *MsSqlElasticPool {
	return &MsSqlElasticPool{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/servers/elasticPools",
			ApiVersions:  []string{"2023-08-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				// enclave_type is ForceNew only when transitioning from a set value to
				// empty (ForceNewIfChange); it can otherwise be changed in place.
				{PropertyPath: "properties.preferredEnclaveType"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[^&%\\\/?]{0,127}[^\s.&%\\\/?]$`,
					Message:      "name can't end with '.' or ' ', can't contain '%,&,\\,/,?' or control characters, and can't be longer than 128 characters",
				},
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"BasicPool", "StandardPool", "PremiumPool", "GP_Gen4", "GP_Gen5", "GP_Fsv2", "GP_DC", "BC_Gen4", "BC_Gen5", "BC_DC", "HS_Gen5", "HS_PRMS", "HS_MOPRMS"},
					Message:       "sku.name must be a supported elastic pool SKU name",
				},
				{
					PropertyPath:  "sku.tier",
					AllowedValues: []string{"Basic", "Standard", "Premium", "GeneralPurpose", "BusinessCritical", "Hyperscale"},
					Message:       "sku.tier must be one of Basic, Standard, Premium, GeneralPurpose, BusinessCritical or Hyperscale",
				},
				{
					PropertyPath:  "sku.family",
					AllowedValues: []string{"Gen4", "Gen5", "Fsv2", "DC", "MOPRMS", "PRMS"},
					Message:       "sku.family must be one of Gen4, Gen5, Fsv2, DC, MOPRMS or PRMS",
				},
				{
					PropertyPath:  "properties.licenseType",
					AllowedValues: []string{"BasePrice", "LicenseIncluded"},
					Message:       "license_type must be one of BasePrice or LicenseIncluded",
				},
				{
					PropertyPath:  "properties.preferredEnclaveType",
					AllowedValues: []string{"Default", "VBS"},
					Message:       "enclave_type must be one of Default or VBS",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "sku.capacity",
					MinValue:     azwise.Ptr(int64(0)),
					Message:      "sku.capacity must be at least 0",
				},
				{
					PropertyPath: "properties.maxSizeBytes",
					MinValue:     azwise.Ptr(int64(0)),
					Message:      "max_size_bytes must be at least 0",
				},
				{
					PropertyPath: "properties.highAvailabilityReplicaCount",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(4)),
					Message:      "high_availability_replica_count must be between 0 and 4",
				},
			},
			FloatRules: []azwise.FloatRule{
				{
					PropertyPath: "properties.perDatabaseSettings.minCapacity",
					MinValue:     azwise.Ptr(0.0),
					Message:      "per_database_settings.min_capacity must be at least 0",
				},
				{
					PropertyPath: "properties.perDatabaseSettings.maxCapacity",
					MinValue:     azwise.Ptr(0.0),
					Message:      "per_database_settings.max_capacity must be at least 0",
				},
			},
			RequiredFields: []string{
				"sku.name",
				"sku.tier",
				"sku.capacity",
				"properties.perDatabaseSettings.minCapacity",
				"properties.perDatabaseSettings.maxCapacity",
			},
		},
	}
}

func init() { azwise.Register(NewMsSqlElasticPool()) }
