package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/typegraph"

// customizeResourceGroup applies the schema rules the bicep type graph cannot
// express for Microsoft.Resources/resourceGroups. The resource-group name is a
// DeployTimeConstant in the ARM body (lifted onto the operational envelope), so
// AzureRM's name validation can't ride the body-path azwise overlay — it is pinned
// directly on the envelope name attribute, rejecting bad names at plan time instead
// of failing the ARM apply.
//
// Rules mirror go-azure-helpers resourcemanager/resourcegroups/validate.go:
//   - non-blank, at most 90 characters (length 1..90)
//   - alphanumerics, dash, underscore, parentheses, period; may not end with a period
func customizeResourceGroup(def *typegraph.ResourceDefinition) {
	def.Envelope.Name.Validators = []typegraph.DescriptionValidator{typegraph.
		LengthValidator(1, 90), typegraph.
		RegexValidator(
			`^[-\w._()]*[-\w_()]$`,
			"resource group name may contain only alphanumerics, dashes, underscores, parentheses and periods, and may not end with a period",
		),
	}
}
