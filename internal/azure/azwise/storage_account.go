package azwise

import (
	"strings"
	"time"
)

// StorageAccount provides resource knowledge for Microsoft.Storage/storageAccounts.
// It overrides CheckForceNew to implement conditional SKU zone-migration logic.
type StorageAccount struct {
	BaseKnowledge
}

var _ ResourceKnowledge = (*StorageAccount)(nil)

// CheckForceNew extends BaseKnowledge — checks standard ForceNew rules first, then adds
// conditional SKU zone-migration logic (cross-zone changes require recreation).
func (s *StorageAccount) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
	if s.BaseKnowledge.CheckForceNew(oldBody, newBody) {
		return true
	}
	oldSku := strings.ToUpper(extractStringValue(oldBody, "sku.name"))
	newSku := strings.ToUpper(extractStringValue(newBody, "sku.name"))
	if oldSku == "" || newSku == "" || oldSku == newSku {
		return false
	}
	zonal := map[string]bool{
		"STANDARD_ZRS":    true,
		"STANDARD_GZRS":   true,
		"STANDARD_RAGZRS": true,
	}
	nonZonal := map[string]bool{
		"STANDARD_LRS":   true,
		"STANDARD_GRS":   true,
		"STANDARD_RAGRS": true,
	}
	return (zonal[oldSku] && nonZonal[newSku]) || (nonZonal[oldSku] && zonal[newSku])
}

// NewStorageAccount returns a StorageAccount knowledge instance.
func NewStorageAccount() *StorageAccount {
	return &StorageAccount{
		BaseKnowledge: BaseKnowledge{
			ResourceType: "Microsoft.Storage/storageAccounts",
			ApiVersions:  []string{"2025-08-01"},
			ForceNew: []ForceNewRule{
				{PropertyPath: "sku.tier"},
				{PropertyPath: "properties.isHnsEnabled"},
				{PropertyPath: "properties.isNfsV3Enabled"},
				{PropertyPath: "properties.encryption.requireInfrastructureEncryption"},
				{PropertyPath: "properties.dnsEndpointType"},
				{PropertyPath: "properties.encryption.services.queue.keyType"},
				{PropertyPath: "properties.encryption.services.table.keyType"},
				{PropertyPath: "properties.immutableStorageWithVersioning"},
				{PropertyPath: "extendedLocation"},
			},
			SoftDelete: true,
			// sku.name combines account_tier + account_replication_type (e.g. "Standard_LRS").
			// kind is Required by ARM API; AzureRM defaults it to "StorageV2".
			RequiredFields: []string{
				"sku.name",
				"kind",
			},
			TimeoutsConfig: &Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			StringRules: []StringRule{
				{
					// Resource name: lowercase alphanumeric only, 3-24 chars
					Regex:     `^[a-z0-9]{3,24}$`,
					MinLength: 3,
					MaxLength: 24,
					Message:   "must be lowercase alphanumeric, 3-24 characters",
				},
				{
					PropertyPath: "sku.name",
					AllowedValues: []string{
						"Premium_LRS", "Premium_ZRS",
						"PremiumV2_LRS", "PremiumV2_ZRS",
						"Standard_GRS", "Standard_GZRS",
						"Standard_LRS", "Standard_RAGRS", "Standard_RAGZRS", "Standard_ZRS",
						"StandardV2_GRS", "StandardV2_GZRS",
						"StandardV2_LRS", "StandardV2_ZRS",
					},
					Message: "must be a valid storage SKU (e.g. Standard_LRS, Premium_ZRS)",
				},
				{
					PropertyPath: "kind",
					AllowedValues: []string{
						"BlobStorage", "BlockBlobStorage", "FileStorage",
						"Storage", "StorageV2",
					},
					Message: "must be a valid storage kind (e.g. StorageV2, BlobStorage)",
				},
				{
					PropertyPath:  "properties.accessTier",
					AllowedValues: []string{"Hot", "Cool", "Cold", "Premium", "Smart"},
					Message:       "must be Hot, Cool, Cold, Premium, or Smart",
				},
				{
					PropertyPath:  "properties.minimumTlsVersion",
					AllowedValues: []string{"TLS1_0", "TLS1_1", "TLS1_2", "TLS1_3"},
					Message:       "must be TLS1_0, TLS1_1, TLS1_2, or TLS1_3",
				},
				{
					PropertyPath:  "properties.dnsEndpointType",
					AllowedValues: []string{"Standard", "AzureDnsZone"},
					Message:       "must be Standard or AzureDnsZone",
				},
				{
					PropertyPath:  "properties.allowedCopyScope",
					AllowedValues: []string{"AAD", "All", "PrivateLink"},
					Message:       "must be AAD, All, or PrivateLink",
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled", "SecuredByPerimeter"},
					Message:       "must be Enabled, Disabled, or SecuredByPerimeter",
				},
				// ── networkAcls ──
				{
					PropertyPath:  "properties.networkAcls.defaultAction",
					AllowedValues: []string{"Allow", "Deny"},
					Message:       "must be Allow or Deny",
				},
				// ── routingPreference ──
				{
					PropertyPath:  "properties.routingPreference.routingChoice",
					AllowedValues: []string{"MicrosoftRouting", "InternetRouting"},
					Message:       "must be MicrosoftRouting or InternetRouting",
				},
				// ── encryption ──
				{
					PropertyPath:  "properties.encryption.services.queue.keyType",
					AllowedValues: []string{"Account", "Service"},
					Message:       "must be Account or Service",
				},
				{
					PropertyPath:  "properties.encryption.services.table.keyType",
					AllowedValues: []string{"Account", "Service"},
					Message:       "must be Account or Service",
				},
				// ── sasPolicy ──
				{
					PropertyPath:  "properties.sasPolicy.expirationAction",
					AllowedValues: []string{"Log", "Block"},
					Message:       "must be Log or Block",
				},
				// ── immutableStorageWithVersioning ──
				{
					PropertyPath:  "properties.immutableStorageWithVersioning.immutabilityPolicy.state",
					AllowedValues: []string{"Disabled", "Unlocked", "Locked"},
					Message:       "must be Disabled, Unlocked, or Locked",
				},
				// ── azureFilesIdentityBasedAuthentication ──
				{
					PropertyPath:  "properties.azureFilesIdentityBasedAuthentication.directoryServiceOptions",
					AllowedValues: []string{"None", "AADDS", "AD", "AADKERB"},
					Message:       "must be None, AADDS, AD, or AADKERB",
				},
			},
			// Keys and connection strings come from the ListKeys API, not the account body.
			// They are sensitive Terraform outputs but not ARM body properties.
			SensitiveFields: []string{
				"properties.primaryAccessKey",
				"properties.secondaryAccessKey",
				"properties.primaryConnectionString",
				"properties.secondaryConnectionString",
				"properties.primaryBlobConnectionString",
				"properties.secondaryBlobConnectionString",
			},
			// Fields explicitly marked Computed-only in AzureRM (not in the ARM Create model).
			// Only include fields that are truly read-only — NOT fields that are simply
			// unsupported by AzureRM but settable via ARM API.
			ComputedFields: []string{
				"properties.primaryEndpoints",
				"properties.secondaryEndpoints",
				"properties.primaryLocation",
				"properties.secondaryLocation",
				"properties.provisioningState",
				"properties.creationTime",
				"properties.statusOfPrimary",
				"properties.statusOfSecondary",
				"properties.geoReplicationStats",
				"properties.privateEndpointConnections",
				"properties.blobRestoreStatus",
				"properties.failoverInProgress",
				"properties.accountMigrationInProgress",
				"properties.isSkuConversionBlocked",
				"properties.keyCreationTime",
				"properties.lastGeoFailoverTime",
				"properties.storageAccountSkuConversionStatus",
			},
			DefaultValues: []DefaultValue{
				// Top-level body defaults from AzureRM schema
				{PropertyPath: "kind", Value: "StorageV2"},
				{PropertyPath: "properties.accessTier", Value: "Hot"},
				{PropertyPath: "properties.minimumTlsVersion", Value: "TLS1_2"},
				{PropertyPath: "properties.supportsHttpsTrafficOnly", Value: true},
				{PropertyPath: "properties.allowBlobPublicAccess", Value: false},
				{PropertyPath: "properties.allowSharedKeyAccess", Value: true},
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				{PropertyPath: "properties.allowCrossTenantReplication", Value: false},
				{PropertyPath: "properties.defaultToOAuthAuthentication", Value: false},
				{PropertyPath: "properties.encryption.requireInfrastructureEncryption", Value: false},
				{PropertyPath: "properties.dnsEndpointType", Value: "Standard"},
				{PropertyPath: "properties.encryption.services.queue.keyType", Value: "Service"},
				{PropertyPath: "properties.encryption.services.table.keyType", Value: "Service"},
				// networkAcls expand defaults when block is absent
				{PropertyPath: "properties.networkAcls.defaultAction", Value: "Allow"},
				{PropertyPath: "properties.networkAcls.bypass", Value: "AzureServices"},
				// Routing block defaults
				{PropertyPath: "properties.routingPreference.routingChoice", Value: "MicrosoftRouting"},
				{PropertyPath: "properties.routingPreference.publishInternetEndpoints", Value: false},
				{PropertyPath: "properties.routingPreference.publishMicrosoftEndpoints", Value: false},
			},
		},
	}
}
