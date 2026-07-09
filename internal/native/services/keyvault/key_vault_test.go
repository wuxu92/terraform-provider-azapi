package keyvault_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/keyvault"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// Azure Key Vault is exercised with the minimum number of live vaults. A vault's
// authorization model (Azure RBAC vs legacy access policies) is fixed at create and can
// only be switched in place with Microsoft.Authorization/roleAssignments/write, so the
// two models cannot share one vault — each is its own nested scope owning a single vault:
//   - the RBAC pole (default model) also carries the purge-on-destroy teardown check;
//   - the access-policy pole carries the access-policy / network-ACL / tags surface.
//
// The resource group base and the shared azapi_client_config data source live at the
// root scope and are reused by every nested scope; the resource group survives while each
// vault scope tears its own vault down. The in-place authorization-model switch has its
// own scope but is permission-gated and currently skipped, so it creates no vault.
var _ = Describe("Azure Key Vault", Ordered, func() {
	ws := acc.NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	// Root scope: the resource group base and the shared azapi_client_config data source,
	// both reused by every nested scope. The client config is declared once here (not in
	// each vault config) so a single data block is shared across the run; each vault reads
	// its tenant/object ids by address.
	rgCfg := resources.NewResourceGroupCfg()
	rg := ws.ResourceFor(rgCfg)
	BeforeAll(func() {
		ws.DataSource(config.ClientConfig)
		rg.Apply(resources.ResourceGroupCfg_Basic(rgCfg), acc.Exists())
	})

	Describe("under the RBAC authorization model", Ordered, func() {
		// One vault under the default (RBAC-on) model. It proves the default authorization
		// model round-trips, then flips purge_on_destroy on in place so its teardown drives
		// the AfterDelete purge hook.
		scope := ws.Scope()
		cfg := keyvault.NewKeyVaultCfg(rgCfg, config.ClientConfig)
		kv := scope.ResourceFor(cfg)

		It("creates a vault with RBAC enabled by default", func() {
			// Basic sets no enable_rbac_authorization, so asserting it is true proves the
			// resource defaults the RBAC model on (azapi's behavior, unlike AzureRM).
			kv.Apply(keyvault.KeyVaultCfg_Basic(cfg), acc.Exists(),
				acc.Key("properties.enable_rbac_authorization").HasValue("true")).ImportVerify()
		})

		It("enables purge-on-destroy in place", NodeTimeout(20 * time.Minute), func() {
			// An in-place add of the write-only purge_on_destroy flag onto the same vault:
			// no ARM body change, so no ImportVerify (ARM never echoes it). Its whole purpose
			// is the scope teardown below, where the AfterDelete hook must purge the vault's
			// soft-deleted shadow; a failing purge POST fails teardown.
			kv.Apply(keyvault.KeyVaultCfg_PurgeOnDestroy(cfg),
				acc.Key("purge_on_destroy").HasValue("true"))
		})
	})

	Describe("under the access-policy authorization model", Ordered, func() {
		// One vault born with RBAC disabled so it can carry access policies (which conflict
		// with RBAC). It exercises the access-policy / network-ACL / tags surface and an
		// in-place update of it.
		scope := ws.Scope()
		cfgAP := keyvault.NewKeyVaultCfg(rgCfg, config.ClientConfig, "ap")
		kvAP := scope.ResourceFor(cfgAP)

		It("creates a vault with an access policy, network ACL and tags", func() {
			// Complete provisions the access-policy surface; the Apply re-plans for drift,
			// so the create round-trips without per-value assertions.
			kvAP.Apply(keyvault.KeyVaultCfg_Complete(cfgAP), acc.Exists()).ImportVerify()
		})

		It("updates the access policy and tags in place", func() {
			// Complete_update flips the tag value and widens the permissions set; each Apply
			// re-plans for drift, so the in-place update round-trips.
			kvAP.Apply(keyvault.KeyVaultCfg_Complete_update(cfgAP)).ImportVerify()
		})
	})

	Describe("switching the authorization model in place", Ordered, func() {
		// The in-place authorization-model switch is a real supported operation but requires
		// unrestricted Microsoft.Authorization/roleAssignments/write (Owner / User Access
		// Administrator); it is gated by acc.SkipIfNoRoleAssignmentWrite and currently
		// skipped outright (see the linked issue), so this scope creates no vault today.
		scope := ws.Scope()
		cfgSw := keyvault.NewKeyVaultCfg(rgCfg, config.ClientConfig, "switch")
		kvSw := scope.ResourceFor(cfgSw)

		It("switches to the RBAC model in place", func() {
			acc.SkipIfNoRoleAssignmentWrite()
			Skip("Skipping Switch to RBAC model due to special permission requirements. See https://learn.microsoft.com/en-us/answers/questions/1396612/not-able-to-change-access-configuration-policy")
			// access-policy -> RBAC: enables RBAC and drops the access policies in one
			// update; the Apply re-plans for drift, proving the switched model round-trips.
			kvSw.Apply(keyvault.KeyVaultCfg_RBAC(cfgSw),
				acc.Key("properties.enable_rbac_authorization").HasValue("true")).ImportVerify()
		})

		It("switches back to the access-policy model in place", func() {
			acc.SkipIfNoRoleAssignmentWrite()
			Skip("Skipping Switch to RBAC model due to special permission requirements. See https://learn.microsoft.com/en-us/answers/questions/1396612/not-able-to-change-access-configuration-policy")
			// RBAC -> access-policy: disables RBAC and re-adds the access policy, proving
			// the reverse switch round-trips too.
			kvSw.Apply(keyvault.KeyVaultCfg_AccessPolicy(cfgSw),
				acc.Key("properties.enable_rbac_authorization").HasValue("false")).ImportVerify()
		})
	})
})
