package services

import "github.com/hashicorp/terraform-plugin-framework/resource/schema"

// Descriptor is the static, dependency-light description of one generated
// resource: its Terraform name, the ARM type + API version it targets, and the
// generated body schema. Each generated file registers one via init().
//
// The runtime resource layer (internal/native/resource) consumes these to build
// working Terraform resources; keeping framework/client dependencies out of this
// package lets the generated files stay pure schema.
type Descriptor struct {
	Name       string // Terraform resource name, e.g. "azapi_storage_account"
	ARMType    string // "Microsoft.Storage/storageAccounts"
	APIVersion string // "2025-01-01"
	Schema     func() schema.Schema
	// WritableScopes is the bicep scope bitmask (Tenant=1, ManagementGroup=2,
	// Subscription=4, ResourceGroup=8, Extension=16). It is retained as metadata;
	// the parent-reference attribute name and validator are now baked into the
	// generated Schema at generation time (see ParentAttr).
	WritableScopes int
	// ParentAttr is the generated operational-envelope parent-reference attribute
	// name (e.g. "resource_group_id", "storage_account_id", "parent_id"). The
	// runtime reads it to compose/parse the ARM resource ID; the schema validator
	// for it is already baked into Schema.
	ParentAttr string
}

// Registry maps Terraform resource names to their descriptors.
var Registry = map[string]Descriptor{}

// Register adds a descriptor to the registry.
func Register(d Descriptor) { Registry[d.Name] = d }
