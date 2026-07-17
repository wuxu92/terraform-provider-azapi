package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/typegraph"

// customizeStorageAccount applies storage-account-specific schema rules that the
// bicep type graph cannot express:
//   - The resource name is not part of the body type graph, so its well-known
//     constraints (3-24 characters, lowercase letters and digits only) are
//     attached to the envelope name attribute.
//   - network_acls.ip_rules[].value must be a public IPv4 address or CIDR range
//     (StorageAccountIPRule, in services/storage/validators).
//   - Several body fields carry generic AzureRM validators ported as shared
//     nativeschema validators: UUID for the AD domain GUID and resource-access-rule
//     tenant IDs, and AzureResourceID for resource-access-rule resource IDs.
func customizeStorageAccount(def *typegraph.ResourceDefinition) {
	def.SetNameValidators(
		typegraph.LengthValidator(3, 24),
		typegraph.RegexValidator(
			`^[a-z0-9]+$`,
			"storage account name must be 3-24 characters of lowercase letters and digits only",
		),
	)

	// IPv4-only. ipRules and ipv6Rules share one bicep element type, so isolate the
	// ipRules element first — otherwise the IPv4 validator would also land on
	// ipv6Rules.value and reject valid IPv6 entries.
	typegraph.IsolateArrayElement(def, "properties.networkAcls.ipRules")
	def.AddValidatorsFor("properties.networkAcls.ipRules.value", typegraph.CustomValidator("StorageAccountIPRule()"))

	// AD domain GUID is a UUID (AzureRM validation.IsUUID).
	def.AddValidatorsFor("properties.azureFilesIdentityBasedAuthentication.activeDirectoryProperties.domainGuid", typegraph.SharedValidator("UUID()"))

	// resourceAccessRules: resourceId is an ARM resource ID, tenantId is a UUID
	// (AzureRM azure.ValidateResourceID / validation.IsUUID). Isolate the element
	// first so the rules don't leak to any array that shares its element type.
	typegraph.IsolateArrayElement(def, "properties.networkAcls.resourceAccessRules")
	def.AddValidatorsFor("properties.networkAcls.resourceAccessRules.resourceId", typegraph.SharedValidator("AzureResourceID()"))
	def.AddValidatorsFor("properties.networkAcls.resourceAccessRules.tenantId", typegraph.SharedValidator("UUID()"))
}
