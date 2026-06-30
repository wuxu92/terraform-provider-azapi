package web_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/generated/resources"
	"github.com/Azure/terraform-provider-azapi/internal/native/generated/web"
)

// Azure Web exercises Microsoft.Web native resources with a native server farm
// dependency; acceptance configs must not fall back to generic azapi_resource for
// dependencies that native mode does not support yet.
var _ = Describe("Azure Web", Ordered, func() {
	ws := acc.NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	rgCfg := resources.NewResourceGroupCfg("rg")
	rg := ws.ResourceFor(rgCfg)
	BeforeAll(func() { rg.Apply(resources.ResourceGroupCfg_Basic(rgCfg), acc.Exists()) })

	Describe("an App Service site", Ordered, func() {
		scope := ws.Scope()
		AfterAll(scope.Teardown)

		farmCfg := web.NewWebServerFarmCfg(rgCfg, "plan")
		farm := scope.ResourceFor(farmCfg)
		cfg := web.NewWebSiteCfg(rgCfg, farmCfg)
		site := scope.ResourceFor(cfg)

		BeforeAll(func() {
			scope.ApplyAll(
				farm.Stage(web.WebServerFarmCfg_Basic(farmCfg), acc.Exists()),
				site.Stage(web.WebSiteCfg_Basic(cfg), acc.Exists()),
			)
			farm.ImportVerify()
			site.ImportVerify()
		})

		It("updates common site properties in place", func() {
			site.Apply(web.WebSiteCfg_Complete(cfg)).ImportVerify()
		})
	})
})
