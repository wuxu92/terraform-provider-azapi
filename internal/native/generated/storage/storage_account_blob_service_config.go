package storage

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/native/generated"
)

// BlobService builds azapi_storage_account_blob_service acceptance-test
// configurations. blobServices is the singleton "default" child of a storage
// account: construct it with NewBlobService. StorageAccountIDRef is the injected
// parent reference — the Terraform address of a storage account applied as a base,
// e.g. the acceptance Resource's IDRef "azapi_storage_account.sa.id". The returned
// HCL is a template rendered by the acceptance framework.
type BlobService struct {
	generated.ResourceConfigBase
	StorageAccountIDRef string
}

// NewBlobService builds a blob-service config with the given Terraform state label
// and parent storage-account address reference (storageAccountIDRef, e.g.
// sa.IDRef()). The resource type is set from generated.TypeStorageAccountBlobService.
func NewBlobService(label, storageAccountIDRef string) BlobService {
	return BlobService{
		ResourceConfigBase:  generated.NewResourceConfigBase(generated.TypeStorageAccountBlobService, label),
		StorageAccountIDRef: storageAccountIDRef,
	}
}

// Basic is the singleton default blob service with change feed enabled.
func (r BlobService) Basic() string { return r.config("default", r.changeFeed(true)) }

// Complete toggles versioning on the default blob service — a fuller configuration
// used to exercise an in-place update from Basic.
func (r BlobService) Complete() string { return r.config("default", r.versioning(true)) }

// WithChangeFeed sets properties.change_feed.enabled.
func (r BlobService) WithChangeFeed(enabled bool) string {
	return r.config("default", r.changeFeed(enabled))
}

// Named builds a blob service with a non-default name, for validation scenarios
// (the schema requires the name "default").
func (r BlobService) Named(name string) string { return r.config(name, r.versioning(true)) }

func (r BlobService) config(name, body string) string {
	return fmt.Sprintf(`
resource %q %q {
  name               = %q
  storage_account_id = %s%s
}
`, r.ResourceType(), r.ResourceLabel(), name, r.StorageAccountIDRef, body)
}

// changeFeed is a BlobService properties fragment toggling change feed. Defined as a
// method (not a package function) so the name stays scoped to BlobService and cannot
// collide with another resource's helpers in this shared package.
func (r BlobService) changeFeed(enabled bool) string {
	return fmt.Sprintf(`
  properties = {
    change_feed = {
      enabled = %t
    }
  }`, enabled)
}

// versioning is a BlobService properties fragment toggling versioning. Defined as a
// method so its name stays scoped to BlobService (see changeFeed).
func (r BlobService) versioning(enabled bool) string {
	return fmt.Sprintf(`
  properties = {
    is_versioning_enabled = %t
  }`, enabled)
}
