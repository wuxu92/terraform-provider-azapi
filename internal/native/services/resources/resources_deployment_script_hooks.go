package resources

import (
	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
	"github.com/Azure/terraform-provider-azapi/internal/native/services"
)

// Deployment script hooks declare the resource-level cross-property constraint
// that is enforced at runtime as a framework ConfigValidator. It is hand-owned
// here (not generated into _gen.go) so it can be tuned without regenerating: the
// deployment script body is a discriminated root, so exactly one variant
// (azure_cli / azure_power_shell) must be set.
func init() {
	nativeresource.RegisterHooks(ResourcesDeploymentScript.Name, &nativeresource.Hooks{
		Relational: []services.RelationalConstraint{
			{Kind: services.ExactlyOneOf, Paths: []string{"azure_cli", "azure_power_shell"}, Message: "exactly one variant of the discriminated block must be set"},
		},
	})
}
