package managedidentity_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/managedidentity"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// Azure User Assigned Identity demonstrates the scoped model and cross-resource config
// reuse: the resource group (its builder lives in the resources package) is a root-scope
// base, and a nested scope owns the identity under test on that shared base. The identity
// is created once in the nested BeforeAll, asserted and mutated by the ordered scenarios,
// and destroyed once when the nested container ends — the resource group survives for any
// sibling container.
var _ = Describe("Azure User Assigned Identity", Ordered, func() {
	ws := acc.NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	// Root scope: the resource group base, reused by every nested scope. Its config
	// comes from the resources package's ResourceGroup builder.
	rgCfg := resources.NewResourceGroupCfg()
	rg := ws.ResourceFor(rgCfg)
	BeforeAll(func() { rg.Apply(resources.ResourceGroupCfg_Basic(rgCfg), acc.Exists()) })

	Describe("a user-assigned identity", Ordered, func() {
		// Child scope: owns the identity, torn down after this container while the
		// resource group base stays. The identity injects the resource group's Terraform
		// address (rgCfg.IDRef) as its resource_group_id reference.
		scope := ws.Scope()
		cfg := managedidentity.NewUserAssignedIdentityCfg(rgCfg)
		uai := scope.ResourceFor(cfg)

		BeforeAll(func() {
			uai.Apply(managedidentity.UserAssignedIdentityCfg_Basic(cfg), acc.Exists()).ImportVerify()
		})

		It("adds tags in place", func() {
			// Complete adds tags — the identity's only in-place-updatable user surface;
			// each Apply re-plans for drift, so the in-place add round-trips without
			// per-value assertions.
			uai.Apply(managedidentity.UserAssignedIdentityCfg_Complete(cfg))
		})

		It("updates tags in place", func() {
			// Complete_update flips the tag value set by Complete; each Apply re-plans for
			// drift, so the update round-trips without per-value assertions.
			uai.Apply(managedidentity.UserAssignedIdentityCfg_Complete_update(cfg))
		})
	})
})
