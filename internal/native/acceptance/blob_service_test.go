package nativeacc

import (
	. "github.com/onsi/ginkgo/v2"
)

// Azure Storage Blob Service groups the blobServices scenarios. The shared base is
// a resource group plus a storage account; the blob service is the singleton child
// resource-under-test (ARM name always "default").
var _ = Describe("Azure Storage Blob Service", Ordered, func() {
	ws := NewWorkspace(`
resource "azapi_resource" "rg" {
  type      = "Microsoft.Resources/resourceGroups@2021-04-01"
  name      = "acctest-rg-{{.RandomInteger}}"
  parent_id = "/subscriptions/{{.SubscriptionID}}"
  location  = "{{.Location}}"
}

resource "azapi_resource" "sa" {
  type      = "Microsoft.Storage/storageAccounts@2025-06-01"
  name      = "acctestsa{{.RandomString}}"
  parent_id = azapi_resource.rg.id
  location  = "{{.Location}}"
  body = {
    sku  = { name = "Standard_LRS" }
    kind = "StorageV2"
  }
}
`)
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	// blobServices is a singleton child whose ARM name is always "default" (the
	// OneOf("default") envelope validator rejects anything else).
	blob := ws.Resource("azapi_storage_account_blob_service", "test")

	It("enables change feed", func() {
		blob.Apply(`
resource "azapi_storage_account_blob_service" "test" {
  name               = "default"
  storage_account_id = azapi_resource.sa.id
  properties = {
    change_feed = {
      enabled = true
    }
  }
}
`,
			Exists(),
			Key("properties.change_feed.enabled").HasValue("true"),
		)
	})

	It("imports cleanly with no drift", func() {
		blob.ImportVerify()
	})

	It("disables change feed in place", func() {
		blob.Apply(`
resource "azapi_storage_account_blob_service" "test" {
  name               = "default"
  storage_account_id = azapi_resource.sa.id
  properties = {
    change_feed = {
      enabled = false
    }
  }
}
`,
			Exists(),
			Key("properties.change_feed.enabled").HasValue("false"),
		)
	})

	It("rejects a name other than \"default\"", func() {
		ws.Resource("azapi_storage_account_blob_service", "invalid").ApplyExpectError(`
resource "azapi_storage_account_blob_service" "invalid" {
  name               = "notdefault"
  storage_account_id = azapi_resource.sa.id
  properties = {
    is_versioning_enabled = true
  }
}
`, `name value must be one of`)
	})
})
