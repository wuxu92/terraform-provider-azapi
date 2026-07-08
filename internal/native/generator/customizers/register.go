package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/armtypes"

// register wires every per-resource customizer into the registry in one place.
// Each customizer function is defined in its own <resource>.go file; this file is
// the single registration point, keyed by ARM resource type (armtypes constant).
// Add a resource by dropping its <resource>.go alongside and adding one Register
// line here.
func init() {
	Register(armtypes.StorageAccount, customizeStorageAccount)
	Register(armtypes.StorageAccountBlobService, customizeStorageAccountBlobService)
	Register(armtypes.ResourceGroup, customizeResourceGroup)
	Register(armtypes.WebServerFarm, customizeWebServerFarm)
	Register(armtypes.WebSite, customizeWebSite)
	Register(armtypes.KeyVault, customizeKeyVault)
	Register(armtypes.RoleDefinition, customizeRoleDefinition)
}
