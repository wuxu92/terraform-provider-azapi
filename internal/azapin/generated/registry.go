package generated

import "github.com/hashicorp/terraform-plugin-framework/resource/schema"

// Descriptor is the static, dependency-light description of one generated
// resource: its Terraform name, the ARM type + API version it targets, and the
// generated body schema. Each generated file registers one via init().
//
// The runtime resource layer (internal/azapin/resource) consumes these to build
// working Terraform resources; keeping framework/client dependencies out of this
// package lets the generated files stay pure schema.
type Descriptor struct {
	Name       string // Terraform resource name, e.g. "azapi_storage_account"
	ARMType    string // "Microsoft.Storage/storageAccounts"
	APIVersion string // "2025-01-01"
	Schema     func() schema.Schema
}

// Registry maps Terraform resource names to their descriptors.
var Registry = map[string]Descriptor{}

// Register adds a descriptor to the registry.
func Register(d Descriptor) { Registry[d.Name] = d }
