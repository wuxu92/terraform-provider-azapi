package storage

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
)

// BlobServiceCfg carries the Terraform address metadata and storage-account
// dependency for azapi_storage_account_blob_service acceptance-test scenarios.
// blobServices is the singleton "default" child of a storage account: construct this
// with NewBlobServiceCfg, then wrap it in a scenario type when applying. The parent
// StorageAccountCfg is held so every scenario renders the same storage_account_id
// reference (e.g. "azapi_storage_account.sa.id").
type BlobServiceCfg struct {
	config.ResourceConfigBase
	storageAccount StorageAccountCfg
}

// NewBlobServiceCfg builds a blob-service config depending on the parent storage-account
// config: the blob service holds it and references its IDRef as storage_account_id. The
// label is optional — omit it for the single-instance default ("test"), or pass an
// explicit label when a scope holds more than one. The resource type is read from the
// StorageAccountBlobService descriptor.
func NewBlobServiceCfg(storageAccount StorageAccountCfg, label ...string) BlobServiceCfg {
	return BlobServiceCfg{
		ResourceConfigBase: config.NewResourceConfigBase(StorageAccountBlobService.Name, label...),
		storageAccount:     storageAccount,
	}
}

// BlobServiceCfg_Basic is the singleton default blob service with change feed enabled.
type BlobServiceCfg_Basic BlobServiceCfg

func (r BlobServiceCfg_Basic) Config() string {
	base := BlobServiceCfg(r)
	return base.config("default", base.changeFeed(true))
}

// BlobServiceCfg_Complete sets a broad valid blob-service configuration to exercise
// the resource surface in one in-place update from Basic. It includes the independent
// service version, CORS, delete-retention, container-retention, change-feed,
// versioning, last-access tracking and restore-policy settings. Read-only children of
// last_access_time_tracking_policy are deliberately omitted; enabling the policy makes
// Azure populate them and covers the non-null-state plan modifier without sending
// server-controlled fields.
type BlobServiceCfg_Complete BlobServiceCfg

func (r BlobServiceCfg_Complete) Config() string {
	base := BlobServiceCfg(r)
	return base.config("default", base.completeProps())
}

// BlobServiceCfg_Complete_update mutates every in-place-updatable property set by
// BlobServiceCfg_Complete to a different valid value, so applying Complete then
// Complete_update proves each survives an in-place update (Update -> Read -> empty
// plan). The point-in-time-restore prerequisites are held on purpose: restore_policy
// requires versioning, change feed and blob soft delete enabled, so is_versioning_enabled,
// change_feed.enabled, delete_retention_policy.enabled and restore_policy.enabled stay
// true; allow_permanent_delete stays false on both retention policies because it
// conflicts with versioning/restore; last_access_time_tracking_policy.enable stays true
// because the block's other fields are server-populated.
type BlobServiceCfg_Complete_update BlobServiceCfg

func (r BlobServiceCfg_Complete_update) Config() string {
	base := BlobServiceCfg(r)
	return base.config("default", base.completeUpdateProps())
}

// BlobServiceCfg_ChangeFeed sets properties.change_feed.enabled.
type BlobServiceCfg_ChangeFeed struct {
	BlobServiceCfg
	Enabled bool
}

func (r BlobServiceCfg_ChangeFeed) Config() string {
	return r.config("default", r.changeFeed(r.Enabled))
}

// BlobServiceCfg_Named renders a blob service with a non-default name, for validation
// scenarios (the schema requires the name "default").
type BlobServiceCfg_Named struct {
	BlobServiceCfg
	Name string
}

func (r BlobServiceCfg_Named) Config() string {
	return r.config(r.Name, r.versioning(true))
}

func (r BlobServiceCfg) config(name, body string) string {
	return r.RenderConfig(config.ConfigEnvelope{Name: name,
		ParentAttr: "storage_account_id",
		ParentRef:  r.storageAccount.IDRef(),
		Body:       body})
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

// completeProps is the BlobServiceCfg properties fragment for Complete. Values stay
// inside the generated validators and Azure semantic constraints: point-in-time
// restore requires versioning, change feed and blob soft delete to be enabled, and
// restore days must be lower than delete-retention days.
func (r BlobServiceCfg) completeProps() string {
	return `
  properties = {
    automatic_snapshot_policy_enabled = false
    change_feed = {
      enabled           = true
      retention_in_days = 7
    }
    container_delete_retention_policy = {
      allow_permanent_delete = false
      days                   = 14
      enabled                = true
    }
    cors = {
      cors_rules = [{
        allowed_headers    = ["x-ms-*", "content-type"]
        allowed_methods    = ["GET", "PUT"]
        allowed_origins    = ["https://example.com"]
        exposed_headers    = ["x-ms-*"]
        max_age_in_seconds = 3600
      }]
    }
    default_service_version = "2023-11-03"
    delete_retention_policy = {
      allow_permanent_delete = false
      days                   = 14
      enabled                = true
    }
    is_versioning_enabled = true
    last_access_time_tracking_policy = {
      enable = true
    }
    restore_policy = {
      days    = 7
      enabled = true
    }
  }`
}

// completeUpdateProps is the BlobServiceCfg properties fragment for Complete_update: it
// flips each in-place-updatable value from completeProps while keeping restore days
// below delete-retention days and holding the point-in-time-restore prerequisites (see
// BlobServiceCfg_Complete_update).
func (r BlobServiceCfg) completeUpdateProps() string {
	return `
  properties = {
    automatic_snapshot_policy_enabled = true
    change_feed = {
      enabled           = true // held: point-in-time restore requires change feed enabled
      retention_in_days = 14
    }
    container_delete_retention_policy = {
      allow_permanent_delete = false // held: conflicts with versioning/point-in-time restore
      days                   = 21
      enabled                = true
    }
    cors = {
      cors_rules = [{
        allowed_headers    = ["x-ms-meta-*"]
        allowed_methods    = ["GET", "POST", "HEAD"]
        allowed_origins    = ["https://updated.example.com"]
        exposed_headers    = ["x-ms-meta-*"]
        max_age_in_seconds = 7200
      }]
    }
    default_service_version = "2021-12-02"
    delete_retention_policy = {
      allow_permanent_delete = false // held: conflicts with versioning/point-in-time restore
      days                   = 21
      enabled                = true // held: point-in-time restore requires blob soft delete
    }
    is_versioning_enabled = true // held: point-in-time restore requires versioning
    last_access_time_tracking_policy = {
      enable = true // held: only writable field; granularity/name/blob types are server-populated
    }
    restore_policy = {
      enabled = false
    }
  }`
}
