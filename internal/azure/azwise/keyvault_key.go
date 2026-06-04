package azwise

import "time"

// KeyVaultKey provides resource knowledge for Microsoft.KeyVault/vaults/keys.
//
// Sources:
//   - AzureRM key_vault_key_resource.go schema + CRUD functions
//   - go-azure-helpers keyvault/nested_item.go (name validation: alphanumeric+dash, 1-127 chars)
//   - go-azure-sdk data-plane/keyvault/7-4/keys/constants.go (key types, curves, operations)
type KeyVaultKey struct {
	BaseKnowledge
}

var _ ResourceKnowledge = (*KeyVaultKey)(nil)

func NewKeyVaultKey() *KeyVaultKey {
	return &KeyVaultKey{
		BaseKnowledge: BaseKnowledge{
			ResourceType: "Microsoft.KeyVault/vaults/keys",
			ForceNew: []ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.kty"},     // key_type → kty in ARM
				{PropertyPath: "properties.keySize"},  // key_size → keySize in ARM
				{PropertyPath: "properties.curveName"}, // curve → curveName in ARM
			},
			SoftDelete: true, // keys inherit vault soft-delete; purge via PurgeSoftDeletedKeysOnDestroy feature flag
			// key_type and key_opts are Required in AzureRM schema.
			RequiredFields: []string{
				"properties.kty",    // key_type: RSA, RSA-HSM, EC, EC-HSM
				"properties.keyOps", // key_opts: list of allowed operations
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
				// ── properties.kty ──
				// AzureRM allows EC, EC-HSM, RSA, RSA-HSM (no oct types for key_vault_key_resource).
				{
					PropertyPath:  "properties.kty",
					AllowedValues: []string{"EC", "EC-HSM", "RSA", "RSA-HSM"},
					Message:       "must be EC, EC-HSM, RSA, or RSA-HSM",
				},
				// ── properties.curveName ──
				// Only for EC/EC-HSM keys; ConflictsWith key_size.
				{
					PropertyPath:  "properties.curveName",
					AllowedValues: []string{"P-256", "P-256K", "P-384", "P-521"},
					Message:       "must be P-256, P-256K, P-384, or P-521",
				},
				// ── properties.keyOps[*] ──
				// AzureRM validates 6 operations (not export/import).
				{
					PropertyPath:  "properties.keyOps[*]",
					AllowedValues: []string{"decrypt", "encrypt", "sign", "unwrapKey", "verify", "wrapKey"},
					Message:       "must be a valid key operation",
				},
				// ── properties.attributes.notBefore ──
				// AzureRM validates with IsRFC3339Time (not_before_date).
				{
					PropertyPath: "properties.attributes.notBefore",
					Regex:        `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`,
					Message:      "must be a valid RFC3339 date/time string",
				},
				// ── properties.attributes.expires ──
				// AzureRM validates with IsRFC3339Time (expiration_date).
				{
					PropertyPath: "properties.attributes.expires",
					Regex:        `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})$`,
					Message:      "must be a valid RFC3339 date/time string",
				},
			},
			// Key material fields returned by Azure after key creation.
			ComputedFields: []string{
				"properties.key.n", // RSA modulus
				"properties.key.e", // RSA exponent
				"properties.key.x", // EC x coordinate
				"properties.key.y", // EC y coordinate
			},
			// curveName is Optional+Computed; Azure infers it for EC key types (no explicit AzureRM default).
			DefaultValues: []DefaultValue{
				{PropertyPath: "properties.curveName"}, // nil Value: server infers from key type
			},
		},
	}
}
