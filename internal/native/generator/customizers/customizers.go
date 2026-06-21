// Package customizers is the developer-facing extension point for per-resource
// schema customization of azapin static resources. It lives under the generator
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
//	func customizeStorageAccount(def *generator.ResourceDefinition) {
//	    // constrain the resource name (bicep does not model it)
//	    def.Envelope.Name.Validators = []generator.DescriptionValidator{
//	        generator.LengthValidator(3, 24),
//	        generator.RegexValidator(`^[a-z0-9]+$`, "name must be 3-24 lowercase letters and digits"),
//	    }
//	    // default an inferred-but-unset property (FindProperty panics on a bad path)
//	    generator.FindProperty(def, "properties.minimumTlsVersion").DefaultValue = "TLS1_2"
//	}
//
//	// register.go
//	func init() {
//	    Register(armtypes.StorageAccount, customizeStorageAccount)
//	}
//
// The generator command imports this package (so the register.go init() runs) and
// calls Apply on the post-processed definitions, after generator.PostProcess, so a
// hand-written customizer has the final say before emission.
package customizers

import "github.com/Azure/terraform-provider-azapi/internal/native/generator"

// Customizer is a developer-authored, per-resource schema hook that runs at
// generation time, after azwise overlay and envelope defaults, before emission.
// It mutates the resolved type graph (Property flags/validators/defaults) and/or
// the envelope spec in place.
type Customizer func(*generator.ResourceDefinition)

// registry maps an ARM resource type (no API version) to its customizer.
var registry = map[string]Customizer{}

// Register attaches a schema customizer to an ARM resource type (use an armtypes
// constant, e.g. armtypes.StorageAccount). Call from register.go's init().
// Registering twice for the same type panics — a resource has exactly one
// customizer.
func Register(armType string, c Customizer) {
	if _, dup := registry[armType]; dup {
		panic("azapin: duplicate customizer for " + armType)
	}
	registry[armType] = c
}

// Apply runs the registered customizer (if any) for each definition. Call after
// generator.PostProcess so customizers see the azwise overlay and envelope
// defaults and have the final say before emission.
func Apply(defs []*generator.ResourceDefinition) {
	for _, def := range defs {
		if def == nil || def.Body == nil {
			continue
		}
		if c := registry[generator.ARMTypeOf(def)]; c != nil {
			c(def)
		}
	}
}
