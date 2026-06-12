package generated

import "github.com/hashicorp/terraform-plugin-framework/resource/schema"

// SchemaFunc is a function that returns a compiled Terraform resource schema.
type SchemaFunc func() schema.Schema

// Registry maps Terraform resource names to their schema functions.
// Each generated file registers itself via init().
var Registry = map[string]SchemaFunc{}

// Register adds a schema function to the registry.
func Register(name string, fn SchemaFunc) {
	Registry[name] = fn
}
