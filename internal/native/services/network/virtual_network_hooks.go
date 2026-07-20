package network

import (
	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
	"github.com/Azure/terraform-provider-azapi/internal/native/services"
)

// Virtual network hooks declare the resource-level cross-property constraint that
// is enforced at runtime as a framework ConfigValidator. It is hand-owned here
// (not generated into _gen.go) so it can be tuned without regenerating: a virtual
// network must define its address space by exactly one of an explicit prefix list
// or an IPAM pool allocation.
//
// The paired schema-level opt-out of UseStateForUnknown for these two mutually
// exclusive members (so dropping one side clears it) stays in generation, baked
// into the generated schema from the azwise overlay (see internal/native/typegraph
// azwise_overlay.go); only the validator lives here.
func init() {
	nativeresource.RegisterHooks(VirtualNetwork.Name, &nativeresource.Hooks{
		Relational: []services.RelationalConstraint{
			{Kind: services.ExactlyOneOf, Paths: []string{"properties.address_space.address_prefixes", "properties.address_space.ipam_pool_prefix_allocations"}, Message: "exactly one of address_space or ip_address_pool must be set"},
		},
	})
}
