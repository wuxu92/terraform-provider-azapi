package keyvault_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/authorization"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/keyvault"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// Azure Key Vault Keys are managed through the ARM management-plane child resource
// (Microsoft.KeyVault/vaults/keys), not Key Vault's data-plane key API. The parent
// vault still needs to authorize the caller for key object operations, so this suite
// creates keys through both supported vault-side authorization prerequisites: legacy
// access policies and Azure RBAC with a resource-group-scoped Key Vault Crypto Officer
// role assignment for the running identity.
var _ = Describe("Azure Key Vault Key", Ordered, func() {
	ws := acc.NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	// Root scope: the resource group base and shared azapi_client_config data source for
	// the parent vaults' tenant_id. The child scopes below own their vault and key so
	// teardown destroys each key before its vault while the resource group survives.
	rgCfg := resources.NewResourceGroupCfg("rg")
	rg := ws.ResourceFor(rgCfg)
	BeforeAll(func() {
		ws.DataSource(config.ClientConfig)
		rg.Apply(resources.ResourceGroupCfg_Basic(rgCfg), acc.Exists())
	})

	Describe("an RSA key in an access-policy vault", Ordered, func() {
		scope := ws.Scope()
		vaultCfg := keyvault.NewKeyVaultCfg(rgCfg, config.ClientConfig, "ap")
		vault := scope.ResourceFor(vaultCfg)
		basicKeyCfg := keyvault.NewKeyVaultKeyAccessPolicyCfg(keyvault.KeyVaultCfg_KeyOperationsAccessPolicy(vaultCfg), "ap_basic")
		basicKey := scope.ResourceFor(basicKeyCfg)
		completeKeyCfg := keyvault.NewKeyVaultKeyAccessPolicyCfg(keyvault.KeyVaultCfg_KeyOperationsAccessPolicy(vaultCfg), "ap_complete")
		completeKey := scope.ResourceFor(completeKeyCfg)

		BeforeAll(func() {
			// The key is created through ARM, but Key Vault still enforces vault-side key
			// permissions. This parent config grants the running identity the required key
			// object permissions without using a data-plane API.
			vault.Apply(keyvault.KeyVaultCfg_KeyOperationsAccessPolicy(vaultCfg), acc.Exists())
		})

		It("creates and imports a minimal RSA key", func() {
			// Basic supplies only the required management-plane key fields (kty and key_ops).
			// Apply asserts create -> read -> empty plan; ImportVerify proves the import/read
			// path reproduces that minimal state without drift.
			basicKey.Apply(keyvault.KeyVaultKeyCfg_Basic(basicKeyCfg), acc.Exists()).ImportVerify()
		})

		It("creates and imports a key with expanded management-plane create parameters", func() {
			// Microsoft.KeyVault/vaults/keys exposes Keys_CreateIfNotExist: ARM can create
			// the first version but cannot update an existing key. Exercise the expanded
			// key_ops/attributes/tags shape as a distinct key, not an in-place mutation.
			completeKey.Apply(keyvault.KeyVaultKeyCfg_Complete(completeKeyCfg), acc.Exists()).ImportVerify()
		})
	})

	Describe("an RSA key in an RBAC vault", Ordered, func() {
		rbacScope := ws.Scope()
		vaultCfg := keyvault.NewKeyVaultCfg(rgCfg, config.ClientConfig, "rbac")
		vault := rbacScope.ResourceFor(vaultCfg)
		roleCfg := authorization.NewKeyVaultCryptoOfficerRoleAssignmentCfg(rgCfg, config.ClientConfig, "rbac_crypto")
		role := rbacScope.ResourceFor(roleCfg)

		BeforeAll(func() {
			// Creating the role assignment requires Microsoft.Authorization/roleAssignments/write;
			// gate the scenario before rendering or applying the role-assignment-backed vault.
			acc.SkipIfNoRoleAssignmentWrite()
			rbacScope.ApplyAll(
				vault.Stage(keyvault.KeyVaultCfg_RBAC(vaultCfg), acc.Exists(),
					acc.Key("properties.enable_rbac_authorization").HasValue("true")),
				role.Stage(roleCfg, acc.Exists()),
			)
		})

		Describe("with a resource-group-scoped Key Vault Crypto Officer assignment", Ordered, func() {
			keyScope := rbacScope.Scope()
			keyCfg := keyvault.NewKeyVaultKeyRBACCfg(keyvault.KeyVaultCfg_RBAC(vaultCfg), roleCfg, "rbac")
			key := keyScope.ResourceFor(keyCfg)

			It("creates and imports a minimal RSA key", func() {
				// This repeats the create/import seam under the RBAC authorization prerequisite only;
				// the access-policy scenario above owns the broader key mutation chain.
				key.Apply(keyvault.KeyVaultKeyCfg_Basic(keyCfg), acc.Exists()).ImportVerify()
			})
		})
	})
})
