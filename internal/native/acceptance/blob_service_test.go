package nativeacc

import (
	"github.com/Azure/terraform-provider-azapi/internal/acceptance"
	. "github.com/onsi/ginkgo/v2"
)

// blobServiceParent provisions a resource group and a storage account. The
// storage account is labelled "parent" so the framework wires
// storage_account_id = azapi_resource.parent.id for the child resource.
const blobServiceParent = `
resource "azapi_resource" "rg" {
  type      = "Microsoft.Resources/resourceGroups@2021-04-01"
  name      = "acctest-rg-{{.RandomInteger}}"
  parent_id = "/subscriptions/{{.SubscriptionID}}"
  location  = "{{.Location}}"
}

resource "azapi_resource" "parent" {
  type      = "Microsoft.Storage/storageAccounts@2025-01-01"
  name      = "acctestsa{{.RandomString}}"
  parent_id = azapi_resource.rg.id
  location  = "{{.Location}}"
  body = {
    sku  = { name = "Standard_LRS" }
    kind = "StorageV2"
  }
}
`

var _ = Describe("azapi_storage_account_blob_service", func() {
	// blobServices is a singleton child whose ARM name is always "default", so the
	// name is pinned (the OneOf("default") envelope validator rejects anything else).
	spec := NewSpec("azapi_storage_account_blob_service",
		"Microsoft.Storage/storageAccounts/blobServices", "2025-01-01").
		WithParent(blobServiceParent).
		WithName(func(acceptance.TestData) string { return "default" })

	It("creates the singleton blob service and imports it", func() {
		spec.Run(
			Body(`
properties = {
  change_feed = {
    enabled = true
  }
}
`).Check(
				Exists(),
				Key("properties.change_feed.enabled").HasValue("true"),
			),
			ImportStep(),
		)
	})

	It("rejects a name other than \"default\"", func() {
		spec.WithName(func(acceptance.TestData) string { return "notdefault" }).Run(
			Body(`
properties = {
  is_versioning_enabled = true
}
`).ExpectError(`blob service name must be "default"`),
		)
	})
})
