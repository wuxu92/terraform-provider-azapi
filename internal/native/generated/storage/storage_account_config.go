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

// Basic is a StorageV2 Standard_LRS account with no caller-set properties. The
// empty properties block is deliberate, not noise: it makes properties a present
// (non-null) object so the framework descends into it and applies azwise's nested
// computed defaults (e.g. minimum_tls_version=TLS1_2). Omitting the block entirely
// leaves properties null, and tftypes.Transform never reaches the children, so
// their defaults never reach the ARM payload.
func (r StorageAccount) Basic() string { return r.config("Standard_LRS", "\n  properties = {}") }

// Complete sets as many optional properties as one valid in-place configuration
// covers — scalars across the access/security/network surface plus the network,
// SAS and key-policy blocks — to exercise the whole resource on an in-place update
// from Basic. ForceNew knobs (is_hns_enabled, dns_endpoint_type, encryption, SKU)
// are deliberately excluded so the update chain stays in-place; the SKU replace
// lives in WithSKU.
func (r StorageAccount) Complete() string { return r.config("Standard_LRS", r.completeProps()) }

// WithSKU overrides the SKU, e.g. Standard_ZRS to exercise the storage overlay's
// zone-migration (Standard_LRS -> Standard_ZRS) ForceNew rule. Like Basic it carries
// a present (empty) properties block so the replacement account still receives
// azwise's nested computed defaults.
func (r StorageAccount) WithSKU(sku string) string { return r.config(sku, "\n  properties = {}") }

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

// completeProps is the StorageAccount properties fragment for Complete: a broad set
// of in-place-updatable optional settings plus a few nested policy blocks. Defined
// as a method (not a package function) so the name stays scoped to StorageAccount
// and cannot collide with another resource's helpers in this shared package.
func (r StorageAccount) completeProps() string {
	return `
  properties = {
    access_tier                      = "Cool"
    allow_blob_public_access         = false
    allow_cross_tenant_replication   = false
    allow_shared_key_access          = true
    allowed_copy_scope               = "AAD"
    default_to_o_auth_authentication = false
    large_file_shares_state          = "Enabled"
    minimum_tls_version              = "TLS1_2"
    public_network_access            = "Enabled"
    supports_https_traffic_only      = true
    network_acls = {
      bypass         = "AzureServices"
      default_action = "Allow"
    }
    sas_policy = {
      expiration_action     = "Log"
      sas_expiration_period = "1.00:00:00"
    }
    key_policy = {
      key_expiration_period_in_days = 7
    }
  }`
}
