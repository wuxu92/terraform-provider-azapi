package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/generator"

// customizeWebServerFarm applies Microsoft.Web/serverfarms schema rules that the
// bicep type graph cannot express:
//   - The App Service plan name is an operational-envelope field, not a body
//     property, so the AzureRM ServicePlanName rule is attached here.
//   - A server farm is not useful without a SKU; AzureRM requires sku_name and
//     maps it to sku.name, while the ARM bicep graph leaves the SKU object
//     optional. Mark sku and sku.name required so native users cannot plan a
//     malformed App Service plan body.
//   - hostingEnvironmentProfile.id is an ARM resource ID. AzureRM validates the
//     App Service Environment ID semantically; attach the shared resource-ID
//     validator where the bicep graph only knows "string".
func customizeWebServerFarm(def *generator.ResourceDefinition) {
	def.Envelope.Name.Validators = []generator.DescriptionValidator{
		generator.LengthValidator(1, 60),
		generator.RegexValidator(
			`^[0-9A-Za-z-_]+$`,
			"App Service plan name may only contain alphanumeric characters, dashes, and underscores, up to 60 characters",
		),
	}

	sku := generator.FindProperty(def, "sku")
	sku.Flags |= generator.FlagRequired
	skuName := generator.FindProperty(def, "sku.name")
	skuName.Flags |= generator.FlagRequired

	aseID := generator.FindProperty(def, "properties.hostingEnvironmentProfile.id")
	aseID.Validators = append(aseID.Validators, generator.SharedValidator("AzureResourceID()"))
}
