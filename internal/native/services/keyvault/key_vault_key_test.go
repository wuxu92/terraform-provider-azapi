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
		keyCfg := keyvault.NewKeyVaultKeyAccessPolicyCfg(keyvault.KeyVaultCfg_KeyOperationsAccessPolicy(vaultCfg), "ap")
		key := scope.ResourceFor(keyCfg)

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
			key.Apply(keyvault.KeyVaultKeyCfg_Basic(keyCfg), acc.Exists()).ImportVerify()
		})

		It("adds management-plane key operations in place", func() {
			// Complete keeps ForceNew fields stable and adds the ARM schema's import operation
			// alongside mutable attributes/tags; the follow-up plan proves the update round-trips.
			key.Apply(keyvault.KeyVaultKeyCfg_Complete(keyCfg)).ImportVerify()
		})

		It("updates mutable attributes and tags in place", func() {
			// Complete_update flips only attributes.enabled and tags while holding kty and
			// key_ops steady, proving a pure mutable update through ARM and import/read.
			key.Apply(keyvault.KeyVaultKeyCfg_Complete_update(keyCfg)).ImportVerify()
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
