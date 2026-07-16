package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/typegraph"

// customizeKustoClusterDatabase applies Microsoft.Kusto/clusters/databases schema
// rules the bicep type graph and azwise overlay cannot express:
//   - The database name is an operational-envelope field (not a body property), so
//     the AzureRM DatabaseName rule (validate/name.go) is attached to the envelope
//     name here rather than as an empty-path azwise StringRule (which the overlay
//     skips because name is not in the body graph).
func customizeKustoClusterDatabase(def *typegraph.ResourceDefinition) {
	def.Envelope.Name.Validators = []typegraph.DescriptionValidator{
		typegraph.LengthValidator(1, 260),
		typegraph.RegexValidator(
			`^[a-zA-Z0-9\s._-]+$`,
			"Kusto database name must be at most 260 characters of alphanumerics, whitespace, dots, dashes, and underscores",
		),
	}
}
