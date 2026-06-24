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
	rgCfg := resources.ResourceGroup{Label: "rg"}
	rg := ws.ResourceFor(rgCfg)
	BeforeAll(func() { rg.Apply(rgCfg.Basic(), acc.Exists()) })

	Describe("on a storage account", Ordered, func() {
		// Child scope: the storage account plus its blob service, torn down together
		// after this container while the resource group survives.
		acct := ws.Scope()
		AfterAll(acct.Teardown)

		saCfg := storage.StorageAccount{Label: "sa", ResourceGroupIDRef: rg.IDRef()}
		sa := acct.ResourceFor(saCfg)
		blobCfg := storage.BlobService{Label: "test", StorageAccountIDRef: sa.IDRef()}
		blob := acct.ResourceFor(blobCfg)

		BeforeAll(func() {
			sa.Apply(saCfg.Basic(), acc.Exists())
		})

		It("creates the singleton default blob service, then imports it", func() {
			// Apply asserts existence and no post-apply drift; ImportVerify (co-located,
			// not a separate It) re-imports and confirms the read path has no drift. The
			// change_feed value set in config round-trips, so it needs no explicit check.
			blob.Apply(blobCfg.Basic(), acc.Exists()).ImportVerify()
		})

		It("toggles change feed and versioning in place", func() {
			// Each Apply re-plans for drift, covering both in-place property changes.
			blob.Apply(blobCfg.WithChangeFeed(false))
			blob.Apply(blobCfg.Complete())
		})

		It("rejects a blob service name other than \"default\"", func() {
			invalid := storage.BlobService{Label: "invalid", StorageAccountIDRef: sa.IDRef()}
			acct.ResourceFor(invalid).ApplyExpectError(
				invalid.Named("notdefault"),
				`name value must be one of`,
			)
		})
	})
})
