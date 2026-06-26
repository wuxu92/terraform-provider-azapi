package resources_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/generated/resources"
)

// Azure Resource Group exercises a subscription-scoped top-level resource. It has no
// base dependencies, so every scenario lives in the workspace root scope. The test
// config comes from the resources.ResourceGroupCfg builder defined alongside the
// generated resource in this package, which also vends the matching acceptance handle.
var _ = Describe("Azure Resource Group", Ordered, func() {
	ws := acc.NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	cfg := resources.NewResourceGroupCfg("test")
	rg := ws.ResourceFor(cfg)

	It("creates a resource group, then imports it with no drift", func() {
		rg.Apply(cfg.Basic(), acc.Exists()).ImportVerify()
	})

	It("rejects a name ending with a period", func() {
		invalid := resources.NewResourceGroupCfg("invalid")
		ws.ResourceFor(invalid).ApplyExpectError(
			invalid.Named("acctest-rg-{{.RandomInteger}}."),
			`may not end with a period`,
		)
	})
})
