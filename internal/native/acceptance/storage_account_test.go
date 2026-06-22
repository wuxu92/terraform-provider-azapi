package nativeacc

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("azapi_storage_account", func() {
	spec := NewSpec("azapi_storage_account", "Microsoft.Storage/storageAccounts", "2025-01-01")

	// The generated schema models sku/properties as nested *attributes*, so HCL
	// uses assignment syntax (`sku = { ... }`), not blocks.

	It("creates a basic account and imports it", func() {
		spec.Run(
			Body(`
location = "{{.Location}}"
kind     = "StorageV2"
sku = {
  name = "Standard_LRS"
}
`).Check(
				Exists(),
				Key("kind").HasValue("StorageV2"),
				Key("sku.name").HasValue("Standard_LRS"),
			),
			ImportStep(),
		)
	})

	It("updates the access tier in place", func() {
		spec.Run(
			Body(`
location = "{{.Location}}"
kind     = "StorageV2"
sku = {
  name = "Standard_LRS"
}
properties = {
  access_tier = "Hot"
}
`).Check(
				Exists(),
				Key("properties.access_tier").HasValue("Hot"),
			),
			Body(`
location = "{{.Location}}"
kind     = "StorageV2"
sku = {
  name = "Standard_LRS"
}
properties = {
  access_tier = "Cool"
}
`).Check(
				Exists(),
				Key("properties.access_tier").HasValue("Cool"),
			),
		)
	})

	It("requires replacement when migrating between zonal and non-zonal SKUs", func() {
		// Exercises the storage overlay's azwise.CheckForceNew SKU zone-migration
		// rule: Standard_LRS -> Standard_ZRS forces a replace.
		spec.Run(
			Body(`
location = "{{.Location}}"
kind     = "StorageV2"
sku = {
  name = "Standard_LRS"
}
`).Check(
				Exists(),
				Key("sku.name").HasValue("Standard_LRS"),
			),
			Body(`
location = "{{.Location}}"
kind     = "StorageV2"
sku = {
  name = "Standard_ZRS"
}
`).Check(
				Exists(),
				Key("sku.name").HasValue("Standard_ZRS"),
			),
		)
	})

	It("applies a verified default when the field is omitted", func() {
		// minimum_tls_version has an azwise verified default of TLS1_2; omitting
		// it should surface that default in state (Optional+Computed+Default).
		spec.Run(
			Body(`
location = "{{.Location}}"
kind     = "StorageV2"
sku = {
  name = "Standard_LRS"
}
`).Check(
				Exists(),
				Key("properties.minimum_tls_version").HasValue("TLS1_2"),
			),
		)
	})
})

var _ = Describe("framework", func() {
	It("builds a spec without panicking", func() {
		Expect(NewSpec("azapi_storage_account", "Microsoft.Storage/storageAccounts", "2025-01-01")).NotTo(BeNil())
	})
})
