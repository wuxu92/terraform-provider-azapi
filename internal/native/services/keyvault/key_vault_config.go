package keyvault

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// KeyVaultCfg carries the Terraform address metadata and dependencies for
// azapi_key_vault acceptance-test scenarios. Construct it with NewKeyVaultCfg, then
// wrap it in a scenario type when applying. The parent ResourceGroupCfg is held so
// every scenario renders the same resource_group_id reference (e.g.
// "azapi_resource_group.rg.id"). The returned HCL is a template rendered by the
// acceptance framework ({{.RandomString}}, {{.Location}}).
//
// The vault's required tenant_id — and the access-policy object_id/tenant_id — resolve
// to the identity the provider runs as via a held ClientConfig data-source dependency
// (the azapi analogue of azurerm's data.azurerm_client_config.current). The scenario
// bodies only reference it by address (ClientConfig.RefOf); the data source itself is
// declared once at the workspace's root scope via Scope.DataSource, so a single
// azapi_client_config block is shared across the run instead of each vault config
// redeclaring it.
type KeyVaultCfg struct {
	config.ResourceConfigBase
	resourceGroup resources.ResourceGroupCfg
	clientConfig  config.ClientConfigData
}

// NewKeyVaultCfg builds a key-vault config depending on the parent resource-group
// config (e.g. the one applied as the base): the vault holds it and references its
// IDRef as resource_group_id. clientConfig is the shared azapi_client_config data
// source (config.ClientConfig) the vault reads for tenant/object ids; declare it once
// at the root scope with ws.DataSource(config.ClientConfig). The label is optional —
// omit it for the single-instance default ("test"), or pass an explicit label when a
// scope holds more than one. The resource type is read from the KeyVault descriptor.
func NewKeyVaultCfg(resourceGroup resources.ResourceGroupCfg, clientConfig config.ClientConfigData, label ...string) KeyVaultCfg {
	return KeyVaultCfg{
		ResourceConfigBase: config.NewResourceConfigBase(KeyVault.Name, label...),
		resourceGroup:      resourceGroup,
		clientConfig:       clientConfig,
	}
}

// KeyVaultCfg_Basic is a minimal standard vault: the required tenant_id and sku only.
// Purge protection is left to its false default so the vault is deletable at teardown
// — an enabled purge lock would pin the soft-deleted vault for the whole retention
// window and leak it across runs. Every other body field rides its azwise default, so
// the Optional+Computed attributes round-trip without drift.
type KeyVaultCfg_Basic KeyVaultCfg

func (r KeyVaultCfg_Basic) Config() string {
	kv := KeyVaultCfg(r)
	return kv.config(fmt.Sprintf(`
  properties = {
    tenant_id = %[1]s
    sku = {
      name   = "standard"
      family = "A"
    }
  }`, kv.clientConfig.RefOf("tenant_id")))
}

// KeyVaultCfg_Complete exercises the vault's in-place-updatable surface: an access
// policy for the running identity, an explicit network ACL, public network access, and
// tags. tenant_id/object_id come from the client-config data source so the access
// policy's tenant matches the vault's (an ARM invariant). Applied after Basic, it
// proves the add path (Update -> Read -> empty plan).
type KeyVaultCfg_Complete KeyVaultCfg

func (r KeyVaultCfg_Complete) Config() string {
	kv := KeyVaultCfg(r)
	return kv.config(fmt.Sprintf(`
  properties = {
    tenant_id             = %[1]s
    public_network_access = "Enabled"
    sku = {
      name   = "standard"
      family = "A"
    }
    network_acls = {
      default_action = "Deny"
      bypass         = "AzureServices"
    }
    access_policies = [
      {
        tenant_id = %[1]s
        object_id = %[2]s
        permissions = {
          keys    = ["Get", "List"]
          secrets = ["Get", "List"]
        }
      }
    ]
  }
  tags = {
    environment = "test"
  }`, kv.clientConfig.RefOf("tenant_id"), kv.clientConfig.RefOf("object_id")))
}

// KeyVaultCfg_Complete_update flips the tag value and widens the access-policy
// permissions set by Complete, so applying Complete then Complete_update proves both
// survive an in-place update (Update -> Read -> empty plan).
type KeyVaultCfg_Complete_update KeyVaultCfg

func (r KeyVaultCfg_Complete_update) Config() string {
	kv := KeyVaultCfg(r)
	return kv.config(fmt.Sprintf(`
  properties = {
    tenant_id             = %[1]s
    public_network_access = "Enabled"
    sku = {
      name   = "standard"
      family = "A"
    }
    network_acls = {
      default_action = "Deny"
      bypass         = "None"
      ip_rules       = []
      virtual_network_rules = []
    }
    access_policies = [
      {
        tenant_id = %[1]s
        object_id = %[2]s
        permissions = {
          keys    = ["Get", "List", "Create", "Delete"]
          secrets = ["Get", "List", "Set", "Delete"]
        }
      }
    ]
  }
  tags = {
    environment = "prod"
  }`, kv.clientConfig.RefOf("tenant_id"), kv.clientConfig.RefOf("object_id")))
}

func (r KeyVaultCfg) config(body string) string {
	return r.RenderConfig(config.ConfigEnvelope{
		Name:       "accazapikv{{.RandomString}}",
		ParentAttr: "resource_group_id",
		ParentRef:  r.resourceGroup.IDRef(),
		Location:   true,
		Body:       body,
	})
}
