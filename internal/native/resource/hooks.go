package resource

import (
	"context"

	"github.com/Azure/terraform-provider-azapi/internal/clients"
	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// CrudCtx is passed to lifecycle hooks (see Hooks). It carries everything a hook
// may inspect or mutate during one CRUD operation, but which fields are live
// (populated and meaningful) versus null depends on the hook point. Reading a
// field that is null at the current hook point yields the zero value, not an
// error — consult the matrix below (and the per-field notes and Hooks field
// docs, which are the authoritative contract) before reading one.
//
// Field liveness by hook point (Ctx, Client, ID, Diags are always live):
//
//	hook point   | Plan | State | Body | Response
//	-------------+------+-------+------+---------
//	BeforeCreate | live | null  | live | null
//	AfterCreate  | live | null  | live | live
//	BeforeUpdate | live | null  | live | null
//	AfterUpdate  | live | null  | live | live
//	BeforeRead   | null | live  | null | null
//	AfterRead    | null | live  | null | live
//	BeforeDelete | null | live  | null | null
//
// A Body mutation reaches ARM only from a Before* create/update hook: the
// CreateOrUpdate PUT reads Body immediately after BeforeCreate/BeforeUpdate
// returns. By an After* hook the PUT is done, so mutating Body is a no-op for
// ARM; massage Response instead — later state mapping flattens Response, not Body.
type CrudCtx struct {
	Ctx    context.Context
	Client *clients.Client
	ID     parse.ResourceId
	// Plan is the typed configuration plan. Live in BeforeCreate/AfterCreate/
	// BeforeUpdate/AfterUpdate; null in BeforeRead/AfterRead/BeforeDelete.
	Plan types.Object
	// State is the typed prior Terraform state. Live in BeforeRead/AfterRead/
	// BeforeDelete; null in BeforeCreate/AfterCreate/BeforeUpdate/AfterUpdate
	// (the create/update path composes Body from Plan and never populates State).
	State types.Object
	// Body is the ARM request body being composed, already stripped of
	// server-computed read-only fields. Live in BeforeCreate/AfterCreate/
	// BeforeUpdate/AfterUpdate; null in BeforeRead/AfterRead/BeforeDelete.
	// Mutating it reaches ARM only from a Before* hook (see the type doc).
	Body map[string]interface{}
	// Response is the ARM GET response as a map. Live in AfterCreate/AfterUpdate
	// (the post-PUT read-back) and AfterRead (the read GET); null in every
	// Before* hook. It is the object flattened into Terraform state, so an
	// After* hook massages Response, not Body.
	Response map[string]interface{}
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
	// BeforeCreate fires on create, after the pre-create existence check and after
	// the ARM Body is composed from Plan and stripped of server-computed fields,
	// but before the CreateOrUpdate PUT. Live: Plan, Body (mutations reach ARM).
	// Null: State, Response.
	BeforeCreate func(*CrudCtx)
	// AfterCreate fires on create, after the CreateOrUpdate PUT and the read-back
	// GET, but before Response is flattened into Terraform state. Live: Plan, Body
	// (no longer sent to ARM), Response (massage this). Null: State.
	AfterCreate func(*CrudCtx)
	// BeforeUpdate fires on update, after the ARM Body is composed from Plan and
	// stripped, but before the CreateOrUpdate PUT (update skips the existence
	// check). Live: Plan, Body (mutations reach ARM). Null: State, Response.
	BeforeUpdate func(*CrudCtx)
	// AfterUpdate fires on update, after the CreateOrUpdate PUT and the read-back
	// GET, but before Response is flattened into state. Live: Plan, Body (no
	// longer sent to ARM), Response (massage this). Null: State.
	AfterUpdate func(*CrudCtx)
	// BeforeRead fires on read, after the timeout is resolved but before the ARM
	// GET — use it for preflight/guard (append a diagnostic to short-circuit the
	// read). Live: State. Null: Plan, Body, Response (the GET has not run yet).
	BeforeRead func(*CrudCtx)
	// AfterRead fires on read, after the ARM GET succeeds but before Response is
	// flattened into state — massage Response here. Live: State, Response. Null:
	// Plan, Body. Not reached on a 404 (the resource is removed from state).
	AfterRead func(*CrudCtx)
	// BeforeDelete fires on delete, after the timeout is resolved but before the
	// ARM DELETE. Live: State. Null: Plan, Body, Response. It does NOT run on the
	// Singleton reset path (see Singleton), so do not rely on it for a singleton
	// default.
	BeforeDelete func(*CrudCtx)

	// Singleton, when non-nil, marks a resource that always exists as a fixed-named
	// default child that ARM neither creates nor deletes on its own — e.g.
	// Microsoft.Storage/storageAccounts/blobServices/default, which comes into being
	// with its parent storage account and supports only GET and PUT (CreateOrUpdate).
	// Create/Update skip the pre-create existence check (a GET always succeeds, so a
	// create is really an in-place update), and Delete resets the resource to
	// Singleton.DefaultBody via PUT instead of issuing an ARM DELETE (which would 405)
	// before the framework removes it from Terraform state. That reset path bypasses
	// BeforeDelete: BeforeDelete does NOT run for a Singleton resource. BeforeCreate/
	// AfterCreate/BeforeUpdate/AfterUpdate/BeforeRead/AfterRead still run as usual.
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
