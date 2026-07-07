package keyvault

import (
	"github.com/Azure/terraform-provider-azapi/internal/native/services"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// clientConfigData is the data source block every key-vault scenario prepends so
// the vault's required tenant_id — and the access-policy object_id/tenant_id —
// resolve to the identity the provider runs as. It is the azapi analogue of
// azurerm's `data.azurerm_client_config.current`. It is declared once per rendered
// config file, and each rendered file holds exactly one vault, so there is no
// duplicate-declaration collision across the working directory.
const clientConfigData = "\ndata \"azapi_client_config\" \"current\" {}\n"

// KeyVaultCfg carries the Terraform address metadata and resource-group dependency
// for azapi_key_vault acceptance-test scenarios. Construct it with NewKeyVaultCfg,
// then wrap it in a scenario type when applying. The parent ResourceGroupCfg is held
// so every scenario renders the same resource_group_id reference (e.g.
// "azapi_resource_group.rg.id"). The returned HCL is a template rendered by the
// acceptance framework ({{.RandomString}}, {{.Location}}).
type KeyVaultCfg struct {
	services.ResourceConfigBase
	resourceGroup resources.ResourceGroupCfg
}

// NewKeyVaultCfg builds a key-vault config depending on the parent resource-group
// config (e.g. the one applied as the base): the vault holds it and references its
// IDRef as resource_group_id. The label is optional — omit it for the single-instance
// default ("test"), or pass an explicit label when a scope holds more than one. The
// resource type is read from the KeyVault descriptor.
func NewKeyVaultCfg(resourceGroup resources.ResourceGroupCfg, label ...string) KeyVaultCfg {
	return KeyVaultCfg{
		ResourceConfigBase: services.NewResourceConfigBase(KeyVault.Name, label...),
		resourceGroup:      resourceGroup,
	}
}

// KeyVaultCfg_Basic is a minimal standard vault: the required tenant_id and sku only.
// Purge protection is left to its false default so the vault is deletable at teardown
// — an enabled purge lock would pin the soft-deleted vault for the whole retention
// window and leak it across runs. Every other body field rides its azwise default, so
// the Optional+Computed attributes round-trip without drift.
type KeyVaultCfg_Basic KeyVaultCfg

func (r KeyVaultCfg_Basic) Config() string {
	return KeyVaultCfg(r).config(`
  properties = {
    tenant_id = data.azapi_client_config.current.tenant_id
    sku = {
      name   = "standard"
      family = "A"
    }
  }`)
}

// KeyVaultCfg_Complete exercises the vault's in-place-updatable surface: an access
// policy for the running identity, an explicit network ACL, public network access, and
// tags. tenant_id/object_id come from the client-config data source so the access
// policy's tenant matches the vault's (an ARM invariant). Applied after Basic, it
// proves the add path (Update -> Read -> empty plan).
type KeyVaultCfg_Complete KeyVaultCfg

func (r KeyVaultCfg_Complete) Config() string {
	return KeyVaultCfg(r).config(`
  properties = {
    tenant_id             = data.azapi_client_config.current.tenant_id
    public_network_access = "Enabled"
    sku = {
      name   = "standard"
      family = "A"
    }
    network_acls = {
      default_action = "Allow"
      bypass         = "AzureServices"
    }
    access_policies = [
      {
        tenant_id = data.azapi_client_config.current.tenant_id
        object_id = data.azapi_client_config.current.object_id
        permissions = {
          keys    = ["Get", "List"]
          secrets = ["Get", "List"]
        }
      }
    ]
  }
  tags = {
    environment = "test"
  }`)
}

// KeyVaultCfg_Complete_update flips the tag value and widens the access-policy
// permissions set by Complete, so applying Complete then Complete_update proves both
// survive an in-place update (Update -> Read -> empty plan).
type KeyVaultCfg_Complete_update KeyVaultCfg

func (r KeyVaultCfg_Complete_update) Config() string {
	return KeyVaultCfg(r).config(`
  properties = {
    tenant_id             = data.azapi_client_config.current.tenant_id
    public_network_access = "Enabled"
    sku = {
      name   = "standard"
      family = "A"
    }
    network_acls = {
      default_action = "Allow"
      bypass         = "AzureServices"
    }
    access_policies = [
      {
        tenant_id = data.azapi_client_config.current.tenant_id
        object_id = data.azapi_client_config.current.object_id
        permissions = {
          keys    = ["Get", "List", "Create", "Delete"]
          secrets = ["Get", "List", "Set", "Delete"]
        }
      }
    ]
  }
  tags = {
    environment = "prod"
  }`)
}

func (r KeyVaultCfg) config(body string) string {
	return clientConfigData + r.RenderConfig(services.ConfigEnvelope{
		Name:       "acctestkv{{.RandomString}}",
		ParentAttr: "resource_group_id",
		ParentRef:  r.resourceGroup.IDRef(),
		Location:   true,
		Body:       body,
	})
}
