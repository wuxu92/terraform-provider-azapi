package storage_test

import (
	. "github.com/onsi/ginkgo/v2"

	acc "github.com/Azure/terraform-provider-azapi/internal/native/acceptance"
	"github.com/Azure/terraform-provider-azapi/internal/native/generated/resources"
	"github.com/Azure/terraform-provider-azapi/internal/native/generated/storage"
)

// Azure Storage Blob Service demonstrates deeper reuse: a resource group (root) and a
// storage account (child scope) are each provisioned once from their own packages'
// config builders, and the blob service — the singleton "default" child of that
// account — is exercised on the shared base, injected with the account's Terraform
// address (sa.IDRef).
//
// The blob service lives in the SAME scope as its storage account: blobServices is a
// singleton with no ARM delete operation, so it can only be torn down together with
// the account (the scope's Teardown destroys both and then asserts both are gone).
var _ = Describe("Azure Storage Blob Service", Ordered, func() {
	ws := acc.NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	// Root scope: the resource group base, reused by the account scope below.
	rgCfg := resources.NewResourceGroupCfg("rg")
	rg := ws.ResourceFor(rgCfg)
	BeforeAll(func() { rg.Apply(resources.ResourceGroupCfg_Basic(rgCfg), acc.Exists()) })

	Describe("on a storage account", Ordered, func() {
		// Child scope: the storage account plus its blob service, torn down together
		// after this container while the resource group survives.
		acct := ws.Scope()
		AfterAll(acct.Teardown)

		saCfg := storage.NewStorageAccountCfg(rgCfg, "sa")
		sa := acct.ResourceFor(saCfg)
		blobCfg := storage.NewBlobServiceCfg(saCfg)
		blob := acct.ResourceFor(blobCfg)

		BeforeAll(func() {
			// sa and blob share this scope and tear down together, so provision the
			// dependent chain in ONE terraform apply instead of an apply each: Stage
			// bundles each resource's config + checks and ApplyAll writes both files,
			// applies once (Terraform orders blob after sa from the storage_account_id
			// reference), asserts a single empty post-apply plan for the set, then runs
			// every staged check (Exists for both).
			acct.ApplyAll(
				sa.Stage(storage.StorageAccountCfg_Basic(saCfg), acc.Exists()),
				blob.Stage(storage.BlobServiceCfg_Basic(blobCfg), acc.Exists()),
			)
		})

		It("imports the singleton default blob service with no drift", func() {
			// ApplyAll already asserted existence and a stable plan for the batch;
			// ImportVerify (co-located, not a separate It) re-imports blob and confirms
			// the read path reproduces the configured Basic state. The Basic change_feed
			// value round-trips via the batch's drift plan, so it needs no check.
			blob.ImportVerify()
		})

		It("toggles change feed, then applies a complete blob-service configuration", func() {
			// Each Apply re-plans for drift, covering both the focused update and the broad Complete scenario.
			blob.Apply(storage.BlobServiceCfg_ChangeFeed{BlobServiceCfg: blobCfg, Enabled: false})
			blob.Apply(storage.BlobServiceCfg_Complete(blobCfg))
		})

		It("rejects a blob service name other than \"default\"", func() {
			invalid := storage.NewBlobServiceCfg(saCfg, "invalid")
			acct.ResourceFor(invalid).ApplyExpectError(
				storage.BlobServiceCfg_Named{BlobServiceCfg: invalid, Name: "notdefault"},
				`name value must be one of`,
			)
		})
	})
})
