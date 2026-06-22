package azwise

// RegisterAll registers knowledge for all known resource types.
// Add new resource types here.
func RegisterAll() {
	Register(NewStorageAccount())
	Register(NewStorageAccountBlobService())
	Register(NewKeyVault())
	Register(NewKeyVaultKey())
	Register(NewKeyVaultSecret())
}
