package kusto

import (
	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
	"github.com/Azure/terraform-provider-azapi/internal/native/services"
)

// Kusto database hooks declare the resource-level cross-property constraint that
// is enforced at runtime as a framework ConfigValidator. It is hand-owned here
// (not generated into _gen.go) so it can be tuned without regenerating: the
// database body is a discriminated block, so exactly one variant
// (read_only_following / read_write) must be set.
func init() {
	nativeresource.RegisterHooks(KustoDatabase.Name, &nativeresource.Hooks{
		Relational: []services.RelationalConstraint{
			{Kind: services.ExactlyOneOf, Paths: []string{"read_only_following", "read_write"}, Message: "exactly one variant of the discriminated block must be set"},
		},
	})
}
