// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package mssqlmanagedinstance

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ManagedInstance provides resource knowledge for Microsoft.Sql/managedInstances
// (Terraform azurerm_mssql_managed_instance).
//
// The ARM body is {location, sku, identity, tags, properties:{...}} against the
// go-azure-sdk managedinstances.ManagedInstance model. Envelope fields (name,
// location, resource_group_name) are handled by the generated native schema and
// are not repeated as body rules; name/location/resource_group are ForceNew via
// commonschema.
//
// `storage_account_type` (GRS/LRS/ZRS/GZRS) is NOT sent verbatim: AzureRM maps it
// 1:1 to properties.requestedBackupStorageRedundancy with GRS→Geo, LRS→Local,
// ZRS→Zone, GZRS→GeoZone. So the enum + default below use the ARM values
// (Geo/GeoZone/Local/Zone), which is what AzAPI transmits.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mssqlmanagedinstance/mssql_managed_instance_resource.go:100-376
//     (Arguments: validators, defaults, ForceNew flags)
//   - .../mssql_managed_instance_resource.go:391-453 (CustomizeDiff: conditional ForceNew + cross-field rules)
//   - .../mssql_managed_instance_resource.go:455-551 (Create → ManagedInstanceProperties mapping)
//   - .../mssql_managed_instance_resource.go:921-957 (expandSkuName → sku.name/tier/family)
//   - .../mssql_managed_instance_resource.go:959-986 (storage_account_type ↔ requestedBackupStorageRedundancy)
//   - .../mssql_managed_instance_resource.go:1000-1018 (azure_active_directory_administrator → properties.administrators)
//   - .../validate/mssql.go:13-16 (ValidateMsSqlManagedInstanceServerName regex)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/managedinstances:
//     model_managedinstance.go:10-19, model_managedinstanceproperties.go:12-54,
//     model_sku.go:6-12, model_managedinstanceexternaladministrator.go:6-13,
//     constants.go (BackupStorageRedundancy, HybridSecondaryUsage,
//     ManagedInstanceDatabaseFormat, ManagedInstanceLicenseType,
//     ManagedInstanceProxyOverride, ServicePrincipalType, PrincipalType)
//
// Intentionally not encoded (documented, not emitted):
//   - Conditional ForceNew (CustomizeDiff): properties.dnsZonePartner (only ""→set),
//     identity (only 1→0 removal), properties.databaseFormat (only AlwaysUpToDate→
//     SQLServer2022 downgrade). A declarative ForceNew fires on ANY change to the
//     path and would wrongly force replacement on the allowed transitions
//     (databaseFormat SQLServer2022→AlwaysUpToDate upgrade; dnsZonePartner A→B),
//     so these belong in a customizer, not a BaseKnowledge ForceNewRule.
//   - Cross-field CustomizeDiff invariants (storage_iops requires
//     general_purpose_v2_enabled and non-BC sku; zone_redundant_enabled XOR
//     general_purpose_v2_enabled; aad admin requires login/password unless
//     authOnly): these are conditional relations over combinations of body flags,
//     not expressible as azwise RelationalRule (which is set/unset presence), so
//     they belong in the resource customizer.
//   - administrator_login / administrator_login_password AtLeastOneOf
//     azure_active_directory_administrator, and administrator_login RequiredWith
//     administrator_login_password: these span a body scalar
//     (properties.administratorLogin/Password) and a body sub-object
//     (properties.administrators); RequiredWith/AtLeastOneOf over an object vs its
//     absence is left to the customizer.
//   - vCores IntInSlice{4,6,8,10,12,16,20,24,32,40,48,56,64,80,96,128}: a discrete
//     allow-list, not a contiguous range; the IntRule below applies the loose
//     4..128 envelope (catches egregious values) — the exact set is a customizer
//     concern.
//   - read-only GET fields (properties.dnsZone, fullyQualifiedDomainName,
//     provisioningState, state, createTime, virtualClusterId, ...) share the PUT
//     struct but are bicep-ReadOnly; the native generator already marks them
//     Computed and strips them, so no ComputedFields entry is needed.
//   - maintenance_configuration_name default "SQL_Default" expands to a
//     subscription-scoped Microsoft.Maintenance/publicMaintenanceConfigurations
//     resource ID (properties.maintenanceConfigurationId); the concrete value is
//     subscription-dependent, so no static DefaultValue is emitted.
type ManagedInstance struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ManagedInstance)(nil)

// NewManagedInstance returns knowledge for Microsoft.Sql/managedInstances.
func NewManagedInstance() *ManagedInstance {
	return &ManagedInstance{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/managedInstances",
			ApiVersions:  []string{"2023-08-01-preview"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 24 * time.Hour,
				Read:   5 * time.Minute,
				Update: 24 * time.Hour,
				Delete: 24 * time.Hour,
			},
			// Unconditional body ForceNew (AzureRM schema ForceNew:true). name,
			// location, resource_group are envelope-owned and handled by the
			// generated schema, so only body properties are listed here.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.collation"},
				{PropertyPath: "properties.administratorLogin"},
				{PropertyPath: "properties.timezoneId"},
			},
			StringRules: []azwise.StringRule{
				// name (PropertyPath == "") — ValidateMsSqlManagedInstanceServerName:
				// lowercase letters, numbers and '-', not leading/trailing '-', ≤63 chars.
				{
					PropertyPath: "",
					Regex:        `^[0-9a-z]([-0-9a-z]{0,61}[0-9a-z])?$`,
					Message:      "must contain only lowercase letters, numbers and '-', can't start or end with '-', and be at most 63 characters",
				},
				// sku_name → sku.name (StringInSlice).
				{
					PropertyPath: "sku.name",
					AllowedValues: []string{
						"BC_Gen4", "BC_Gen5", "BC_Gen8IH", "BC_Gen8IM",
						"GP_Gen4", "GP_Gen5", "GP_Gen8IH", "GP_Gen8IM",
					},
					Message: "must be a valid SQL Managed Instance SKU (e.g. GP_Gen5, BC_Gen5)",
				},
				// license_type → properties.licenseType (full SDK enum set).
				{
					PropertyPath:  "properties.licenseType",
					AllowedValues: []string{"BasePrice", "LicenseIncluded"},
					Message:       "must be one of BasePrice or LicenseIncluded",
				},
				// database_format → properties.databaseFormat.
				{
					PropertyPath:  "properties.databaseFormat",
					AllowedValues: []string{"AlwaysUpToDate", "SQLServer2022"},
					Message:       "must be one of AlwaysUpToDate or SQLServer2022",
				},
				// hybrid_secondary_usage → properties.hybridSecondaryUsage.
				{
					PropertyPath:  "properties.hybridSecondaryUsage",
					AllowedValues: []string{"Active", "Passive"},
					Message:       "must be one of Active or Passive",
				},
				// minimum_tls_version → properties.minimalTlsVersion. SDK field is a
				// free string; the ARM-valid set for MI is 1.0/1.1/1.2 (AzureRM 5.0
				// restricts config to 1.2 but the service accepts the wider set).
				{
					PropertyPath:  "properties.minimalTlsVersion",
					AllowedValues: []string{"1.0", "1.1", "1.2"},
					Message:       "must be one of 1.0, 1.1 or 1.2",
				},
				// proxy_override → properties.proxyOverride (full SDK enum set).
				{
					PropertyPath:  "properties.proxyOverride",
					AllowedValues: []string{"Default", "Proxy", "Redirect"},
					Message:       "must be one of Default, Proxy or Redirect",
				},
				// service_principal_type → properties.servicePrincipal.type (full SDK set).
				{
					PropertyPath:  "properties.servicePrincipal.type",
					AllowedValues: []string{"None", "SystemAssigned"},
					Message:       "must be one of None or SystemAssigned",
				},
				// storage_account_type → properties.requestedBackupStorageRedundancy
				// (ARM values after the GRS→Geo etc. mapping).
				{
					PropertyPath:  "properties.requestedBackupStorageRedundancy",
					AllowedValues: []string{"Geo", "GeoZone", "Local", "Zone"},
					Message:       "must be one of Geo, GeoZone, Local or Zone",
				},
				// azure_active_directory_administrator block → properties.administrators
				// (single object, MaxItems 1). login_username → login (not empty).
				{
					PropertyPath: "properties.administrators.login",
					MinLength:    1,
					Message:      "login_username must not be empty",
				},
				// object_id → properties.administrators.sid (validation.IsUUID).
				{
					PropertyPath: "properties.administrators.sid",
					Regex:        `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
					Message:      "object_id must be a valid UUID",
				},
				// tenant_id → properties.administrators.tenantId (validation.IsUUID).
				{
					PropertyPath: "properties.administrators.tenantId",
					Regex:        `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
					Message:      "tenant_id must be a valid UUID",
				},
				// principal_type → properties.administrators.principalType (full SDK set).
				{
					PropertyPath:  "properties.administrators.principalType",
					AllowedValues: []string{"Application", "Group", "User"},
					Message:       "principal_type must be one of Application, Group or User",
				},
			},
			IntRules: []azwise.IntRule{
				// storage_size_in_gb → properties.storageSizeInGB (IntBetween 32..32768).
				{PropertyPath: "properties.storageSizeInGB", MinValue: azwise.Ptr(int64(32)), MaxValue: azwise.Ptr(int64(32768))},
				// storage_iops → properties.storageIOps (IntBetween 300..80000).
				{PropertyPath: "properties.storageIOps", MinValue: azwise.Ptr(int64(300)), MaxValue: azwise.Ptr(int64(80000))},
				// vcores → properties.vCores. AzureRM uses a discrete IntInSlice; the
				// loose 4..128 envelope is applied here (see doc comment).
				{PropertyPath: "properties.vCores", MinValue: azwise.Ptr(int64(4)), MaxValue: azwise.Ptr(int64(128))},
			},
			FloatRules:      []azwise.FloatRule{},
			ArrayRules:      []azwise.ArrayRule{},
			SensitiveFields: []string{"properties.administratorLoginPassword"},
			ComputedFields:  []string{},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.collation", Value: "SQL_Latin1_General_CP1_CI_AS"},
				{PropertyPath: "properties.databaseFormat", Value: "SQLServer2022"},
				{PropertyPath: "properties.hybridSecondaryUsage", Value: "Active"},
				{PropertyPath: "properties.isGeneralPurposeV2", Value: false},
				{PropertyPath: "properties.minimalTlsVersion", Value: "1.2"},
				{PropertyPath: "properties.proxyOverride", Value: "Redirect"},
				{PropertyPath: "properties.publicDataEndpointEnabled", Value: false},
				{PropertyPath: "properties.requestedBackupStorageRedundancy", Value: "Geo"},
				{PropertyPath: "properties.timezoneId", Value: "UTC"},
				{PropertyPath: "properties.zoneRedundant", Value: false},
			},
			// AzureRM Required:true body fields (envelope name/location/resource_group excluded).
			RequiredFields: []string{
				"sku.name",
				"properties.licenseType",
				"properties.storageSizeInGB",
				"properties.subnetId",
				"properties.vCores",
			},
		},
	}
}

func init() { azwise.Register(NewManagedInstance()) }
