package authorization_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/authorization"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
)

// Azure role assignments are exercised as a subscription-scope native static resource
// assigning the built-in Reader role to the current provider identity. Creating any live
// role assignment requires unrestricted Microsoft.Authorization/roleAssignments/write,
// so the scenario is gated before apply and skips cleanly for Contributor-class runs.
var _ = Describe("Azure Role Assignment", Ordered, func() {
	ws := acc.NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)
	BeforeAll(func() {
		ws.DataSource(config.ClientConfig)
	})

	cfg := authorization.NewRoleAssignmentCfg(config.ClientConfig)
	ra := ws.ResourceFor(cfg)

	It("assigns a built-in role at subscription scope", func() {
		acc.SkipIfNoRoleAssignmentWrite()
		ra.Apply(authorization.RoleAssignmentCfg_Basic(cfg), acc.Exists()).ImportVerify()
	})
})
