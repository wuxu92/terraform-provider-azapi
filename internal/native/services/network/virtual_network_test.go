package network_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/network"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// Azure Virtual Network is exercised as a resource-group-scoped native static
// resource. The resource group base lives at the workspace root scope; the vnet
// scenario creates a minimal single-CIDR network, imports it drift-free, then widens
// the address space and sets the flow-timeout / encryption body fields in place. The
// test config comes from the network.VirtualNetworkCfg builder defined alongside the
// generated resource in this package, which also vends the matching acceptance handle.
var _ = Describe("Azure Virtual Network", Ordered, func() {
	ws := acc.NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	// Root scope: the resource group base reused by the vnet scenario.
	rgCfg := resources.NewResourceGroupCfg()
	rg := ws.ResourceFor(rgCfg)
	BeforeAll(func() {
		rg.Apply(resources.ResourceGroupCfg_Basic(rgCfg), acc.Exists())
	})

	cfg := network.NewVirtualNetworkCfg(rgCfg)
	vnet := ws.ResourceFor(cfg)

	It("creates a virtual network, then imports it with no drift", func() {
		vnet.Apply(network.VirtualNetworkCfg_Basic(cfg), acc.Exists()).ImportVerify()
	})

	When("Complete", func() {
		It("applies the complete body surface in place", func() {
			// Complete adds a widened address space, DHCP DNS, flow-timeout, encryption,
			// tags, and two subnets (service endpoints + a delegation); each Apply re-plans
			// for drift, so the in-place update round-trips.
			vnet.Apply(network.VirtualNetworkCfg_Complete(cfg)).ImportVerify()
		})

		It("mutates the complete surface in place", func() {
			// Complete_update retags, swaps DNS servers, bumps the flow timeout, flips both
			// subnet private-network policies, widens the service-endpoint set, and adds a
			// third subnet — proving the full mutable surface survives Update -> Read.
			vnet.Apply(network.VirtualNetworkCfg_Complete_update(cfg)).ImportVerify()
		})

	})
})
