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
		// Child scope: owns the vault, torn down after this container while the resource
		// group base stays. The vault injects the resource group's Terraform address
		// (rgCfg.IDRef) as its resource_group_id reference.
		scope := ws.Scope()
		cfg := keyvault.NewKeyVaultCfg(rgCfg, config.ClientConfig)
		kv := scope.ResourceFor(cfg)

		It("Basic", func() {
			kv.Apply(keyvault.KeyVaultCfg_Basic(cfg), acc.Exists()).ImportVerify()
		})

		When("Complete and Update", func() {
			It("adds access policy, network acls and tags in place", func() {
				// Complete adds an access policy, an explicit network ACL, public network
				// access and tags — the vault's in-place-updatable surface; each Apply
				// re-plans for drift, so the in-place add round-trips without per-value assertions.
				kv.Apply(keyvault.KeyVaultCfg_Complete(cfg)).ImportVerify()
			})

			It("updates access policy and tags in place", func() {
				// Complete_update flips the tag value and widens the access-policy permissions
				// set by Complete; each Apply re-plans for drift, so the update round-trips
				// without per-value assertions.
				kv.Apply(keyvault.KeyVaultCfg_Complete_update(cfg)).ImportVerify()
			})

		})
	})
})
