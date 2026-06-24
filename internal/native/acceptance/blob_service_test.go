package nativeacc

import (
	. "github.com/onsi/ginkgo/v2"
)

// Azure Storage Blob Service demonstrates deeper reuse: a resource group (root) and
// a storage account (child scope) are each provisioned once, and the blob service —
// the singleton "default" child of that account — is exercised on the shared base,
// injected with the account reference by Terraform address.
//
// The blob service lives in the SAME scope as its storage account: blobServices is a
// singleton with no ARM delete operation, so it can only be torn down together with
// the account (the scope's Teardown destroys both and then asserts both are gone).
var _ = Describe("Azure Storage Blob Service", Ordered, func() {
	ws := NewWorkspace()
	BeforeAll(ws.Start)
	AfterAll(ws.Destroy)

	// Root scope: the resource group base, reused by the account scope below.
	rg := ws.Resource("azapi_resource_group", "rg")
	BeforeAll(func() { rg.Apply(resourceGroupConfig("rg"), Exists()) })

	Describe("on a storage account", Ordered, func() {
		const rgRef = "azapi_resource_group.rg.id"
		const saRef = "azapi_storage_account.sa.id"

		// Child scope: the storage account plus its blob service, torn down together
		// after this container while the resource group survives.
		acct := ws.Scope()
		AfterAll(acct.Teardown)
		sa := acct.Resource("azapi_storage_account", "sa")
		blob := acct.Resource("azapi_storage_account_blob_service", "test")

		BeforeAll(func() {
			sa.Apply(storageAccountConfig("sa", rgRef, "Standard_LRS", ""), Exists())
		})

		It("creates the singleton default blob service, then imports it", func() {
			// Apply asserts existence and no post-apply drift; ImportVerify (co-located,
			// not a separate It) re-imports and confirms the read path has no drift. The
			// change_feed value set in config round-trips, so it needs no explicit Key check.
			blob.Apply(blobServiceConfig("test", saRef, "default", changeFeed(true)), Exists()).
				ImportVerify()
		})

		It("toggles change feed and versioning in place", func() {
			// Each Apply re-plans for drift, covering both in-place property changes.
			blob.Apply(blobServiceConfig("test", saRef, "default", changeFeed(false)))
			blob.Apply(blobServiceConfig("test", saRef, "default", versioning(true)))
		})

		It("rejects a blob service name other than \"default\"", func() {
			acct.Resource("azapi_storage_account_blob_service", "invalid").ApplyExpectError(
				blobServiceConfig("invalid", saRef, "notdefault", versioning(true)),
				`name value must be one of`,
			)
		})
	})
})
