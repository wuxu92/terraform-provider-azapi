package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/typegraph"

// customizeDeploymentScript applies Microsoft.Resources/deploymentScripts schema
// rules the bicep type graph and azwise overlay cannot express:
//   - The deployment-script name is an operational-envelope field (not a body
//     property), so the AzureRM name rule is attached to the envelope name here
//     rather than as an empty-path azwise StringRule (which the overlay skips
//     because name is not in the body graph).
func customizeDeploymentScript(def *typegraph.ResourceDefinition) {
	def.Envelope.Name.Validators = []typegraph.DescriptionValidator{
		typegraph.LengthValidator(1, 260),
		typegraph.RegexValidator(
			`^[a-zA-Z0-9_().-]{0,259}[a-zA-Z0-9_()-]$`,
			"deployment script name must be 1-260 characters of alphanumerics, underscore, parentheses, hyphen, and period, and cannot end with a period",
		),
	}
}
