package resources_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// Azure Resource Group exercises a subscription-scoped top-level resource. It has no
// base dependencies, so every scenario lives in the workspace root scope. The test
// config comes from the resources.ResourceGroupCfg builder defined alongside the
// generated resource in this package, which also vends the matching acceptance handle.
var _ = Describe("Azure Resource Group", Ordered, func() {
	ws := acc.NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	cfg := resources.NewResourceGroupCfg()
	rg := ws.ResourceFor(cfg)

	It("creates a resource group, then imports it with no drift", func() {
		rg.Apply(resources.ResourceGroupCfg_Basic(cfg), acc.Exists()).ImportVerify()
	})

	It("reads the resource group back through its data source", func() {
		// The data source reuses the resource read path (compose id, GET, mapper
		// flatten) against a schema converted from the generated resource schema. Its
		// state must reproduce the created group: same id, name, and a set location.
		// The data block is declared before Apply so the create apply reads it.
		ds := ws.DataSourceUnderTest(resources.NewResourceGroupDataCfg())
		rg.Apply(resources.ResourceGroupCfg_Basic(cfg), acc.Exists())
		ds.Check(
			acc.Key("id").HasValue(rg.IDValue()),
			acc.Key("name").HasValue(rg.AttrValue("name")),
			acc.Key("location").HasValue(rg.AttrValue("location")),
		)
	})

	It("rejects a name ending with a period", func() {
		invalid := resources.NewResourceGroupCfg("invalid")
		ws.ResourceFor(invalid).ApplyExpectError(
			resources.ResourceGroupCfg_Named{ResourceGroupCfg: invalid, Name: "acctest-rg-{{.RandomInteger}}."},
			`may not end with a period`,
		)
	})
})
