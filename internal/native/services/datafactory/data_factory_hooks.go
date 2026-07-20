package datafactory

import (
	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
	"github.com/Azure/terraform-provider-azapi/internal/native/services"
)

// Data factory hooks declare the resource-level cross-property constraints that
// are enforced at runtime as framework ConfigValidators. They are hand-owned here
// (not generated into _gen.go) so they can be tuned without regenerating:
//   - Customer-managed key encryption requires a user-assigned identity.
//   - repo_configuration is a discriminated block: at most one variant
//     (GitHub / VSTS) may be set.
func init() {
	nativeresource.RegisterHooks(DataFactory.Name, &nativeresource.Hooks{
		Relational: []services.RelationalConstraint{
			{Kind: services.RequiredWith, Paths: []string{"properties.encryption.key_name", "properties.encryption.identity.user_assigned_identity"}, Message: "customer-managed-key encryption requires a user-assigned identity (properties.encryption.identity.userAssignedIdentity)"},
			{Kind: services.AtMostOneOf, Paths: []string{"properties.repo_configuration.factory_git_hub_configuration", "properties.repo_configuration.factory_vsts_configuration"}, Message: "at most one variant of the discriminated block may be set"},
		},
	})
}
