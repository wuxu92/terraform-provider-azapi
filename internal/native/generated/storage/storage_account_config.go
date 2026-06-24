package storage

import "fmt"

// StorageAccount builds azapi_storage_account acceptance-test configurations.
//
// Label is the Terraform state label; ResourceGroupIDRef is the injected parent
// reference — the Terraform address of a resource group applied as a base, e.g. the
// acceptance Resource's IDRef "azapi_resource_group.rg.id". The returned HCL is a
// template rendered by the acceptance framework ({{.RandomString}}, {{.Location}}).
type StorageAccount struct {
	Label              string
	ResourceGroupIDRef string
}

// Basic is a StorageV2 Standard_LRS account with no optional properties.
func (r StorageAccount) Basic() string { return r.config("Standard_LRS", "") }

// Complete is a StorageV2 Standard_LRS account with an explicit access tier — a
// fuller configuration used to exercise an in-place update from Basic.
func (r StorageAccount) Complete() string { return r.config("Standard_LRS", r.accessTier("Cool")) }

// WithSKU overrides the SKU, e.g. Standard_ZRS to exercise the storage overlay's
// zone-migration (Standard_LRS -> Standard_ZRS) ForceNew rule.
func (r StorageAccount) WithSKU(sku string) string { return r.config(sku, "") }

func (r StorageAccount) config(sku, extra string) string {
	return fmt.Sprintf(`
resource "azapi_storage_account" %q {
  name              = "acctestsa{{.RandomString}}"
  resource_group_id = %s
  location          = "{{.Location}}"
  kind              = "StorageV2"
  sku = {
    name = %q
  }%s
}
`, r.Label, r.ResourceGroupIDRef, sku, extra)
}

// accessTier is a StorageAccount properties fragment setting access_tier. Defined as
// a method (not a package function) so the name stays scoped to StorageAccount and
// cannot collide with another resource's helpers in this shared package.
func (r StorageAccount) accessTier(tier string) string {
	return fmt.Sprintf(`
  properties = {
    access_tier = %q
  }`, tier)
}

// ResourceType is the Terraform type this builder configures; with ResourceLabel it
// satisfies acceptance.ResourceConfig, so a scope can vend the matching handle
// directly — acct.ResourceFor(cfg) — without restating the literal type and label.
func (r StorageAccount) ResourceType() string { return "azapi_storage_account" }

// ResourceLabel is the Terraform state label (see ResourceType).
func (r StorageAccount) ResourceLabel() string { return r.Label }
