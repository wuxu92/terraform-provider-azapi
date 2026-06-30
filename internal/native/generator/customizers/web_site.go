package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/generator"

// customizeWebSite applies Microsoft.Web/sites schema rules that the bicep type
// graph cannot express:
//   - The site name is an operational-envelope field, not a body property, so the
//     AzureRM WebAppName rule is attached here.
//   - The associated App Service plan ID and optional subnet ID are ARM resource
//     IDs. AzureRM validates them semantically; attach shared resource-ID
//     validators where the bicep graph only knows "string".
//   - siteConfig is a write-only ARM shape. Static child defaults there create
//     synthetic Terraform plan diffs because Azure read responses cannot reliably
//     confirm omitted/defaulted children; keep explicit user values only.
func customizeWebSite(def *generator.ResourceDefinition) {
	def.Envelope.Name.Validators = []generator.DescriptionValidator{
		generator.LengthValidator(1, 60),
		generator.RegexValidator(
			`^[0-9A-Za-z-]+$`,
			"site name may only contain alphanumeric characters and dashes, up to 60 characters",
		),
	}

	for _, path := range []string{
		"properties.serverFarmId",
		"properties.virtualNetworkSubnetId",
	} {
		p := generator.FindProperty(def, path)
		p.Validators = append(p.Validators, generator.SharedValidator("AzureResourceID()"))
	}

	siteConfig := generator.FindProperty(def, "properties.siteConfig")
	clearDefaultValues(siteConfig.Type)
}

func clearDefaultValues(typ *generator.Type) {
	if typ == nil {
		return
	}
	switch typ.Kind {
	case generator.KindObject:
		for _, prop := range typ.Properties {
			if prop == nil {
				continue
			}
			prop.DefaultValue = ""
			clearDefaultValues(prop.Type)
		}
	case generator.KindArray, generator.KindMap:
		clearDefaultValues(typ.ElementType)
	}
}
