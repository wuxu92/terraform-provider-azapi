package keyvault_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/keyvault"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
)

// Azure Key Vault demonstrates the scoped model and cross-resource config reuse: the
// resource group (its builder lives in the resources package) is a root-scope base, and
// a nested scope owns the vault under test on that shared base. The vault is created once
// in the nested BeforeAll, asserted and mutated by the ordered scenarios, and destroyed
// once when the nested container ends — the resource group survives for any sibling container.
var _ = Describe("Azure Key Vault", Ordered, func() {
	ws := acc.NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	// Root scope: the resource group base and the shared azapi_client_config data source,
	// both reused by every nested scope. The client config is declared once here (not in
	// each vault config) so a single data block is shared across the run; the vault reads
	// its tenant/object ids by address.
	rgCfg := resources.NewResourceGroupCfg()
	rg := ws.ResourceFor(rgCfg)
	BeforeAll(func() {
		ws.DataSource(config.ClientConfig)
		rg.Apply(resources.ResourceGroupCfg_Basic(rgCfg), acc.Exists())
	})

	Describe("a key vault", Ordered, func() {
		// Child scope: owns the vaults, torn down after this container while the resource
		// group base stays. Each vault injects the resource group's Terraform address
		// (rgCfg.IDRef) as its resource_group_id reference.
		scope := ws.Scope()

		// Vault under the RBAC-on-by-default authorization model. It carries no access
		// policies (which conflict with RBAC), so Basic exercises the default round-trip.
		cfg := keyvault.NewKeyVaultCfg(rgCfg, config.ClientConfig)
		kv := scope.ResourceFor(cfg)

		// Separate vault born with RBAC disabled so it can carry access policies. A
		// vault's permission model can't be switched in place without
		// Microsoft.Authorization/roleAssignments/write, so the access-policy scenarios
		// run on their own vault rather than mutating the RBAC-on vault above.
		cfgAP := keyvault.NewKeyVaultCfg(rgCfg, config.ClientConfig, "ap")
		kvAP := scope.ResourceFor(cfgAP)

		// Vault born under the RBAC authorization model with enable_rbac_authorization set
		// explicitly — the unconditional enable-RBAC acceptance case. Creating under RBAC is
		// a plain create, not a permission-model change, so it needs no elevated role.
		cfgRbac := keyvault.NewKeyVaultCfg(rgCfg, config.ClientConfig, "rbac")
		kvRbac := scope.ResourceFor(cfgRbac)

		// Fourth vault for the authorization-model switch: created under the access-policy
		// model, flipped to RBAC, then flipped back. Switching a live vault's permission
		// model requires unrestricted Microsoft.Authorization/roleAssignments/write (Owner /
		// User Access Administrator); it runs by default and is gated by the shared
		// acc.SkipIfNoRoleAssignmentWrite knob (set ARM_TEST_ROLE_ASSIGNMENT_WRITE=false to skip).
		cfgSw := keyvault.NewKeyVaultCfg(rgCfg, config.ClientConfig, "switch")
		kvSw := scope.ResourceFor(cfgSw)

		It("Basic", func() {
			kv.Apply(keyvault.KeyVaultCfg_Basic(cfg), acc.Exists()).ImportVerify()
		})

		When("Complete and Update", func() {
			It("creates a vault with access policy, network acls and tags", func() {
				// Complete provisions an RBAC-disabled vault with an access policy, an
				// explicit network ACL, public network access and tags — the vault's
				// access-policy surface; the Apply re-plans for drift, so the create
				// round-trips without per-value assertions.
				kvAP.Apply(keyvault.KeyVaultCfg_Complete(cfgAP), acc.Exists()).ImportVerify()
			})

			It("updates access policy and tags in place", func() {
				// Complete_update flips the tag value and widens the access-policy permissions
				// set by Complete; each Apply re-plans for drift, so the in-place update
				// round-trips without per-value assertions.
				kvAP.Apply(keyvault.KeyVaultCfg_Complete_update(cfgAP)).ImportVerify()
			})

		})

		It("enables the RBAC authorization model", func() {
			// A vault created directly with enable_rbac_authorization = true (no access
			// policies, which conflict with RBAC). Creating under RBAC is a plain create, not
			// a permission-model change, so it runs unconditionally.
			kvRbac.Apply(keyvault.KeyVaultCfg_RBAC(cfgRbac), acc.Exists(),
				acc.Key("properties.enable_rbac_authorization").HasValue("true")).ImportVerify()
		})

		When("Authorization model switch", func() {
			It("creates under the access-policy model", func() {
				acc.SkipIfNoRoleAssignmentWrite()
				// Born under the legacy access-policy model (RBAC disabled, one policy) — the
				// starting pole of the switch.
				kvSw.Apply(keyvault.KeyVaultCfg_AccessPolicy(cfgSw), acc.Exists(),
					acc.Key("properties.enable_rbac_authorization").HasValue("false")).ImportVerify()
			})

			It("switches to the RBAC model in place", func() {
				acc.SkipIfNoRoleAssignmentWrite()
				// access-policy -> RBAC: enables RBAC and drops the access policies in one
				// update; the Apply re-plans for drift, proving the switched model round-trips.
				kvSw.Apply(keyvault.KeyVaultCfg_RBAC(cfgSw),
					acc.Key("properties.enable_rbac_authorization").HasValue("true")).ImportVerify()
			})

			It("switches back to the access-policy model in place", func() {
				acc.SkipIfNoRoleAssignmentWrite()
				// RBAC -> access-policy: disables RBAC and re-adds the access policy, proving
				// the reverse switch round-trips too.
				kvSw.Apply(keyvault.KeyVaultCfg_AccessPolicy(cfgSw),
					acc.Key("properties.enable_rbac_authorization").HasValue("false")).ImportVerify()
			})
		})
	})
})
