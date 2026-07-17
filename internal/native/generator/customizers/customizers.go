// Package customizers is the developer-facing extension point for per-resource
// schema customization of native static resources. It lives under the generator
// (not in the generator package itself) so the growing set of per-resource
// customizers has a dedicated home, separate from the generator core.
//
// The whole point is that customization happens at GENERATION time, not at
// runtime: a customizer mutates the parsed type graph and the operational-
// envelope spec, and the emitter bakes the result into the generated _gen.go.
// The shipped schema is therefore always final — the runtime never rewrites it,
// and nothing in this package is compiled into the provider runtime.
//
// To customize a resource, add its customizer function in a `<resource>.go` file
// here and wire it up with a single Register line in register.go (the one place
// that holds the init() and registers every customizer, keyed by ARM type):
//
//	// storage_account.go
//	func customizeStorageAccount(def *typegraph.ResourceDefinition) {
//	    // constrain the resource name (bicep does not model it)
//	    def.SetNameValidators(
//	        typegraph.LengthValidator(3, 24),
//	        typegraph.RegexValidator(`^[a-z0-9]+$`, "name must be 3-24 lowercase letters and digits"),
//	    )
//	    // fluent, path-variadic helpers on *ResourceDefinition (typegraph/customize.go)
//	    // chain and panic on a bad path; see that file for the full vocabulary
//	    def.Default("properties.minimumTlsVersion", "TLS1_2").Required("sku", "sku.name")
//	}
//
//	// register.go
//	func init() {
//	    Register(armtype.StorageAccount, customizeStorageAccount)
//	}
//
// The generator command imports this package (so the register.go init() runs) and
// calls Apply on the post-processed definitions, after generator.PostProcess, so a
// hand-written customizer has the final say before emission.
package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/typegraph"

// Customizer is a developer-authored, per-resource schema hook that runs at
// generation time, after azwise overlay and envelope defaults, before emission.
// It mutates the resolved type graph (Property flags/validators/defaults) and/or
// the envelope spec in place.
type Customizer func(*typegraph.ResourceDefinition)

// registry maps an ARM resource type (no API version) to its customizer.
var registry = map[string]Customizer{}

// Register attaches a schema customizer to an ARM resource type (use an armtype
// constant, e.g. armtype.StorageAccount). Call from register.go's init().
// Registering twice for the same type panics — a resource has exactly one
// customizer.
func Register(armType string, c Customizer) {
	if _, dup := registry[armType]; dup {
		panic("native: duplicate customizer for " + armType)
	}
	registry[armType] = c
}

// Unregister removes the customizer for an ARM resource type if present. It is a
// no-op when none is registered. Intended for tests that register a throwaway
// customizer and clean up after themselves.
func Unregister(armType string) {
	delete(registry, armType)
}

// Apply runs the registered customizer (if any) for each definition. Call after
// generator.PostProcess so customizers see the azwise overlay and envelope
// defaults and have the final say before emission.
func Apply(defs []*typegraph.ResourceDefinition) {
	for _, def := range defs {
		if def == nil || def.Body == nil {
			continue
		}
		if c := registry[typegraph.ARMTypeOf(def)]; c != nil {
			c(def)
		}
	}
}
