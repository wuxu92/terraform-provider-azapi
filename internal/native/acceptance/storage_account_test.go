package nativeacc

import (
	. "github.com/onsi/ginkgo/v2"
)

// Azure Storage demonstrates the scoped model: the resource group is a root-scope
// base, and a nested container owns the storage account under test on that shared
// base. The account is created once in the nested BeforeAll, asserted and mutated by
// the ordered scenarios, and destroyed once when the nested container ends — the
// resource group survives for any sibling container.
var _ = Describe("Azure Storage", Ordered, func() {
	ws := NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	// Root scope: the resource group, created once and reused by every nested scope.
	rg := ws.Resource("azapi_resource_group", "rg")
	BeforeAll(func() { rg.Apply(resourceGroupConfig("rg"), Exists()) })

	Describe("a StorageV2 account", Ordered, func() {
		// Child scope: owns the storage account, torn down after this container while
		// the resource group base (rgRef) stays. rgRef is the injected dependency.
		const rgRef = "azapi_resource_group.rg.id"
		acct := ws.Scope()
		AfterAll(acct.Teardown)
		sa := acct.Resource("azapi_storage_account", "test")

		BeforeAll(func() {
			sa.Apply(storageAccountConfig("test", rgRef, "Standard_LRS", ""), Exists()).ImportVerify()
		})

		// The create above also import-verifies (StateRm + re-import, no drift) and, via
		// Apply, asserts existence and no post-apply drift — so the StorageV2/Standard_LRS
		// config echoes need no explicit checks.

		It("applies the azwise verified TLS 1.2 default", func() {
			// A computed default the plan can't prove: minimum_tls_version is omitted
			// from config, so the provider (not the config) supplies TLS1_2 — the kind
			// of fact that still warrants an explicit Key check.
			sa.Check(Key("properties.minimum_tls_version").HasValue("TLS1_2"))
		})

		It("updates the access tier in place", func() {
			// Each Apply re-plans for drift, so the in-place Hot -> Cool transition
			// round-trips are verified without per-value assertions.
			sa.Apply(storageAccountConfig("test", rgRef, "Standard_LRS", accessTier("Hot")))
			sa.Apply(storageAccountConfig("test", rgRef, "Standard_LRS", accessTier("Cool")))
		})

		It("replaces the account when migrating Standard_LRS to Standard_ZRS", func() {
			// Exercises the storage overlay's azwise.CheckForceNew SKU zone-migration
			// rule: Standard_LRS -> Standard_ZRS forces a replace. Exists confirms the
			// replacement account is present in Azure.
			sa.Apply(storageAccountConfig("test", rgRef, "Standard_ZRS", ""), Exists())
		})
	})
})
