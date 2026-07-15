// Package armtype is the single source of truth for the ARM resource type
// strings the native generator handles. Reference these constants instead of repeating the
// literal "<Namespace>/<type>" strings throughout the codebase (customizer
// registration, generator wiring, tests), so a type is spelled exactly once and
// renames stay mechanical.
//
// Values are the bare ARM resource type without an API version — the form used
// to key customizers and to compose resource IDs. Append "@<api-version>" where
// a fully-qualified tag is needed.
package armtype

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

	// KeyVault is Microsoft.KeyVault/vaults, a resource-group-scoped standalone
	// key vault (envelope parent: resource_group_id).
	KeyVault = "Microsoft.KeyVault/vaults"

	// KeyVaultKey is Microsoft.KeyVault/vaults/keys, the management-plane child
	// resource for keys in a Key Vault.
	KeyVaultKey = "Microsoft.KeyVault/vaults/keys"

	// RoleAssignment is Microsoft.Authorization/roleAssignments, a scope-based
	// extension resource assigning a role definition to a principal at that scope.
	RoleAssignment = "Microsoft.Authorization/roleAssignments"

	// RoleDefinition is Microsoft.Authorization/roleDefinitions, a custom RBAC role
	// definition. It is a scope-based extension resource (writable at tenant,
	// management-group, subscription, resource-group, and extension scopes), so its
	// envelope parent is the generic parent_id (the scope the role is defined at).
	RoleDefinition = "Microsoft.Authorization/roleDefinitions"

	// VirtualNetwork is Microsoft.Network/virtualNetworks, a resource-group-scoped
	// standalone virtual network (envelope parent: resource_group_id).
	VirtualNetwork = "Microsoft.Network/virtualNetworks"

	// DataFactory is Microsoft.DataFactory/factories, a resource-group-scoped
	// data integration service. Its body carries a nested discriminated property
	// (repoConfiguration, discriminated by type: FactoryVSTSConfiguration /
	// FactoryGitHubConfiguration).
	DataFactory = "Microsoft.DataFactory/factories"

	// DeploymentScript is Microsoft.Resources/deploymentScripts, a
	// resource-group-scoped resource whose body is itself a discriminated type
	// (root discriminated by kind: AzureCLI / AzurePowerShell).
	DeploymentScript = "Microsoft.Resources/deploymentScripts"

	// KustoClusterDatabase is Microsoft.Kusto/clusters/databases, a child of an
	// Azure Data Explorer cluster (envelope parent: cluster_id) whose body is a
	// discriminated type (root discriminated by kind: ReadWrite /
	// ReadOnlyFollowing).
	KustoClusterDatabase = "Microsoft.Kusto/clusters/databases"

	// DocumentDBDatabaseAccount is Microsoft.DocumentDB/databaseAccounts, a
	// resource-group-scoped Cosmos DB account. Its body carries a nested
	// discriminated property (backupPolicy, discriminated by type: Periodic /
	// Continuous).
	DocumentDBDatabaseAccount = "Microsoft.DocumentDB/databaseAccounts"
)
