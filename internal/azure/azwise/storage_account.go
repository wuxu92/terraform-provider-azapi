package azwise

import (
	"strings"
	"time"
)

// StorageAccount provides resource knowledge for Microsoft.Storage/storageAccounts.
// It overrides CheckForceNew to add conditional replacement rules (SKU zone
// migration, account_kind migration, large-file-share disablement).
//
// Sources:
//   - AzureRM internal/services/storage/storage_account_resource.go (schema + CustomizeDiff + expand)
//   - go-azure-sdk .../storage/2025-08-01/storageaccounts: model_storageaccountpropertiescreateparameters.go,
//     model_encryption.go, model_encryptionidentity.go, model_storageaccountcreateparameters.go (ARM body paths),
//     constants.go (enum values)
//
// Cross-field (relational) constraints — see the RequiredWith block below for the one
// account-body, object-level relation that is representable, plus inline notes on the two
// schema relations that are deliberately NOT encoded (non-mappable / sub-service).
type StorageAccount struct {
	BaseKnowledge
}

var _ ResourceKnowledge = (*StorageAccount)(nil)

// CheckForceNew extends BaseKnowledge with the conditional replacement rules a
// static ForceNew list cannot express, mirroring AzureRM's storage account
// CustomizeDiff (storage_account_resource.go): declarative rules first, then SKU
// zone migration, account_kind migration, and large-file-share disablement.
func (s *StorageAccount) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
	if s.BaseKnowledge.CheckForceNew(oldBody, newBody) {
		return true
	}
	return skuZoneMigration(oldBody, newBody) ||
		accountKindRequiresReplace(oldBody, newBody) ||
		largeFileShareDisabled(oldBody, newBody)
}

// skuZoneMigration reports whether the account_replication_type (the suffix of
// sku.name, e.g. "ZRS" in "Standard_ZRS") crosses the zonal/non-zonal boundary,
// which requires recreation. Matches AzureRM ForceNewIfChange("account_replication_type")
// on the replication suffix alone — tier-agnostic, so Premium_LRS <-> Premium_ZRS is
// covered as well as Standard.
func skuZoneMigration(oldBody, newBody map[string]interface{}) bool {
	oldSku := strings.ToUpper(extractStringValue(oldBody, "sku.name"))
	newSku := strings.ToUpper(extractStringValue(newBody, "sku.name"))
	if oldSku == "" || newSku == "" || oldSku == newSku {
		return false
	}
	oldRep := replicationType(oldSku)
	newRep := replicationType(newSku)
	zonal := map[string]bool{"ZRS": true, "GZRS": true, "RAGZRS": true}
	nonZonal := map[string]bool{"LRS": true, "GRS": true, "RAGRS": true}
	return (zonal[oldRep] && nonZonal[newRep]) || (nonZonal[oldRep] && zonal[newRep])
}

// replicationType returns the redundancy suffix of an upper-cased sku.name, e.g.
// "STANDARD_RAGZRS" -> "RAGZRS", "PREMIUMV2_LRS" -> "LRS".
func replicationType(sku string) string {
	if i := strings.LastIndex(sku, "_"); i >= 0 {
		return sku[i+1:]
	}
	return sku
}

// accountKindRequiresReplace reports whether an account_kind change forces
// replacement. AzureRM permits exactly one in-place migration, Storage -> StorageV2;
// every other kind change requires recreation (storage_account_resource.go:1103).
func accountKindRequiresReplace(oldBody, newBody map[string]interface{}) bool {
	oldKind := extractStringValue(oldBody, "kind")
	newKind := extractStringValue(newBody, "kind")
	if oldKind == "" || oldKind == newKind {
		return false
	}
	return oldKind != "Storage" && newKind != "StorageV2"
}

// largeFileShareDisabled reports whether large file shares are being turned off.
// Once enabled the feature cannot be disabled in place; Enabled -> anything else
// forces replacement (storage_account_resource.go:1116).
func largeFileShareDisabled(oldBody, newBody map[string]interface{}) bool {
	oldLFS := extractStringValue(oldBody, "properties.largeFileSharesState")
	newLFS := extractStringValue(newBody, "properties.largeFileSharesState")
	return strings.EqualFold(oldLFS, "Enabled") && !strings.EqualFold(newLFS, "Enabled")
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
			// ── Relational (cross-property) constraints ──
			// Genuine ARM constraint: a customer-managed-key encryption identity references a
			// user-assigned identity that MUST be assigned to the account. AzureRM enforces this in
			// expandAccountCustomerManagedKey (storage_account_resource.go L2524-2526) — CMK "can only
			// be configured when the storage account uses a `UserAssigned` or `SystemAssigned,
			// UserAssigned` managed identity" — and marks customer_managed_key.user_assigned_identity_id
			// Required (storage_account_resource.go L1248-1252). Both paths are account-body and
			// object-level (verified in model_encryption.go / model_encryptionidentity.go).
			RequiredWith: []RelationalRule{
				{
					Paths: []string{
						"properties.encryption.identity.userAssignedIdentity",
						"identity.userAssignedIdentities",
					},
					Message: "customer-managed key encryption requires the account to carry the referenced user-assigned identity",
				},
			},
			// Schema relations found in storage_account_resource.go but NOT encoded as RelationalRules:
			//   - customer_managed_key {key_vault_key_id ExactlyOneOf managed_hsm_key_id}
			//     (L1236/L1244): both Terraform fields expand to the SAME ARM path
			//     properties.encryption.keyvaultproperties (keyName/keyVersion/keyVaultUri), so the
			//     ExactlyOneOf has no distinct object-level ARM representation. Skipped (non-mappable).
			//   - blob_properties {restore_policy RequiredWith delete_retention_policy} (L522):
			//     sub-service — belongs to Microsoft.Storage/storageAccounts/blobServices/default,
			//     not the account body. Encode on the blob-service knowledge file. Skipped.
		},
	}
}
