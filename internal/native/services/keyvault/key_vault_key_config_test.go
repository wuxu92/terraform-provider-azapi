package keyvault_test

import (
	"testing"

	"github.com/Azure/terraform-provider-azapi/internal/native/services/authorization"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/keyvault"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

func TestKeyVaultKeyCfgAuthDependencies(t *testing.T) {
	rg := resources.NewResourceGroupCfg("rg")
	vault := keyvault.NewKeyVaultCfg(rg, config.ClientConfig, "kv")

	t.Run("access policy vault has no role assignment dependency", func(t *testing.T) {
		got := keyvault.KeyVaultKeyCfg_Basic(
			keyvault.NewKeyVaultKeyAccessPolicyCfg(keyvault.KeyVaultCfg_KeyOperationsAccessPolicy(vault), "ap"),
		).Config()
		want := `
resource "azapi_key_vault_key" "ap" {
  name         = "acctestkeyap{{.RandomString}}"
  key_vault_id = azapi_key_vault.kv.id
  properties = {
    kty     = "RSA"
    key_ops = ["encrypt", "decrypt"]
  }
}
`
		if got != want {
			t.Errorf("access-policy key config mismatch\n got: %q\nwant: %q", got, want)
		}
	})

	t.Run("RBAC vault depends on role assignment", func(t *testing.T) {
		role := authorization.NewKeyVaultCryptoOfficerRoleAssignmentCfg(rg, config.ClientConfig, "crypto")
		got := keyvault.KeyVaultKeyCfg_Basic(
			keyvault.NewKeyVaultKeyRBACCfg(keyvault.KeyVaultCfg_RBAC(vault), role, "rbac"),
		).Config()
		want := `
resource "azapi_key_vault_key" "rbac" {
  name         = "acctestkeyrbac{{.RandomString}}"
  key_vault_id = azapi_key_vault.kv.id
  properties = {
    kty     = "RSA"
    key_ops = ["encrypt", "decrypt"]
  }
  depends_on = [
    azapi_role_assignment.crypto,
  ]
}
`
		if got != want {
			t.Errorf("RBAC key config mismatch\n got: %q\nwant: %q", got, want)
		}
	})
}
