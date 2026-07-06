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
	def.Envelope.Name.Validators = []typegraph.DescriptionValidator{typegraph.
		LengthValidator(3, 24), typegraph.
		RegexValidator(
			`^[a-z0-9]+$`,
			"storage account name must be 3-24 characters of lowercase letters and digits only",
		),
	}
	typegraph.

		// IPv4-only. ipRules and ipv6Rules share one bicep element type, so isolate
		// the ipRules element first — otherwise the IPv4 validator would also land on
		// ipv6Rules.value and reject valid IPv6 entries. Both helpers panic on a bad
		// path, so a typo fails the generator immediately.
		IsolateArrayElement(def, "properties.networkAcls.ipRules")
	ipv4 := typegraph.FindProperty(def, "properties.networkAcls.ipRules.value")
	ipv4.Validators = append(ipv4.Validators, typegraph.CustomValidator("StorageAccountIPRule()"))

	// AD domain GUID is a UUID (AzureRM validation.IsUUID).
	domainGUID := typegraph.FindProperty(def, "properties.azureFilesIdentityBasedAuthentication.activeDirectoryProperties.domainGuid")
	domainGUID.Validators = append(domainGUID.Validators, typegraph.SharedValidator("UUID()"))
	typegraph.

		// resourceAccessRules: resourceId is an ARM resource ID, tenantId is a UUID
		// (AzureRM azure.ValidateResourceID / validation.IsUUID). Isolate the element
		// first so the rules don't leak to any array that shares its element type.
		IsolateArrayElement(def, "properties.networkAcls.resourceAccessRules")
	rarID := typegraph.FindProperty(def, "properties.networkAcls.resourceAccessRules.resourceId")
	rarID.Validators = append(rarID.Validators, typegraph.SharedValidator("AzureResourceID()"))
	rarTenant := typegraph.FindProperty(def, "properties.networkAcls.resourceAccessRules.tenantId")
	rarTenant.Validators = append(rarTenant.Validators, typegraph.SharedValidator("UUID()"))
}
