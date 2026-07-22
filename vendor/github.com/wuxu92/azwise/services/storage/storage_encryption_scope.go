package storage

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StorageEncryptionScope provides resource knowledge for
// Microsoft.Storage/storageAccounts/encryptionScopes.
//
// In AzureRM this is the azurerm_storage_encryption_scope resource. ARM models
// it as a sub-resource of the storage account (encryptionScopes/{name}), so it
// gets its own knowledge file rather than being folded into storageAccounts.
//
// Sources:
//   - terraform-provider-azurerm internal/services/storage/storage_encryption_scope_resource.go:47-83
//     (schema: name/source/key_vault_key_id/infrastructure_encryption_required)
//   - .../storage_encryption_scope_resource.go:118-129 (create payload → ARM body)
//   - .../internal/services/storage/validate/storage_encryption_scope_name.go (name regex)
//   - go-azure-sdk resource-manager/storage/2025-08-01/encryptionscopes:
//     model_encryptionscopeproperties.go (ARM json paths), constants.go
//     (EncryptionScopeSource values), id_encryptionscope.go (type-segment casing)
type StorageEncryptionScope struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StorageEncryptionScope)(nil)

// NewStorageEncryptionScope returns knowledge for the encryptionScopes sub-resource.
func NewStorageEncryptionScope() *StorageEncryptionScope {
	return &StorageEncryptionScope{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Storage/storageAccounts/encryptionScopes",
			ApiVersions:  []string{"2025-08-01"},
			// infrastructure_encryption_required is ForceNew (schema L80) and maps to
			// the body field properties.requireInfrastructureEncryption. name and
			// storage_account_id are ForceNew too but are envelope/parent references.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.requireInfrastructureEncryption"},
			},
			// source is Required (schema L62-69); it must be present in the ARM body.
			RequiredFields: []string{
				"properties.source",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name: alphanumeric only, 4-63 chars
					// (validate.StorageEncryptionScopeName: ^[0-9a-zA-Z]{4,63}$)
					Regex:     `^[0-9a-zA-Z]{4,63}$`,
					MinLength: 4,
					MaxLength: 63,
					Message:   "must be alphanumeric, 4-63 characters",
				},
				{
					PropertyPath: "properties.source",
					// encryptionscopes.PossibleValuesForEncryptionScopeSource()
					AllowedValues: []string{"Microsoft.KeyVault", "Microsoft.Storage"},
					Message:       "must be Microsoft.KeyVault or Microsoft.Storage",
				},
			},
			// Conditional relation NOT encoded: AzureRM requires key_vault_key_id
			// (properties.keyVaultProperties.keyUri) only when source ==
			// "Microsoft.KeyVault" (storage_encryption_scope_resource.go:112-116). This is a
			// value-conditional requirement (fires on a specific enum value), which the
			// RelationalRule kinds (presence-based RequiredWith/ExactlyOneOf) cannot
			// express. Skipped (non-mappable as a declarative relation).
		},
	}
}

func init() { azwise.Register(NewStorageEncryptionScope()) }
