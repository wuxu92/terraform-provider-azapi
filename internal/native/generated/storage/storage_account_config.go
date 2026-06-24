package storage

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/native/generated"
)

// StorageAccount builds azapi_storage_account acceptance-test configurations.
// Construct it with NewStorageAccount. ResourceGroupIDRef is the injected parent
// reference — the Terraform address of a resource group applied as a base, e.g.
// the acceptance Resource's IDRef "azapi_resource_group.rg.id". The returned HCL
// is a template rendered by the acceptance framework ({{.RandomString}},
// {{.Location}}).
type StorageAccount struct {
	generated.ResourceConfigBase
	ResourceGroupIDRef string
}

// NewStorageAccount builds a storage-account config with the given Terraform state
// label and parent resource-group address reference (resourceGroupIDRef, e.g.
// rg.IDRef()). The resource type is set from generated.TypeStorageAccount.
func NewStorageAccount(label, resourceGroupIDRef string) StorageAccount {
	return StorageAccount{
		ResourceConfigBase: generated.NewResourceConfigBase(generated.TypeStorageAccount, label),
		ResourceGroupIDRef: resourceGroupIDRef,
	}
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
resource %q %q {
  name              = "acctestsa{{.RandomString}}"
  resource_group_id = %s
  location          = "{{.Location}}"
  kind              = "StorageV2"
  sku = {
    name = %q
  }%s
}
`, r.ResourceType(), r.ResourceLabel(), r.ResourceGroupIDRef, sku, extra)
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
