package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/typegraph"

// customizeKeyVault applies key-vault-specific schema rules that neither the bicep
// type graph nor the azwise overlay can express:
//   - The vault name is not part of the body type graph (it maps to the ARM ID),
//     so its AzureRM naming constraint (validate.VaultName: 3-24 chars, start with
//     a letter, end with a letter or digit, only alphanumerics and
//     non-consecutive hyphens) is attached to the envelope name attribute. The
//     azwise name rule carries the same regex, but ApplyAzwise skips empty-path
//     (name) StringRules, so it must be re-expressed here.
//   - The access-policy permission lists (certificates/keys/secrets/storage) are
//     bicep `string[]` with no enum, but AzureRM constrains each to a fixed,
//     case-insensitive value set (see the azurerm_key_vault_access_policy docs).
//     Each list gets a case-insensitive one-of validator over its elements; ARM
//     accepts any casing, so OneOfCaseInsensitive matches the API behavior.
//
// Every other azwise rule (sku.name/sku.family/createMode/publicNetworkAccess
// enums, tenantId UUID, softDeleteRetentionInDays range, defaults) targets a real
// body dot-path and is overlaid automatically by ApplyAzwise.
func customizeKeyVault(def *typegraph.ResourceDefinition) {
	def.Envelope.Name.Validators = []typegraph.DescriptionValidator{
		typegraph.LengthValidator(3, 24),
		typegraph.RegexValidator(
			`^[a-zA-Z](-?[a-zA-Z0-9])*$`,
			"key vault name must be 3-24 characters, start with a letter, end with a letter or digit, and contain only alphanumeric characters or non-consecutive hyphens",
		),
	}

	// Access-policy permission enums (case-insensitive; ARM normalizes casing).
	for path, allowed := range map[string][]string{
		"properties.accessPolicies.permissions.certificates": {
			"Backup", "Create", "Delete", "DeleteIssuers", "Get", "GetIssuers",
			"Import", "List", "ListIssuers", "ManageContacts", "ManageIssuers",
			"Purge", "Recover", "Restore", "SetIssuers", "Update",
		},
		"properties.accessPolicies.permissions.keys": {
			"Backup", "Create", "Decrypt", "Delete", "Encrypt", "Get", "Import",
			"List", "Purge", "Recover", "Restore", "Sign", "UnwrapKey", "Update",
			"Verify", "WrapKey", "Release", "Rotate", "GetRotationPolicy",
			"SetRotationPolicy",
		},
		"properties.accessPolicies.permissions.secrets": {
			"Backup", "Delete", "Get", "List", "Purge", "Recover", "Restore", "Set",
		},
		"properties.accessPolicies.permissions.storage": {
			"Backup", "Delete", "DeleteSAS", "Get", "GetSAS", "List", "ListSAS",
			"Purge", "Recover", "RegenerateKey", "Restore", "Set", "SetSAS", "Update",
		},
	} {
		p := typegraph.FindProperty(def, path)
		p.Validators = append(p.Validators, typegraph.OneOfCaseInsensitiveValidator(
			"must be a valid key vault permission (case-insensitive)", allowed...,
		))
	}

	// ipRules / virtualNetworkRules are Optional+Computed lists that plan known-null
	// when the user configures network_acls with only default_action/bypass, but the
	// Key Vault RP always echoes them back as []. An empty-list default makes the
	// omitted-config plan value ([]) match the API's [] on apply/read/import, while an
	// explicitly-configured list (including an explicit []) is left untouched.
	for _, path := range []string{
		"properties.networkAcls.ipRules",
		"properties.networkAcls.virtualNetworkRules",
	} {
		typegraph.FindProperty(def, path).DefaultEmptyList = true
	}

	// purge_on_destroy is a synthetic behavior-only attribute (not part of the bicep
	// body): when true, destroying the vault also purges its soft-deleted shadow so
	// the name is immediately reusable. It defaults off (null) because a purge bypasses
	// the soft-delete recovery window and is irreversible; a runtime AfterDelete hook
	// (see key_vault_hooks.go) reads it from state and runs the deletedVaults purge.
	def.Envelope.Meta = append(def.Envelope.Meta, typegraph.MetaAttr{
		Name: "purge_on_destroy",
		Description: "When `true`, permanently purges the vault's soft-deleted shadow on destroy so its " +
			"name can be reused immediately, instead of leaving it recoverable until Azure's retention " +
			"window expires. Ignored when purge protection is enabled (Azure blocks the purge). Defaults to `false`.",
	})
}
