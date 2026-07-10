package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/typegraph"

// customizeRoleAssignment applies role-assignment-specific schema rules that the
// bicep type graph and azwise overlay cannot express:
//   - The ARM resource name is the role assignment GUID (azurerm's optional name),
//     so the envelope name gets the UUID validator.
//   - delegatedManagedIdentityResourceId is an ARM resource ID string; AzureRM uses
//     azure.ValidateResourceID, so attach the shared native resource-id validator.
func customizeRoleAssignment(def *typegraph.ResourceDefinition) {
	def.Envelope.Name.Validators = []typegraph.DescriptionValidator{
		typegraph.SharedValidator("UUID()"),
	}
	def.Envelope.Parent.Name = "scope_id"
	def.Envelope.Parent.Description = "The scope ID where this role assignment applies."

	delegatedID := typegraph.FindProperty(def, "properties.delegatedManagedIdentityResourceId")
	delegatedID.Validators = append(delegatedID.Validators, typegraph.SharedValidator("AzureResourceID()"))
}
