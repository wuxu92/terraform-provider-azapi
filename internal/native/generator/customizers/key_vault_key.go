package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/typegraph"

// customizeKeyVaultKey applies Key Vault key schema rules that need to be pinned
// explicitly for the generated Terraform schema:
//   - The key name maps to the ARM ID, not a body property, so the management-plane
//     name constraint must be attached to the generated envelope name attribute.
//   - keyOps is an array whose element enum is not emitted mechanically; use the
//     management-plane ARM operation set, including import and release.
func customizeKeyVaultKey(def *typegraph.ResourceDefinition) {
	def.Envelope.Name.Validators = []typegraph.DescriptionValidator{
		typegraph.LengthValidator(1, 127),
		typegraph.RegexValidator(
			`^[0-9a-zA-Z-]+$`,
			"key vault key name must contain only alphanumeric characters and dashes, 1-127 characters",
		),
	}

	// AzureRM requires key_type and key_opts, and azwise RequiredFields is planner
	// metadata rather than a native schema flag. Promote the matching ARM properties
	// so invalid static-resource configs fail before ARM rejects the PUT.
	typegraph.FindProperty(def, "properties.kty").Flags |= typegraph.FlagRequired
	keyOps := typegraph.FindProperty(def, "properties.keyOps")
	keyOps.Flags |= typegraph.FlagRequired
	keyOps.Validators = append(keyOps.Validators, typegraph.OneOfValidator(
		"must be a valid management-plane key operation",
		"decrypt", "encrypt", "import", "release", "sign", "unwrapKey", "verify", "wrapKey",
	))
}
