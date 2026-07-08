package authorization_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/authorization"
)

// Azure role definitions are exercised as a single custom role scoped at the
// subscription (the azapi analogue of azurerm's data.azurerm_subscription scope), so
// the suite needs no base resource. The role's GUID (its ForceNew name) is fixed for
// the run, so the three scenarios apply to one role in an in-place update chain:
// minimal defaults -> full surface -> narrowed permissions. The role is created at the
// workspace root scope and destroyed by ws.Destroy.
var _ = Describe("Azure Role Definition", Ordered, func() {
	ws := acc.NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	cfg := authorization.NewRoleDefinitionCfg()
	rd := ws.ResourceFor(cfg)

	It("creates a custom role with a defaulted type and assignable scope", func() {
		// Basic sets no type and no assignable_scopes: asserting type == CustomRole proves
		// the azwise default, and a set assignable_scopes[0] proves the BeforeCreate hook
		// injected the role's own (subscription) scope. The Apply re-plans for drift, so
		// the create round-trips.
		rd.Apply(authorization.RoleDefinitionCfg_Basic(cfg), acc.Exists(),
			acc.Key("properties.type").HasValue("CustomRole"),
			acc.Key("properties.assignable_scopes.0").IsSet()).ImportVerify()
	})

	It("updates to the complete surface in place", func() {
		// Complete adds a description and the full four-way permission set; the Apply re-
		// plans for drift, so the in-place update round-trips.
		rd.Apply(authorization.RoleDefinitionCfg_Complete(cfg), acc.Exists()).ImportVerify()
	})

	It("updates the description and permissions in place", func() {
		// Complete_update flips the description and narrows the permissions; the Apply re-
		// plans for drift, so the in-place update round-trips.
		rd.Apply(authorization.RoleDefinitionCfg_Complete_update(cfg)).ImportVerify()
	})
})
