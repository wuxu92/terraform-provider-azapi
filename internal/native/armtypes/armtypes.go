// Package armtypes is the single source of truth for the ARM resource type
// strings the native generator handles. Reference these constants instead of repeating the
// literal "<Namespace>/<type>" strings throughout the codebase (customizer
// registration, generator wiring, tests), so a type is spelled exactly once and
// renames stay mechanical.
//
// Values are the bare ARM resource type without an API version — the form used
// to key customizers and to compose resource IDs. Append "@<api-version>" where
// a fully-qualified tag is needed.
package armtypes

const (
	// StorageAccount is Microsoft.Storage/storageAccounts.
	StorageAccount = "Microsoft.Storage/storageAccounts"

	// StorageAccountBlobService is Microsoft.Storage/storageAccounts/blobServices,
	// the singleton blob-service child of a storage account (name always "default").
	StorageAccountBlobService = "Microsoft.Storage/storageAccounts/blobServices"

	// ResourceGroup is Microsoft.Resources/resourceGroups, a subscription-scoped
	// top-level resource (envelope parent: subscription_id).
	ResourceGroup = "Microsoft.Resources/resourceGroups"

	// WebServerFarm is Microsoft.Web/serverfarms, the App Service plan resource
	// required by App Service web apps and function apps.
	WebServerFarm = "Microsoft.Web/serverfarms"

	// WebSite is Microsoft.Web/sites, the shared ARM resource type used by
	// App Service web apps and function apps.
	WebSite = "Microsoft.Web/sites"

	// UserAssignedIdentity is Microsoft.ManagedIdentity/userAssignedIdentities, a
	// resource-group-scoped standalone managed identity (envelope parent:
	// resource_group_id).
	UserAssignedIdentity = "Microsoft.ManagedIdentity/userAssignedIdentities"
)
