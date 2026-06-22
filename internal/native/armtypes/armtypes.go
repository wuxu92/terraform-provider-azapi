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
)
