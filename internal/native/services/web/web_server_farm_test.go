package web_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/web"
)

// Azure Web Server Farm exercises the native Microsoft.Web/serverfarms resource
// directly. Web-site acceptance may reuse this resource as a dependency, but the
// server farm still owns its own create/import/update coverage here.
var _ = Describe("Azure Web Server Farm", Ordered, func() {
	ws := acc.NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	rgCfg := resources.NewResourceGroupCfg("rg")
	rg := ws.ResourceFor(rgCfg)
	BeforeAll(func() { rg.Apply(resources.ResourceGroupCfg_Basic(rgCfg), acc.Exists()) })

	Describe("an App Service plan", Ordered, func() {
		scope := ws.Scope()

		cfg := web.NewWebServerFarmCfg(rgCfg)
		farm := scope.ResourceFor(cfg)

		BeforeAll(func() {
			farm.Apply(web.WebServerFarmCfg_Basic(cfg), acc.Exists()).ImportVerify()
		})

		It("updates common plan properties in place", func() {
			farm.Apply(web.WebServerFarmCfg_Complete(cfg))
		})
	})
})
