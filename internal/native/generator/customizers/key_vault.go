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
}
