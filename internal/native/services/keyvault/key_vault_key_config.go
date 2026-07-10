package keyvault

import (
	"strings"

	"github.com/Azure/terraform-provider-azapi/internal/native/services/authorization"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
)

// KeyVaultKeyCfg carries the Terraform address metadata plus the authorization-ready
// key-vault dependency for azapi_key_vault_key acceptance-test scenarios. Construct it
// with NewKeyVaultKeyAccessPolicyCfg when the parent vault is access-policy enabled,
// or NewKeyVaultKeyRBACCfg when the parent vault is RBAC enabled and has a role
// assignment dependency. The returned HCL is a template rendered by the acceptance
// framework ({{.RandomString}}).
type KeyVaultKeyCfg struct {
	config.ResourceConfigBase
	keyVault  KeyVaultCfg
	dependsOn []string
}

// NewKeyVaultKeyAccessPolicyCfg builds a key config whose parent key vault scenario
// grants key permissions through its access_policies block. The key references the
// vault's ID as key_vault_id; no extra depends_on is needed beyond that direct parent
// reference.
func NewKeyVaultKeyAccessPolicyCfg(keyVault KeyVaultCfg_KeyOperationsAccessPolicy, label ...string) KeyVaultKeyCfg {
	return newKeyVaultKeyCfg(KeyVaultCfg(keyVault), nil, label...)
}

// NewKeyVaultKeyRBACCfg builds a key config whose parent key vault scenario uses
// Azure RBAC. The role assignment is not part of the key's ARM parent chain, so the
// emitted key config carries depends_on to force Terraform to create the grant before
// the key.
func NewKeyVaultKeyRBACCfg(keyVault KeyVaultCfg_RBAC, roleAssignment authorization.KeyVaultCryptoOfficerRoleAssignmentCfg, label ...string) KeyVaultKeyCfg {
	return newKeyVaultKeyCfg(KeyVaultCfg(keyVault), []string{roleAssignment.ResourceType() + "." + roleAssignment.ResourceLabel()}, label...)
}

func newKeyVaultKeyCfg(keyVault KeyVaultCfg, dependsOn []string, label ...string) KeyVaultKeyCfg {
	return KeyVaultKeyCfg{
		ResourceConfigBase: config.NewResourceConfigBase(KeyVaultKey.Name, label...),
		keyVault:           keyVault,
		dependsOn:          dependsOn,
	}
}

// KeyVaultKeyCfg_Basic is a minimal valid management-plane RSA key: kty and key_ops
// are the only writable fields required by Microsoft.KeyVault/vaults/keys. ForceNew
// fields such as name, key_vault_id and kty are held stable across the scenario chain;
// key_size is deliberately omitted so Azure's RSA default can round-trip as computed.
type KeyVaultKeyCfg_Basic KeyVaultKeyCfg

func (r KeyVaultKeyCfg_Basic) Config() string {
	return KeyVaultKeyCfg(r).config(`
  properties = {
    kty     = "RSA"
    key_ops = ["encrypt", "decrypt"]
  }`)
}

// KeyVaultKeyCfg_Complete expands the Basic key using management-plane-valid mutable
// fields only: key_ops gains the ARM schema's import operation, attributes.enabled is
// set explicitly, and tags are added. The release operation is intentionally omitted
// because live Azure may require a secure-key-release policy/exportable key pairing;
// import is the management-plane operation this scenario pins without those extra
// service-side prerequisites.
type KeyVaultKeyCfg_Complete KeyVaultKeyCfg

func (r KeyVaultKeyCfg_Complete) Config() string {
	return KeyVaultKeyCfg(r).config(`
  properties = {
    kty     = "RSA"
    key_ops = ["encrypt", "decrypt", "sign", "verify", "wrapKey", "unwrapKey", "import"]
    attributes = {
      enabled = true
    }
  }
  tags = {
    scenario = "complete"
  }`)
}

// KeyVaultKeyCfg_Complete_update flips only in-place-updatable fields from Complete:
// attributes.enabled and tags. kty remains unchanged because it is ForceNew, and the
// expanded key_ops list is held stable so this step proves a clean mutable update
// rather than depending on operation-list replacement semantics.
type KeyVaultKeyCfg_Complete_update KeyVaultKeyCfg

func (r KeyVaultKeyCfg_Complete_update) Config() string {
	return KeyVaultKeyCfg(r).config(`
  properties = {
    kty     = "RSA"
    key_ops = ["encrypt", "decrypt", "sign", "verify", "wrapKey", "unwrapKey", "import"]
    attributes = {
      enabled = false
    }
  }
  tags = {
    scenario = "updated"
  }`)
}

func (r KeyVaultKeyCfg) config(body string) string {
	if len(r.dependsOn) > 0 {
		var b strings.Builder
		b.WriteString(body)
		b.WriteString("\n  depends_on = [")
		for _, dep := range r.dependsOn {
			b.WriteString("\n    ")
			b.WriteString(dep)
			b.WriteString(",")
		}
		b.WriteString("\n  ]")
		body = b.String()
	}

	return r.RenderConfig(config.ConfigEnvelope{
		Name:       "acctestkey{{.RandomString}}",
		ParentAttr: "key_vault_id",
		ParentRef:  r.keyVault.IDRef(),
		Body:       body,
	})
}
