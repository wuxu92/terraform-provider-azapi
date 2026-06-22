package nativeacc

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/acceptance"
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("azapi_resource_group", func() {
	// resourceGroups is a subscription-scoped top-level resource: the parent is the
	// subscription itself, so there is no parent resource to provision — the parent
	// template is dropped and the envelope subscription_id is set to the scope ID.
	spec := NewSpec("azapi_resource_group").
		WithParent("").
		WithParentRef(`"/subscriptions/{{.SubscriptionID}}"`).
		WithName(func(td acceptance.TestData) string {
			return fmt.Sprintf("acctest-rg-%d", td.RandomInteger)
		})

	It("creates a resource group and imports it", func() {
		spec.Run(
			Body(`
location = "{{.Location}}"
`).Check(
				Exists(),
			),
			ImportStep(),
		)
	})

	It("rejects a name ending with a period", func() {
		spec.WithName(func(td acceptance.TestData) string {
			return fmt.Sprintf("acctest-rg-%d.", td.RandomInteger)
		}).Run(
			Body(`
location = "{{.Location}}"
`).ExpectError(`may not end with a period`),
		)
	})
})
