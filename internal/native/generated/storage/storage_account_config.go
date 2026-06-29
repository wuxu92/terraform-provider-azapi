package storage

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/native/generated"
	"github.com/Azure/terraform-provider-azapi/internal/native/generated/resources"
)

// StorageAccountCfg carries the Terraform address metadata and resource-group
// dependency for azapi_storage_account acceptance-test scenarios. Construct it with
// NewStorageAccountCfg, then wrap it in a scenario type when applying. The parent
// ResourceGroupCfg is held so every scenario renders the same resource_group_id
// reference (e.g. "azapi_resource_group.rg.id"). The returned HCL is a template
// rendered by the acceptance framework ({{.RandomString}}, {{.Location}}).
type StorageAccountCfg struct {
	generated.ResourceConfigBase
	resourceGroup resources.ResourceGroupCfg
}

// NewStorageAccountCfg builds a storage-account config depending on the parent
// resource-group config (e.g. the one applied as the base): the account holds it and
// references its IDRef as resource_group_id. The label is optional — omit it for the
// single-instance default ("test"), or pass an explicit label when a scope holds more
// than one. The resource type is read from the StorageAccount descriptor.
func NewStorageAccountCfg(resourceGroup resources.ResourceGroupCfg, label ...string) StorageAccountCfg {
	return StorageAccountCfg{
		ResourceConfigBase: generated.NewResourceConfigBase(StorageAccount.Name, label...),
		resourceGroup:      resourceGroup,
	}
}

// StorageAccountCfg_Basic is a StorageV2 Standard_LRS account with no caller-set
// properties. The empty properties block is deliberate, not noise: it makes
// properties a present (non-null) object so the framework descends into it and applies
// azwise's nested computed defaults (e.g. minimum_tls_version=TLS1_2). Omitting the
// block entirely leaves properties null, and tftypes.Transform never reaches the
// children, so their defaults never reach the ARM payload.
type StorageAccountCfg_Basic StorageAccountCfg

func (r StorageAccountCfg_Basic) Config() string {
	return StorageAccountCfg(r).config("Standard_LRS", "\n  properties = {}")
}

// StorageAccountCfg_Complete sets as many optional properties as one valid in-place
// configuration covers — scalars across the access/security/network surface plus the
// network, SAS and key-policy blocks — to exercise the whole resource on an in-place
// update from Basic. ForceNew knobs (is_hns_enabled, dns_endpoint_type, encryption,
// SKU) are deliberately excluded so the update chain stays in-place; the SKU replace
// lives in StorageAccountCfg_SKU.
type StorageAccountCfg_Complete StorageAccountCfg

func (r StorageAccountCfg_Complete) Config() string {
	base := StorageAccountCfg(r)
	return base.config("Standard_LRS", base.completeProps())
}

// StorageAccountCfg_SKU overrides the SKU, e.g. Standard_ZRS to exercise the storage
// overlay's zone-migration (Standard_LRS -> Standard_ZRS) ForceNew rule. Like Basic it
// carries a present (empty) properties block so the replacement account still receives
// azwise's nested computed defaults.
type StorageAccountCfg_SKU struct {
	StorageAccountCfg
	SKU string
}

func (r StorageAccountCfg_SKU) Config() string {
	return r.config(r.SKU, "\n  properties = {}")
}

func (r StorageAccountCfg) config(sku, extra string) string {
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
`, r.ResourceType(), r.ResourceLabel(), r.resourceGroup.IDRef(), sku, extra)
}

// completeProps is the StorageAccountCfg properties fragment for Complete: a broad
// set of in-place-updatable optional settings plus a few nested policy blocks.
// Defined as a method (not a package function) so the name stays scoped to
// StorageAccountCfg and cannot collide with another resource's helpers in this
// shared package.
func (r StorageAccountCfg) completeProps() string {
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
