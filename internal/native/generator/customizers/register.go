package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/armtype"

// register wires every per-resource customizer into the registry in one place.
// Each customizer function is defined in its own <resource>.go file; this file is
// the single registration point, keyed by ARM resource type (armtype constant).
// Add a resource by dropping its <resource>.go alongside and adding one Register
// line here.
func init() {
	Register(armtype.StorageAccount, customizeStorageAccount)
	Register(armtype.StorageAccountBlobService, customizeStorageAccountBlobService)
	Register(armtype.ResourceGroup, customizeResourceGroup)
	Register(armtype.WebServerFarm, customizeWebServerFarm)
	Register(armtype.WebSite, customizeWebSite)
	Register(armtype.KeyVault, customizeKeyVault)
	Register(armtype.RoleDefinition, customizeRoleDefinition)
}
