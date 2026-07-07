package storage_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/managedidentity"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/storage"
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
	BeforeAll(func() { rg.Apply(resources.ResourceGroupCfg_Basic(rgCfg), acc.Exists()) })

	Describe("a StorageV2 account", Ordered, func() {
		// Child scope: owns the storage account, torn down after this container while
		// the resource group base stays. The account injects the resource group's
		// Terraform address (rg.IDRef) as its parent reference.
		acct := ws.Scope()
		cfg := storage.NewStorageAccountCfg(rgCfg)
		sa := acct.ResourceFor(cfg)
		// Native user-assigned identity dependency, created in this scope and referenced
		// by ARM ID in the account's root identity block (azapi_user_assigned_identity
		// now exists as a native resource).
		uaiCfg := managedidentity.NewUserAssignedIdentityCfg(rgCfg)
		uai := acct.ResourceFor(uaiCfg)

		BeforeAll(func() {
			sa.Apply(storage.StorageAccountCfg_Basic(cfg), acc.Exists()).ImportVerify()
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
			sa.Apply(storage.StorageAccountCfg_Complete(cfg))
		})

		It("re-updates every mutable account property in place", func() {
			// Complete_update flips each in-place-updatable value set by Complete; each
			// Apply re-plans for drift, so the update round-trips without per-value
			// assertions.
			sa.Apply(storage.StorageAccountCfg_Complete_update(cfg))
		})

		It("replaces the account when migrating Standard_LRS to Standard_ZRS", func() {
			// Exercises the storage account hook's azwise.CheckForceNew SKU zone-migration
			// rule: Standard_LRS -> Standard_ZRS forces a replace. Exists confirms the
			// replacement account is present in Azure.
			sa.Apply(storage.StorageAccountCfg_SKU{StorageAccountCfg: cfg, SKU: "Standard_ZRS"}, acc.Exists())
		})

		It("assigns a user-assigned identity in place", func() {
			// One apply creates the identity and updates the account together; Terraform
			// orders the identity first from the account's user_assigned_identities ref.
			// SKU stays Standard_ZRS (the account's current SKU after the migration step)
			// so the identity is added in place, not via a ForceNew replace. The drift
			// plan proves the identity block round-trips; the Key check pins the type.
			acct.ApplyAll(
				uai.Stage(managedidentity.UserAssignedIdentityCfg_Basic(uaiCfg), acc.Exists()),
				sa.Stage(storage.StorageAccountCfg_Identity{
					StorageAccountCfg: cfg,
					SKU:               "Standard_ZRS",
					Identity:          uaiCfg.IDRef(),
				}, acc.Key("identity.type").HasValue("UserAssigned")),
			)
		})
	})
})
