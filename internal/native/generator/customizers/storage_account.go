package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/generator"

// customizeStorageAccount applies storage-account-specific schema rules that the
// bicep type graph cannot express:
//   - The resource name is not part of the body type graph, so its well-known
//     constraints (3-24 characters, lowercase letters and digits only) are
//     attached to the envelope name attribute.
//   - network_acls.ip_rules[].value must be a public IPv4 address or CIDR range
//     (StorageAccountIPRule, in generated/storage/validators).
//   - Several body fields carry generic AzureRM validators ported as shared
//     nativeschema validators: UUID for the AD domain GUID and resource-access-rule
//     tenant IDs, and AzureResourceID for resource-access-rule resource IDs.
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

	// AD domain GUID is a UUID (AzureRM validation.IsUUID).
	domainGUID := generator.FindProperty(def, "properties.azureFilesIdentityBasedAuthentication.activeDirectoryProperties.domainGuid")
	domainGUID.Validators = append(domainGUID.Validators, generator.SharedValidator("UUID()"))

	// resourceAccessRules: resourceId is an ARM resource ID, tenantId is a UUID
	// (AzureRM azure.ValidateResourceID / validation.IsUUID). Isolate the element
	// first so the rules don't leak to any array that shares its element type.
	generator.IsolateArrayElement(def, "properties.networkAcls.resourceAccessRules")
	rarID := generator.FindProperty(def, "properties.networkAcls.resourceAccessRules.resourceId")
	rarID.Validators = append(rarID.Validators, generator.SharedValidator("AzureResourceID()"))
	rarTenant := generator.FindProperty(def, "properties.networkAcls.resourceAccessRules.tenantId")
	rarTenant.Validators = append(rarTenant.Validators, generator.SharedValidator("UUID()"))
}
