package resource

import (
	"context"

	"github.com/Azure/terraform-provider-azapi/internal/clients"
	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CrudCtx is passed to lifecycle hooks. It exposes everything a hook may need to
// inspect or mutate during a single CRUD operation.
type CrudCtx struct {
	Ctx      context.Context
	Client   *clients.Client
	ID       parse.ResourceId
	Plan     types.Object           // typed plan (create/update); null otherwise
	State    types.Object           // typed prior state (read/update/delete); null otherwise
	Body     map[string]interface{} // ARM body being composed — mutate in Before* hooks
	Response map[string]interface{} // ARM GET response — read in After* hooks
	Diags    *diag.Diagnostics
}

// Hooks holds optional per-resource customization of runtime BEHAVIOR only. A nil
// Hooks (or nil field) means "use the base behavior". Generated service packages
// register hand-written hooks via RegisterHooks from <resource>_hooks.go files.
//
// Schema customization (attribute validators, defaults, the parent-reference
// name) is NOT done here — it is baked into the generated schema at generation
// time via generator customizers (see internal/native/generator/customizers), so the
// runtime never mutates the schema.
type Hooks struct {
	BeforeCreate func(*CrudCtx)
	AfterCreate  func(*CrudCtx)
	BeforeUpdate func(*CrudCtx)
	AfterUpdate  func(*CrudCtx)
	BeforeRead   func(*CrudCtx)
	AfterRead    func(*CrudCtx)
	BeforeDelete func(*CrudCtx)

	// Singleton, when non-nil, marks a resource that always exists as a fixed-named
	// default child that ARM neither creates nor deletes on its own — e.g.
	// Microsoft.Storage/storageAccounts/blobServices/default, which comes into being
	// with its parent storage account and supports only GET and PUT (CreateOrUpdate).
	// Create/Update skip the pre-create existence check (a GET always succeeds, so a
	// create is really an in-place update), and Delete resets the resource to
	// Singleton.DefaultBody via PUT instead of issuing an ARM DELETE (which would 405)
	// before the framework removes it from Terraform state.
	Singleton *SingletonDefault

	// Full overrides of the framework lifecycle methods. When set, the base
	// invokes these after its own default work (for ModifyPlan/ValidateConfig)
	// or instead of nothing (ImportState extension).
	ValidateConfig func(context.Context, resource.ValidateConfigRequest, *resource.ValidateConfigResponse)
	ModifyPlan     func(context.Context, resource.ModifyPlanRequest, *resource.ModifyPlanResponse)
}

// SingletonDefault configures a Hooks.Singleton resource: a fixed-named default
// child that ARM neither creates nor deletes independently. It only supplies the
// baseline body used to reset the resource when Terraform destroys it.
type SingletonDefault struct {
	// DefaultBody is the ARM request body a Terraform destroy PUTs to return the
	// resource to its baseline state (ARM has no DELETE for it, so removal is a
	// reset). For blobServices/default this disables the static website, clears
	// CORS rules and disables delete retention.
	DefaultBody map[string]interface{}
}

// hookRegistry holds hand-written runtime hooks keyed by Terraform resource name.
var hookRegistry = map[string]*Hooks{}

// RegisterHooks attaches customization hooks to a generated resource. Call from
// a generated service package's <resource>_hooks.go init().
func RegisterHooks(name string, h *Hooks) { hookRegistry[name] = h }
