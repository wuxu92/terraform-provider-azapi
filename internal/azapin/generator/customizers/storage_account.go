package customizers

import "github.com/Azure/terraform-provider-azapi/internal/azapin/generator"

// customizeStorageAccount applies storage-account-specific schema rules that the
// bicep type graph cannot express:
//   - The resource name is not part of the body type graph, so its well-known
//     constraints (3-24 characters, lowercase letters and digits only) are
//     attached to the envelope name attribute.
//   - network_acls.ip_rules[].value must be a public IPv4 address or CIDR range,
//     a semantic rule ported from AzureRM that a plain regex/enum can't express;
//     it references the StorageAccountIPRule validator in the storage service's
//     validators package (generated/storage/validators).
func customizeStorageAccount(def *generator.ResourceDefinition) {
	def.Envelope.Name.Validators = []generator.DescriptionValidator{
		generator.LengthValidator(3, 24),
		generator.RegexValidator(
			`^[a-z0-9]+$`,
			"storage account name must be 3-24 characters of lowercase letters and digits only",
		),
	}

	// IPv4-only. ipRules and ipv6Rules share one bicep element type, so isolate
	// the ipRules element first — otherwise the IPv4 validator would also land on
	// ipv6Rules.value and reject valid IPv6 entries. Both helpers panic on a bad
	// path, so a typo fails the generator immediately.
	generator.IsolateArrayElement(def, "properties.networkAcls.ipRules")
	ipv4 := generator.FindProperty(def, "properties.networkAcls.ipRules.value")
	ipv4.Validators = append(ipv4.Validators, generator.CustomValidator("StorageAccountIPRule()"))
}
