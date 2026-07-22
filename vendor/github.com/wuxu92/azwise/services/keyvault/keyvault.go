package keyvault

import (
	"time"

	"github.com/wuxu92/azwise"
)

// KeyVault provides resource knowledge for Microsoft.KeyVault/vaults.
//
// Sources:
//   - AzureRM key_vault_resource.go schema + CRUD functions
//   - AzureRM access_policy_schema.go (permission enums)
//   - AzureRM key_vault_access_policy_resource.go (application_id → accessPolicies[*].applicationId, IsUUID)
//   - AzureRM validate/vault_name.go
//   - AzureRM helpers/validate/network.go (IPv4/CIDR)
//   - Azure SDK vaults/constants.go (2023-02-01)
type KeyVault struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*KeyVault)(nil)

// NewKeyVault returns a KeyVault knowledge instance.
func NewKeyVault() *KeyVault {
	return &KeyVault{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.KeyVault/vaults",
			ApiVersions:  []string{"2023-02-01"},
			ForceNew:     []azwise.ForceNewRule{{PropertyPath: "name"}},
			SoftDelete:   true,
			// ARM API requires tenantId, sku.name, and sku.family for vault creation.
			// AzureRM hardcodes sku.family to "A" (the only valid value).
			RequiredFields: []string{
				"properties.tenantId",
				"properties.sku.name",
				"properties.sku.family",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ──
				// AzureRM validate.VaultName: ^[a-zA-Z0-9-]{3,24}$, must start with
				// letter, end with letter/digit, no consecutive hyphens.
				// The regex below encodes all three checks in one pattern:
				//   [a-zA-Z]        → starts with letter
				//   (-?[a-zA-Z0-9]) → optional single hyphen then alphanumeric (no --)
				//   *               → zero or more of the above group
				// Ending on [a-zA-Z0-9] is guaranteed because the last group element is alphanumeric.
				// Length is enforced via MinLength/MaxLength.
				{
					Regex:     `^[a-zA-Z](-?[a-zA-Z0-9])*$`,
					MinLength: 3,
					MaxLength: 24,
					Message:   "must start with a letter, end with letter or digit, contain only alphanumeric characters or non-consecutive hyphens, 3-24 characters",
				},
				// ── properties.sku.family ──
				// Hardcoded as vaults.SkuFamilyA ("A") in create and update.
				{
					PropertyPath:  "properties.sku.family",
					AllowedValues: []string{"A"},
					Message:       "SKU family must be A",
				},
				// ── properties.sku.name ──
				// validation.StringInSlice(["standard","premium"], false) in schema.
				{
					PropertyPath:  "properties.sku.name",
					AllowedValues: []string{"standard", "premium"},
					Message:       "must be standard or premium",
				},
				// ── properties.createMode ──
				// Hardcoded to vaults.CreateModeDefault ("default") or
				// vaults.CreateModeRecover ("recover") in resourceKeyVaultCreate.
				{
					PropertyPath:  "properties.createMode",
					AllowedValues: []string{"default", "recover"},
					Message:       "must be default or recover",
				},
				// ── properties.publicNetworkAccess ──
				// Hardcoded to "Enabled" or "Disabled" in create and update.
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be Enabled or Disabled",
				},
				// ── properties.networkAcls.defaultAction ──
				// validation.StringInSlice(["Allow","Deny"], false) in schema.
				{
					PropertyPath:  "properties.networkAcls.defaultAction",
					AllowedValues: []string{"Allow", "Deny"},
					Message:       "must be Allow or Deny",
				},
				// ── properties.networkAcls.bypass ──
				// validation.StringInSlice(["AzureServices","None"], false) in schema.
				{
					PropertyPath:  "properties.networkAcls.bypass",
					AllowedValues: []string{"AzureServices", "None"},
					Message:       "must be AzureServices or None",
				},
				// ── properties.networkAcls.ipRules[*].value ──
				// validation.Any(commonValidate.IPv4Address, commonValidate.CIDR) in schema.
				// Accepts a plain IPv4 address or an IPv4 CIDR block.
				{
					PropertyPath: "properties.networkAcls.ipRules[*].value",
					Regex:        `^([0-9]{1,3}\.){3}[0-9]{1,3}(/([0-9]|[1-2][0-9]|3[0-2]))?$`,
					Message:      "must be a valid IPv4 address or CIDR (e.g. 10.0.0.1 or 10.0.0.0/24)",
				},
				// ── properties.tenantId ──
				// validation.IsUUID in schema.
				{
					PropertyPath: "properties.tenantId",
					Regex:        `(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
					Message:      "must be a valid UUID",
				},
				// ── properties.accessPolicies[*].tenantId ──
				// validation.IsUUID in schema.
				{
					PropertyPath: "properties.accessPolicies[*].tenantId",
					Regex:        `(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
					Message:      "must be a valid UUID",
				},
				// ── properties.accessPolicies[*].objectId ──
				// validation.IsUUID in schema.
				{
					PropertyPath: "properties.accessPolicies[*].objectId",
					Regex:        `(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
					Message:      "must be a valid UUID",
				},
				// ── properties.accessPolicies[*].applicationId ──
				// key_vault_access_policy_resource.go application_id: validation.IsUUID (Optional, ForceNew).
				{
					PropertyPath: "properties.accessPolicies[*].applicationId",
					Regex:        `(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
					Message:      "must be a valid UUID",
				},
				// ── properties.accessPolicies[*].permissions.certificates[*] ──
				// validation.StringInSlice(certificatePermissions(), false) in schema.
				{
					PropertyPath: "properties.accessPolicies[*].permissions.certificates[*]",
					AllowedValues: []string{
						"Get", "List", "Update", "Create", "Import", "Delete",
						"Recover", "Backup", "Restore",
						"ManageContacts", "ManageIssuers", "GetIssuers", "ListIssuers", "SetIssuers", "DeleteIssuers",
						"Purge",
					},
					Message: "must be a valid certificate permission",
				},
				// ── properties.accessPolicies[*].permissions.keys[*] ──
				// validation.StringInSlice(keyPermissions(), false) in schema.
				{
					PropertyPath: "properties.accessPolicies[*].permissions.keys[*]",
					AllowedValues: []string{
						"Get", "List", "Update", "Create", "Import", "Delete",
						"Recover", "Backup", "Restore",
						"Decrypt", "Encrypt", "UnwrapKey", "WrapKey", "Verify", "Sign",
						"Purge", "Release",
						"Rotate", "GetRotationPolicy", "SetRotationPolicy",
					},
					Message: "must be a valid key permission",
				},
				// ── properties.accessPolicies[*].permissions.secrets[*] ──
				// validation.StringInSlice(secretPermissions(), false) in schema.
				{
					PropertyPath: "properties.accessPolicies[*].permissions.secrets[*]",
					AllowedValues: []string{
						"Get", "List", "Set", "Delete",
						"Recover", "Backup", "Restore",
						"Purge",
					},
					Message: "must be a valid secret permission",
				},
				// ── properties.accessPolicies[*].permissions.storage[*] ──
				// validation.StringInSlice(storagePermissions(), false) in schema.
				{
					PropertyPath: "properties.accessPolicies[*].permissions.storage[*]",
					AllowedValues: []string{
						"Backup", "Delete", "DeleteSAS", "Get", "GetSAS",
						"List", "ListSAS", "Purge", "Recover", "RegenerateKey",
						"Restore", "Set", "SetSAS", "Update",
					},
					Message: "must be a valid storage permission",
				},
			},
			IntRules: []azwise.IntRule{
				// ── properties.softDeleteRetentionInDays ──
				// validation.IntBetween(7, 90), default 90.
				{
					PropertyPath: "properties.softDeleteRetentionInDays",
					MinValue:     azwise.Ptr(int64(7)),
					MaxValue:     azwise.Ptr(int64(90)),
					Message:      "soft delete retention must be between 7 and 90 days",
				},
			},
			ArrayRules: []azwise.ArrayRule{
				// ── properties.accessPolicies ──
				// MaxItems: 1024 in schema.
				{
					PropertyPath: "properties.accessPolicies",
					MaxItems:     1024,
					Message:      "a key vault supports a maximum of 1024 access policies",
				},
			},
			// vault_uri is returned by Azure, never set by the user.
			ComputedFields: []string{
				"properties.vaultUri",
			},
			// These properties are optional; Azure/AzureRM fills in defaults if omitted.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.accessPolicies"},                                // empty list when omitted
				{PropertyPath: "properties.enableRbacAuthorization", Value: true},          // RBAC is Azure's default authorization model for new vaults (learn.microsoft.com/azure/key-vault/general/access-control-default); azapi defaults it on, unlike AzureRM which defaults false
				{PropertyPath: "properties.networkAcls.defaultAction", Value: "Allow"},     // expand default when block absent
				{PropertyPath: "properties.networkAcls.bypass", Value: "AzureServices"},    // expand default when block absent
				{PropertyPath: "properties.softDeleteRetentionInDays", Value: float64(90)}, // Azure default 90 days
				{PropertyPath: "properties.enableSoftDelete", Value: true},                 // Azure enforced true since 2025
				// Azure rejects an explicit properties.enablePurgeProtection=false in the
				// PUT body ("cannot be set to false. Enabling the purge protection ... is
				// an irreversible action"): the flag accepts only true or omission.
				// AzureRM omits it unless enabling, so no explicit default — nil Value
				// leaves the attribute Optional+Computed and the body omits it when unset.
				{PropertyPath: "properties.enablePurgeProtection"},
			},
		},
	}
}

func init() { azwise.Register(NewKeyVault()) }
