package nativeacc

import (
	. "github.com/onsi/ginkgo/v2"
)

// Azure Resource Group groups the resourceGroups scenarios. resourceGroups is a
// subscription-scoped top-level resource, so there is no provisioned parent: the
// workspace has no base, and the envelope subscription_id is the subscription scope.
var _ = Describe("Azure Resource Group", Ordered, func() {
	ws := NewWorkspace("")
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	rg := ws.Resource("azapi_resource_group", "test")

	It("creates a resource group", func() {
		rg.Apply(`
resource "azapi_resource_group" "test" {
  name            = "acctest-rg-{{.RandomInteger}}"
  subscription_id = "/subscriptions/{{.SubscriptionID}}"
  location        = "{{.Location}}"
}
`,
			Exists(),
		)
	})

	It("imports cleanly with no drift", func() {
		rg.ImportVerify()
	})

	It("rejects a name ending with a period", func() {
		ws.Resource("azapi_resource_group", "invalid").ApplyExpectError(`
resource "azapi_resource_group" "invalid" {
  name            = "acctest-rg-{{.RandomInteger}}."
  subscription_id = "/subscriptions/{{.SubscriptionID}}"
  location        = "{{.Location}}"
}
`, `may not end with a period`)
	})
})
