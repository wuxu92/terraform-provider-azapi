package azwise

import "time"

// KeyVaultSecret provides resource knowledge for Microsoft.KeyVault/vaults/secrets.
//
// Sources:
//   - AzureRM key_vault_secret_resource.go schema + CRUD functions
//   - go-azure-helpers keyvault/nested_item.go (name validation: alphanumeric+dash, 1-127 chars)
type KeyVaultSecret struct {
	BaseKnowledge
}

var _ ResourceKnowledge = (*KeyVaultSecret)(nil)

func NewKeyVaultSecret() *KeyVaultSecret {
	return &KeyVaultSecret{
		BaseKnowledge: BaseKnowledge{
			ResourceType: "Microsoft.KeyVault/vaults/secrets",
			ForceNew: []ForceNewRule{
				{PropertyPath: "name"},
			},
			SoftDelete: true, // secrets inherit vault soft-delete; purge via PurgeSoftDeletedSecretsOnDestroy feature flag
			// value is Required (ExactlyOneOf with value_wo in AzureRM).
			// It's also Sensitive, so it may appear in sensitive_body instead of body.
			RequiredFields: []string{
				"properties.value",
			},
			TimeoutsConfig: &Timeouts{
				Create: 30 * time.Minute,
				Read:   30 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []StringRule{
				// ── Resource name ──
				// ValidateNestedItemName: ^[0-9a-zA-Z-]+$, 1-127 chars.
				{
					Regex:     `^[0-9a-zA-Z-]+$`,
					MinLength: 1,
					MaxLength: 127,
					Message:   "must contain only alphanumeric characters and dashes, 1-127 characters",
				},
				// ── not_before_date → properties.attributes.notBefore ──
				// IsRFC3339Time validation in AzureRM schema.
				{
					PropertyPath: "properties.attributes.notBefore",
					Regex:        `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`,
					Message:      "must be a valid RFC3339 date/time (e.g. 2025-01-01T00:00:00Z)",
				},
				// ── expiration_date → properties.attributes.expires ──
				// IsRFC3339Time validation in AzureRM schema.
				{
					PropertyPath: "properties.attributes.expires",
					Regex:        `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`,
					Message:      "must be a valid RFC3339 date/time (e.g. 2025-01-01T00:00:00Z)",
				},
			},
			SensitiveFields: []string{
				"properties.value", // Sensitive: true in AzureRM schema
			},
		},
	}
}
