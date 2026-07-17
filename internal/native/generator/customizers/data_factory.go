package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/typegraph"

// customizeDataFactory applies Microsoft.DataFactory/factories schema rules that
// the bicep type graph and azwise overlay cannot express:
//   - The factory name is an operational-envelope field (not a body property), so
//     the AzureRM DataFactoryName rule (validate/datafactory.go) is attached to the
//     envelope name here rather than as an empty-path azwise StringRule (which the
//     overlay skips because name is not in the body graph).
func customizeDataFactory(def *typegraph.ResourceDefinition) {
	def.SetNameValidators(
		typegraph.LengthValidator(3, 63),
		typegraph.RegexValidator(
			`^[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*$`,
			"Data Factory name must be 3-63 characters of alphanumerics and single hyphens, starting and ending with an alphanumeric character",
		),
	)
}
