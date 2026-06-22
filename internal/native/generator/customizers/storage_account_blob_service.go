package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/generator"

// customizeStorageAccountBlobService applies the schema rules the bicep type
// graph cannot express for Microsoft.Storage/storageAccounts/blobServices:
//   - blobServices is a singleton child resource whose ARM name is always
//     "default". The name is not part of the body type graph, so the constraint
//     is attached to the envelope name attribute, rejecting any other value at
//     plan time instead of failing the ARM apply.
func customizeStorageAccountBlobService(def *generator.ResourceDefinition) {
	def.Envelope.Name.Validators = []generator.DescriptionValidator{
		generator.OneOfValidator(
			`blob service name must be "default" (blobServices is a singleton child resource)`,
			"default",
		),
	}
}
