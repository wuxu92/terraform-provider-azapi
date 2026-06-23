package nativeacc

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
)

// Azure Storage groups the storage-account scenarios. The shared base is a single
// resource group; the storage account is the resource-under-test, created once and
// mutated by the ordered scenarios below, then destroyed with the workspace.
var _ = Describe("Azure Storage", Ordered, func() {
	ws := NewWorkspace(`
resource "azapi_resource" "rg" {
  type      = "Microsoft.Resources/resourceGroups@2021-04-01"
  name      = "acctest-rg-{{.RandomInteger}}"
  parent_id = "/subscriptions/{{.SubscriptionID}}"
  location  = "{{.Location}}"
}
`)
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	sa := ws.Resource("azapi_storage_account", "test")

	It("creates a basic Standard_LRS StorageV2 account", func() {
		sa.Apply(storageAccountConfig("Standard_LRS", ""),
			Exists(),
			Key("kind").HasValue("StorageV2"),
			Key("sku.name").HasValue("Standard_LRS"),
			// minimum_tls_version has an azwise verified default of TLS1_2; omitting
			// it surfaces that default in state (Optional+Computed+Default).
			Key("properties.minimum_tls_version").HasValue("TLS1_2"),
		)
	})

	It("imports cleanly with no drift", func() {
		sa.ImportVerify()
	})

	It("sets the access tier to Hot", func() {
		sa.Apply(storageAccountConfig("Standard_LRS", `
  properties = {
    access_tier = "Hot"
  }`),
			Exists(),
			Key("properties.access_tier").HasValue("Hot"),
		)
	})

	It("updates the access tier to Cool in place", func() {
		sa.Apply(storageAccountConfig("Standard_LRS", `
  properties = {
    access_tier = "Cool"
  }`),
			Exists(),
			Key("properties.access_tier").HasValue("Cool"),
		)
	})

	It("replaces the account when migrating Standard_LRS to Standard_ZRS", func() {
		// Exercises the storage overlay's azwise.CheckForceNew SKU zone-migration
		// rule: Standard_LRS -> Standard_ZRS forces a replace.
		sa.Apply(storageAccountConfig("Standard_ZRS", ""),
			Exists(),
			Key("sku.name").HasValue("Standard_ZRS"),
		)
	})
})

// storageAccountConfig renders a StorageV2 account block with the given SKU name
// and an optional extra fragment (e.g. a `properties = { ... }` block, leading
// newline + 2-space indent). The name is stable for the workspace lifetime so
// successive Applies update the same account in place.
func storageAccountConfig(sku, extra string) string {
	return fmt.Sprintf(`
resource "azapi_storage_account" "test" {
  name              = "acctestsa{{.RandomString}}"
  resource_group_id = azapi_resource.rg.id
  location          = "{{.Location}}"
  kind              = "StorageV2"
  sku = {
    name = %q
  }%s
}
`, sku, extra)
}
