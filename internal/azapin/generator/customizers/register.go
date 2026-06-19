package customizers

// register wires every per-resource customizer into the registry in one place.
// Each customizer function is defined in its own <resource>.go file; this file is
// the single registration point, keyed by ARM resource type. Add a resource by
// dropping its <resource>.go alongside and adding one Register line here.
func init() {
	Register("Microsoft.Storage/storageAccounts", customizeStorageAccount)
}
