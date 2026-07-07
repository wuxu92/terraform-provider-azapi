package web_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/web"
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

	// Keep the plan and site in separate nested scopes. Scoped teardown removes all
	// configs owned by that scope before re-applying; nesting makes Ginkgo destroy
	// the site first, then the server farm it uses.
	Describe("an App Service plan", Ordered, func() {
		planScope := ws.Scope()

		farmCfg := web.NewWebServerFarmCfg(rgCfg, "plan")
		farm := planScope.ResourceFor(farmCfg)

		BeforeAll(func() {
			farm.Apply(web.WebServerFarmCfg_Basic(farmCfg), acc.Exists()).ImportVerify()
		})

		Describe("an App Service site", Ordered, func() {
			siteScope := planScope.Scope()

			cfg := web.NewWebSiteCfg(rgCfg, farmCfg)
			site := siteScope.ResourceFor(cfg)

			BeforeAll(func() {
				site.Apply(web.WebSiteCfg_Basic(cfg), acc.Exists()).ImportVerify()
			})

			It("updates common site properties in place", func() {
				site.Apply(web.WebSiteCfg_Complete(cfg)).ImportVerify()
			})

			It("re-updates every mutable site property in place", func() {
				// Complete_update flips each in-place-updatable value set by Complete;
				// each Apply re-plans for drift, so the update round-trips through the
				// config/web read without per-value assertions.
				site.Apply(web.WebSiteCfg_Complete_update(cfg)).ImportVerify()
			})
		})
	})
})
