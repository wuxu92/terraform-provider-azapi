package storage_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/generated/resources"
	"github.com/Azure/terraform-provider-azapi/internal/native/generated/storage"
)

// Azure Storage demonstrates the scoped model and cross-resource config reuse: the
// resource group (its builder lives in the resources package) is a root-scope base,
// and a nested container owns the storage account under test on that shared base. The
// account is created once in the nested BeforeAll, asserted and mutated by the ordered
// scenarios, and destroyed once when the nested container ends — the resource group
// survives for any sibling container.
var _ = Describe("Azure Storage", Ordered, func() {
	ws := acc.NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	// Root scope: the resource group base, reused by every nested scope. Its config
	// comes from the resources package's ResourceGroup builder.
	rgCfg := resources.NewResourceGroupCfg("rg")
	rg := ws.ResourceFor(rgCfg)
	BeforeAll(func() { rg.Apply(rgCfg.Basic(), acc.Exists()) })

	Describe("a StorageV2 account", Ordered, func() {
		// Child scope: owns the storage account, torn down after this container while
		// the resource group base stays. The account injects the resource group's
		// Terraform address (rg.IDRef) as its parent reference.
		acct := ws.Scope()
		AfterAll(acct.Teardown)
		cfg := storage.NewStorageAccountCfg(rgCfg)
		sa := acct.ResourceFor(cfg)

		BeforeAll(func() {
			sa.Apply(cfg.Basic(), acc.Exists()).ImportVerify()
		})

		It("applies the azwise verified TLS 1.2 default", func() {
			// A computed default the plan can't prove: minimum_tls_version is omitted
			// from config, so the provider (not the config) supplies TLS1_2 — the kind
			// of fact that still warrants an explicit Key check.
			sa.Check(acc.Key("properties.minimum_tls_version").HasValue("TLS1_2"))
		})

		It("updates to a complete configuration in place", func() {
			// Complete adds an access tier; each Apply re-plans for drift, so the
			// in-place update round-trips without per-value assertions.
			sa.Apply(cfg.Complete())
		})

		It("replaces the account when migrating Standard_LRS to Standard_ZRS", func() {
			// Exercises the storage overlay's azwise.CheckForceNew SKU zone-migration
			// rule: Standard_LRS -> Standard_ZRS forces a replace. Exists confirms the
			// replacement account is present in Azure.
			sa.Apply(cfg.WithSKU("Standard_ZRS"), acc.Exists())
		})
	})
})
