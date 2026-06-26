package storage

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/native/generated"
)

// BlobServiceCfg builds azapi_storage_account_blob_service acceptance-test
// configurations. blobServices is the singleton "default" child of a storage
// account: construct it with NewBlobServiceCfg, passing the parent storage-account
// config. It holds that StorageAccountCfg and renders its IDRef (e.g.
// "azapi_storage_account.sa.id") as the blob service's storage_account_id, so the
// dependency is config-to-config and the held parent is available for any further
// derived fields. The returned HCL is a template rendered by the acceptance framework.
type BlobServiceCfg struct {
	generated.ResourceConfigBase
	storageAccount StorageAccountCfg
}

// NewBlobServiceCfg builds a blob-service config depending on the parent storage-account
// config: the blob service holds it and references its IDRef as storage_account_id. The
// label is optional — omit it for the single-instance default ("test"), or pass an
// explicit label when a scope holds more than one. The resource type is read from the
// StorageAccountBlobService descriptor.
func NewBlobServiceCfg(storageAccount StorageAccountCfg, label ...string) BlobServiceCfg {
	return BlobServiceCfg{
		ResourceConfigBase: generated.NewResourceConfigBase(StorageAccountBlobService.Name, label...),
		storageAccount:     storageAccount,
	}
}

// Basic is the singleton default blob service with change feed enabled.
func (r BlobServiceCfg) Basic() string { return r.config("default", r.changeFeed(true)) }

// Complete toggles versioning on the default blob service — a fuller configuration
// used to exercise an in-place update from Basic.
func (r BlobServiceCfg) Complete() string { return r.config("default", r.versioning(true)) }

// WithChangeFeed sets properties.change_feed.enabled.
func (r BlobServiceCfg) WithChangeFeed(enabled bool) string {
	return r.config("default", r.changeFeed(enabled))
}

// Named builds a blob service with a non-default name, for validation scenarios
// (the schema requires the name "default").
func (r BlobServiceCfg) Named(name string) string { return r.config(name, r.versioning(true)) }

func (r BlobServiceCfg) config(name, body string) string {
	return fmt.Sprintf(`
resource %q %q {
  name               = %q
  storage_account_id = %s%s
}
`, r.ResourceType(), r.ResourceLabel(), name, r.storageAccount.IDRef(), body)
}

// changeFeed is a BlobServiceCfg properties fragment toggling change feed. Defined as
// a method (not a package function) so the name stays scoped to BlobServiceCfg and
// cannot collide with another resource's helpers in this shared package.
func (r BlobServiceCfg) changeFeed(enabled bool) string {
	return fmt.Sprintf(`
  properties = {
    change_feed = {
      enabled = %t
    }
  }`, enabled)
}

// versioning is a BlobServiceCfg properties fragment toggling versioning. Defined as
// a method so its name stays scoped to BlobServiceCfg (see changeFeed).
func (r BlobServiceCfg) versioning(enabled bool) string {
	return fmt.Sprintf(`
  properties = {
    is_versioning_enabled = %t
  }`, enabled)
}
