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
// window and leak it across runs. Every other body field rides its azwise default —
// including enable_rbac_authorization, which the resource defaults on, so Basic
// exercises the RBAC-on authorization model round-tripping without drift.
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

// KeyVaultCfg_Complete provisions a vault carrying an access policy for the running
// identity, an explicit network ACL, public network access, and tags. tenant_id/
// object_id come from the client-config data source so the access policy's tenant
// matches the vault's (an ARM invariant).
//
// enable_rbac_authorization is set false because access policies conflict with RBAC
// (an ARM invariant; azurerm marks access_policy ConflictsWith enable_rbac_authorization)
// and the resource defaults RBAC on. A vault's permission model cannot be switched in
// place without Microsoft.Authorization/roleAssignments/write, so this scenario runs on
// its own RBAC-disabled vault rather than mutating the RBAC-on Basic vault.
type KeyVaultCfg_Complete KeyVaultCfg

func (r KeyVaultCfg_Complete) Config() string {
	kv := KeyVaultCfg(r)
	return kv.config(fmt.Sprintf(`
  properties = {
    tenant_id             = %[1]s
    public_network_access = "Enabled"
    enable_rbac_authorization = false
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
// survive an in-place update (Update -> Read -> empty plan). Like Complete, it keeps
// RBAC disabled so the vault's permission model is stable across the update.
type KeyVaultCfg_Complete_update KeyVaultCfg

func (r KeyVaultCfg_Complete_update) Config() string {
	kv := KeyVaultCfg(r)
	return kv.config(fmt.Sprintf(`
  properties = {
    tenant_id             = %[1]s
    public_network_access = "Enabled"
    enable_rbac_authorization = false
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

// KeyVaultCfg_AccessPolicy is a minimal vault under the legacy access-policy
// authorization model: enable_rbac_authorization is false and a single access policy
// grants the running identity get/list on keys and secrets. It is the access-policy
// pole of the authorization-model switch scenario — applied, then swapped with
// KeyVaultCfg_RBAC and back — and is deliberately minimal (tenant_id + sku + the policy)
// so a switch diff touches only enable_rbac_authorization and access_policies.
type KeyVaultCfg_AccessPolicy KeyVaultCfg

func (r KeyVaultCfg_AccessPolicy) Config() string {
	kv := KeyVaultCfg(r)
	return kv.config(fmt.Sprintf(`
  properties = {
    tenant_id                 = %[1]s
    enable_rbac_authorization = false
    sku = {
      name   = "standard"
      family = "A"
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
  }`, kv.clientConfig.RefOf("tenant_id"), kv.clientConfig.RefOf("object_id")))
}

// KeyVaultCfg_RBAC is the same minimal vault under the Azure RBAC authorization model:
// enable_rbac_authorization is true and no access policies are configured (they conflict
// with RBAC). It is the RBAC pole of the authorization-model switch scenario, and doubles
// as the standalone "RBAC enabled" acceptance config. Switching a live vault's permission
// model requires Microsoft.Authorization/roleAssignments/write (Owner / User Access
// Administrator), so the switch scenarios run under credentials that carry it.
type KeyVaultCfg_RBAC KeyVaultCfg

func (r KeyVaultCfg_RBAC) Config() string {
	kv := KeyVaultCfg(r)
	return kv.config(fmt.Sprintf(`
  properties = {
    tenant_id                 = %[1]s
    enable_rbac_authorization = true
    sku = {
      name   = "standard"
      family = "A"
    }
  }`, kv.clientConfig.RefOf("tenant_id")))
}

func (r KeyVaultCfg) config(body string) string {
	return r.RenderConfig(config.ConfigEnvelope{
		Name:       "accazapikv" + r.ResourceLabel() + "{{.RandomString}}",
		ParentAttr: "resource_group_id",
		ParentRef:  r.resourceGroup.IDRef(),
		Location:   true,
		Body:       body,
	})
}
