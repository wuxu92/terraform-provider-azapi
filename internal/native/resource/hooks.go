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
// Hooks (or nil field) means "use the base behavior". Overlay files register
// Hooks via RegisterHooks.
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

	// SkipARMDelete marks a resource that has no ARM delete operation — e.g. a
	// singleton child such as Microsoft.Storage/storageAccounts/blobServices/default
	// that exists for the lifetime of its parent and would 405 on DELETE. When true,
	// Delete removes the resource from Terraform state without calling ARM.
	SkipARMDelete bool

	// Full overrides of the framework lifecycle methods. When set, the base
	// invokes these after its own default work (for ModifyPlan/ValidateConfig)
	// or instead of nothing (ImportState extension).
	ValidateConfig func(context.Context, resource.ValidateConfigRequest, *resource.ValidateConfigResponse)
	ModifyPlan     func(context.Context, resource.ModifyPlanRequest, *resource.ModifyPlanResponse)
}

// hookRegistry holds hand-written overlays keyed by Terraform resource name.
var hookRegistry = map[string]*Hooks{}

// RegisterHooks attaches customization hooks to a generated resource. Call from
// an overlay file's init().
func RegisterHooks(name string, h *Hooks) { hookRegistry[name] = h }
