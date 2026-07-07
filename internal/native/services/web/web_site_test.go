package web_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/managedidentity"
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
			// Native user-assigned identity dependency, created in this scope and
			// referenced by ARM ID in the site's root identity block.
			uaiCfg := managedidentity.NewUserAssignedIdentityCfg(rgCfg)
			uai := siteScope.ResourceFor(uaiCfg)

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

			It("assigns a user-assigned identity in place", func() {
				// One apply creates the identity and updates the site together; Terraform
				// orders the identity first from the site's user_assigned_identities ref.
				// Transitions the identity from SystemAssigned (Complete) to UserAssigned
				// in place; the drift plan proves it round-trips and the Key check pins it.
				siteScope.ApplyAll(
					uai.Stage(managedidentity.UserAssignedIdentityCfg_Basic(uaiCfg), acc.Exists()),
					site.Stage(web.WebSiteCfg_Identity{
						WebSiteCfg: cfg,
						Identity:   uaiCfg.IDRef(),
					}, acc.Key("identity.type").HasValue("UserAssigned")),
				)
			})
		})
	})
})
