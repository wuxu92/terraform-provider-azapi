package services

import (
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

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
	// Relational carries resource-level cross-property constraints (ConflictsWith /
	// RequiredWith / ExactlyOneOf / AtLeastOneOf) lowered from azwise at generation
	// time. The runtime resource layer builds framework ConfigValidators from these.
	Relational []RelationalConstraint
	// Timeouts carries the per-operation timeout defaults baked in at generation
	// time from azwise (source of truth). A zero field means "no override" and the
	// runtime falls back to its built-in default for that operation.
	Timeouts Timeouts
}

// Timeouts holds the per-operation timeout defaults for a generated resource.
// A zero value for an operation means the runtime uses its built-in fallback.
type Timeouts struct {
	Create time.Duration
	Read   time.Duration
	Update time.Duration
	Delete time.Duration
}

// RelationalKind identifies a resource-level cross-property constraint kind.
type RelationalKind int

const (
	// ConflictsWith: subject (Paths[0]) must not be set together with any of Paths[1:].
	ConflictsWith RelationalKind = iota
	// RequiredWith: when subject (Paths[0]) is set, every path in Paths[1:] must be set.
	RequiredWith
	// ExactlyOneOf: exactly one of Paths must be set.
	ExactlyOneOf
	// AtLeastOneOf: at least one of Paths must be set.
	AtLeastOneOf
)

// RelationalConstraint is a cross-property constraint over Terraform attribute
// paths, each a slice of snake_case segments absolute from the schema root (e.g.
// {"properties", "curve_name"}). For ConflictsWith and RequiredWith, Paths[0] is
// the subject. Lowered from azwise by the generator; consumed by the runtime
// resource layer to build framework ConfigValidators.
type RelationalConstraint struct {
	Kind    RelationalKind
	Paths   [][]string
	Message string
}

// Registry maps Terraform resource names to their descriptors.
var Registry = map[string]Descriptor{}

// Register adds a descriptor to the registry.
func Register(d Descriptor) { Registry[d.Name] = d }
