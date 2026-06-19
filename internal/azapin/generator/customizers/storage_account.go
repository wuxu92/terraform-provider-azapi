package customizers

import "github.com/Azure/terraform-provider-azapi/internal/azapin/generator"

func init() {
	Register("Microsoft.Storage/storageAccounts", customizeStorageAccount)
}

// customizeStorageAccount applies storage-account-specific schema rules that the
// bicep type graph cannot express. The resource name is not part of the body
// type graph, so its well-known constraints (3-24 characters, lowercase letters
// and digits only) are attached to the envelope name attribute here and baked
// into the generated schema for plan-time validation.
func customizeStorageAccount(def *generator.ResourceDefinition) {
	def.Envelope.Name.Validators = []generator.DescriptionValidator{
		generator.LengthValidator(3, 24),
		generator.RegexValidator(
			`^[a-z0-9]+$`,
			"storage account name must be 3-24 characters of lowercase letters and digits only",
		),
	}
}
