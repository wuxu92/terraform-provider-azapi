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
// fields such as name, key_vault_id, kty and key_ops are held stable; key_size is
// deliberately omitted so Azure's RSA default can round-trip as computed.
type KeyVaultKeyCfg_Basic KeyVaultKeyCfg

func (r KeyVaultKeyCfg_Basic) Config() string {
	return KeyVaultKeyCfg(r).config(`
  properties = {
    kty     = "RSA"
    key_ops = ["encrypt", "decrypt"]
  }`)
}

// KeyVaultKeyCfg_Complete is a create-time-only key shape that exercises additional
// management-plane-valid fields. Microsoft.KeyVault/vaults/keys exposes
// Keys_CreateIfNotExist: ARM can create the first version but cannot update an
// existing key, so acceptance tests must apply this shape to a distinct key instance
// rather than mutate a Basic key in place.
type KeyVaultKeyCfg_Complete KeyVaultKeyCfg

func (r KeyVaultKeyCfg_Complete) Config() string {
	// do not include "import" in key_ops as it's for rsa-hsm only
	return KeyVaultKeyCfg(r).config(`
  properties = {
    kty     = "RSA"
    key_ops = ["encrypt", "decrypt", "sign", "verify", "wrapKey", "unwrapKey"]
    attributes = {
      enabled = true
    }
  }
  tags = {
    scenario = "complete"
  }`)
}

// KeyVaultKeyCfg_Complete_update is retained as a replacement-plan fixture: all
// writable Key Vault key fields are create-time-only through ARM, so a live acceptance
// test must not apply this over an existing key unless it also changes the key name.
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
		Name:       keyVaultKeyName(r.ResourceLabel()),
		ParentAttr: "key_vault_id",
		ParentRef:  r.keyVault.IDRef(),
		Body:       body,
	})
}

func keyVaultKeyName(label string) string {
	suffix := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-':
			return r
		default:
			return -1
		}
	}, label)
	if suffix == "" {
		suffix = "test"
	}
	return "acctestkey" + suffix + "{{.RandomString}}"
}
