package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/typegraph"

// customizeKustoCluster applies Microsoft.Kusto/clusters schema rules the bicep
// type graph and azwise overlay cannot express:
//   - The cluster name is an operational-envelope field (not a body property), so
//     the AzureRM ClusterName rule (validate/name.go) is attached to the envelope
//     name here rather than as an empty-path azwise StringRule (which the overlay
//     skips because name is not in the body graph).
//   - A Kusto cluster is not valid without a SKU; AzureRM requires sku_name/sku_tier
//     while the bicep graph leaves the SKU object and its fields optional. Mark sku,
//     sku.name, and sku.tier required so native users cannot plan a malformed body.
func customizeKustoCluster(def *typegraph.ResourceDefinition) {
	def.SetNameValidators(
		typegraph.LengthValidator(4, 22),
		typegraph.RegexValidator(
			`^[a-z][a-z0-9-]+$`,
			"Kusto cluster name must be 4-22 characters, start with a lowercase letter, and contain only lowercase letters, numbers, and hyphens",
		),
	)

	def.Required("sku", "sku.name", "sku.tier")
}
