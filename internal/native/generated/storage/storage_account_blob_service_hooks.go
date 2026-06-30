package storage

import nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"

// Storage account blob service hooks model the ARM lifecycle: the singleton
// Microsoft.Storage/storageAccounts/blobServices/default child is created with
// its parent storage account, has no ARM delete operation (DELETE returns 405),
// and disappears when the parent account is deleted. SkipARMDelete makes a
// Terraform destroy a state-only removal for this resource.
func init() {
	nativeresource.RegisterHooks(StorageAccountBlobService.Name, &nativeresource.Hooks{
		SkipARMDelete: true,
	})
}
