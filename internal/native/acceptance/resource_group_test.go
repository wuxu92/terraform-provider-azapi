package nativeacc

import (
	. "github.com/onsi/ginkgo/v2"
)

// Azure Resource Group exercises a subscription-scoped top-level resource. It has no
// base dependencies, so every scenario lives in the workspace root scope.
var _ = Describe("Azure Resource Group", Ordered, func() {
	ws := NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	rg := ws.Resource("azapi_resource_group", "test")

	It("creates a resource group, then imports it with no drift", func() {
		rg.Apply(resourceGroupConfig("test"), Exists()).ImportVerify()
	})

	It("rejects a name ending with a period", func() {
		ws.Resource("azapi_resource_group", "invalid").ApplyExpectError(
			resourceGroupConfigNamed("invalid", "acctest-rg-{{.RandomInteger}}."),
			`may not end with a period`,
		)
	})
})
