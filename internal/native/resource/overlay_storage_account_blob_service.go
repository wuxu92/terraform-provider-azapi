package resource

// Storage account blob service overlay. Microsoft.Storage/storageAccounts/blobServices
// is a singleton child whose ARM name is always "default": it is created implicitly
// with the storage account, has no ARM delete operation (DELETE returns 405), and is
// removed only when the parent account is deleted. SkipARMDelete makes a Terraform
// destroy of this resource a state-only removal so it does not error, and the actual
// Azure cleanup happens when the parent storage account is deleted.
func init() {
	RegisterHooks("azapi_storage_account_blob_service", &Hooks{
		SkipARMDelete: true,
	})
}
