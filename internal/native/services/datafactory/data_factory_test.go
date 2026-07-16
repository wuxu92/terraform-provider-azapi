package datafactory_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/datafactory"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// Azure Data Factory exercises a native resource with a nested discriminated property:
// properties.repo_configuration is discriminated by type (FactoryGitHubConfiguration /
// FactoryVSTSConfiguration). Those variants need a live Git account, so this suite
// covers the dependency-free lifecycle instead: a minimal factory is created and
// imported drift-free, then an in-place update disables public network access and adds
// a global parameter — each Apply re-plans for drift, so the round-trip proves the
// update path and the mapper flatten without any external dependency.
var _ = Describe("Azure Data Factory", Ordered, func() {
	ws := acc.NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	// Root scope: the resource group base reused by the factory scenario.
	rgCfg := resources.NewResourceGroupCfg()
	rg := ws.ResourceFor(rgCfg)
	BeforeAll(func() {
		rg.Apply(resources.ResourceGroupCfg_Basic(rgCfg), acc.Exists())
	})

	cfg := datafactory.NewDataFactoryCfg(rgCfg)
	factory := ws.ResourceFor(cfg)

	It("creates a data factory, then imports it with no drift", func() {
		factory.Apply(datafactory.DataFactoryCfg_Basic(cfg), acc.Exists()).ImportVerify()
	})

	It("disables public network access in place", func() {
		// A plain enum toggle (public_network_access -> Disabled) that ARM echoes back
		// verbatim; each Apply re-plans for drift, so the in-place change round-trips.
		factory.Apply(datafactory.DataFactoryCfg_Update(cfg)).ImportVerify()
	})
})
