package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/typegraph"

// customizeRoleAssignment applies role-assignment-specific schema rules that the
// bicep type graph and azwise overlay cannot express:
//   - The ARM resource name is the role assignment GUID (azurerm's optional name),
//     so the envelope name gets the UUID validator.
//   - delegatedManagedIdentityResourceId is an ARM resource ID string; AzureRM uses
//     azure.ValidateResourceID, so attach the shared native resource-id validator.
func customizeRoleAssignment(def *typegraph.ResourceDefinition) {
	def.SetNameValidators(typegraph.SharedValidator("UUID()"))
	def.SetParent("scope_id", "The scope ID where this role assignment applies.")

	def.AddValidatorsFor("properties.delegatedManagedIdentityResourceId", typegraph.SharedValidator("AzureResourceID()"))
}
