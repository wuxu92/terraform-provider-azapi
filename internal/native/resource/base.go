// Package resource implements the runtime Terraform resource for native-generated
// static schemas. A single generic Base implements the framework interface and a
// unified payload-composition path (the mapper); per-resource Hooks customize it.
package resource

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/terraform-provider-azapi/internal/azure/azwise"
	"github.com/Azure/terraform-provider-azapi/internal/clients"
	"github.com/Azure/terraform-provider-azapi/internal/native/armjson"
	"github.com/Azure/terraform-provider-azapi/internal/native/generated"
	"github.com/Azure/terraform-provider-azapi/internal/native/generator"
	"github.com/Azure/terraform-provider-azapi/internal/native/mapper"
	"github.com/Azure/terraform-provider-azapi/internal/services/parse"
	"github.com/Azure/terraform-provider-azapi/utils"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Base is the generic resource implementation shared by every native static
// resource. It is constructed per Terraform name from a generated.Descriptor
// plus optional Hooks.
type Base struct {
	desc     generated.Descriptor
	hooks    *Hooks
	provider *clients.Client
}

// Interface assertions.
var (
	_ resource.Resource                   = &Base{}
	_ resource.ResourceWithConfigure      = &Base{}
	_ resource.ResourceWithModifyPlan     = &Base{}
	_ resource.ResourceWithValidateConfig = &Base{}
	_ resource.ResourceWithImportState    = &Base{}
)

// New builds a resource for the given generated Terraform name. The provider
// constructs one per entry in generated.Registry.
func New(name string) resource.Resource {
	d, ok := generated.Registry[name]
	if !ok {
		// Programmer error: provider iterates the registry, so this can't happen
		// in normal operation.
		panic(fmt.Sprintf("native: no generated descriptor for %q", name))
	}
	return &Base{
		desc:  d,
		hooks: hookRegistry[name],
	}
}

func (b *Base) typeAndVersion() string { return b.desc.ARMType + "@" + b.desc.APIVersion }

func (b *Base) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	// desc.Name already carries the provider prefix ("azapi_storage_account").
	resp.TypeName = req.ProviderTypeName + strings.TrimPrefix(b.desc.Name, "azapi")
}

func (b *Base) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if v, ok := req.ProviderData.(*clients.Client); ok {
		b.provider = v
	}
}

func (b *Base) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = b.composeSchema(ctx)
}

// composeSchema serves the generated schema, adding only the ctx-bound timeouts
// block. The generated schema is already complete: the body attributes plus the
// operational envelope (name, the per-resource parent reference, id) with all
// validators baked in at generation time (see internal/native/generator). The
// runtime never mutates the schema — timeouts is the sole addition because
// timeouts.Block needs a context the static generated function cannot hold.
func (b *Base) composeSchema(ctx context.Context) schema.Schema {
	s := b.desc.Schema()
	if s.Blocks == nil {
		s.Blocks = map[string]schema.Block{}
	}
	s.Blocks["timeouts"] = timeouts.Block(ctx, timeouts.Opts{Create: true, Read: true, Update: true, Delete: true})
	return s
}

// bodyGraph loads (and caches) the bicep body type graph for the mapper.
func (b *Base) bodyGraph() (*generator.Type, error) {
	return loadBody(b.desc.ARMType, b.desc.APIVersion)
}

func (b *Base) objectType(ctx context.Context) basetypes.ObjectType {
	return b.composeSchema(ctx).Type().(basetypes.ObjectType)
}

// ---------------------------------------------------------------------------
// Validate / ModifyPlan / Import
// ---------------------------------------------------------------------------

func (b *Base) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Schema-level validators (enum OneOf, ranges, length) run automatically.
	if b.hooks != nil && b.hooks.ValidateConfig != nil {
		b.hooks.ValidateConfig(ctx, req, resp)
	}
}

func (b *Base) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Schema-level RequiresReplace (from azwise ForceNew) is applied by the
	// framework. Conditional ForceNew (e.g. SKU zone migration) lives in a hook
	// that consults azwise.CheckForceNew.
	if b.hooks != nil && b.hooks.ModifyPlan != nil {
		b.hooks.ModifyPlan(ctx, req, resp)
	}
}

func (b *Base) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if b.provider == nil {
		resp.Diagnostics.AddError("Provider not configured", "the azapi provider was not configured")
		return
	}
	bt, err := b.bodyGraph()
	if err != nil {
		resp.Diagnostics.AddError("Failed to load resource schema", err.Error())
		return
	}

	id, err := parse.ResourceIDWithResourceType(req.ID, b.typeAndVersion())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Resource ID", err.Error())
		return
	}

	respBody, err := b.provider.ResourceClient.Get(ctx, id.AzureResourceId, id.ApiVersion, clients.DefaultRequestOptions())
	if err != nil {
		if utils.ResponseErrorWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to retrieve resource", err.Error())
		return
	}

	hc := &CrudCtx{Ctx: ctx, Client: b.provider, ID: id, Response: armjson.AsMap(respBody), Diags: &resp.Diagnostics}
	if b.hooks != nil && b.hooks.AfterRead != nil {
		b.hooks.AfterRead(hc)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	objType := b.objectType(ctx)
	envelope := map[string]attr.Value{
		"id":              types.StringValue(id.ID()),
		"name":            types.StringValue(id.Name),
		b.desc.ParentAttr: types.StringValue(id.ParentId),
	}
	stateObj, diags := mapper.Flatten(ctx, hc.Response, objType, bt, envelope)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(setState(ctx, &resp.State, stateObj)...)
}

// ---------------------------------------------------------------------------
// Create / Update / Read / Delete
// ---------------------------------------------------------------------------

func (b *Base) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	planObj, diags := b.rawToObject(ctx, req.Plan.Raw)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var to timeouts.Value
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("timeouts"), &to)...)
	if resp.Diagnostics.HasError() {
		return
	}
	b.put(ctx, planObj, true, to, &resp.State, &resp.Diagnostics)
}

func (b *Base) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	planObj, diags := b.rawToObject(ctx, req.Plan.Raw)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var to timeouts.Value
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("timeouts"), &to)...)
	if resp.Diagnostics.HasError() {
		return
	}
	b.put(ctx, planObj, false, to, &resp.State, &resp.Diagnostics)
}

// put is the unified create/update path.
func (b *Base) put(ctx context.Context, planObj types.Object, isNew bool, to timeouts.Value, state *tfsdk.State, diags *diag.Diagnostics) {
	if b.provider == nil {
		diags.AddError("Provider not configured", "the azapi provider was not configured")
		return
	}
	bt, err := b.bodyGraph()
	if err != nil {
		diags.AddError("Failed to load resource schema", err.Error())
		return
	}

	name := attrString(planObj, "name")
	parentID := attrString(planObj, b.desc.ParentAttr)
	id, err := parse.NewResourceID(name, parentID, b.typeAndVersion())
	if err != nil {
		diags.AddError("Invalid configuration", err.Error())
		return
	}

	op := "update"
	def := 30 * time.Minute
	if isNew {
		op = "create"
	}
	def = azwise.TimeoutDefault(b.desc.ARMType, b.desc.APIVersion, op, def)
	var timeout time.Duration
	var tdiags diag.Diagnostics
	if isNew {
		timeout, tdiags = to.Create(ctx, def)
	} else {
		timeout, tdiags = to.Update(ctx, def)
	}
	if diags.Append(tdiags...); diags.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Compose the ARM body and strip server-computed read-only fields.
	armBody := mapper.Expand(planObj, bt)
	azwise.StripComputedFields(b.desc.ARMType, b.desc.APIVersion, armBody)

	hc := &CrudCtx{Ctx: ctx, Client: b.provider, ID: id, Plan: planObj, Body: armBody, Diags: diags}
	b.runHook(hookBefore(b.hooks, isNew), hc)
	if diags.HasError() {
		return
	}

	if _, err := b.provider.ResourceClient.CreateOrUpdate(ctx, id.AzureResourceId, id.ApiVersion, armBody, clients.DefaultRequestOptions()); err != nil {
		diags.AddError("Failed to create/update resource", fmt.Errorf("creating/updating %s: %w", id.ID(), err).Error())
		return
	}

	respBody, err := b.provider.ResourceClient.Get(ctx, id.AzureResourceId, id.ApiVersion, clients.DefaultRequestOptions())
	if err != nil {
		diags.AddError("Failed to retrieve resource", fmt.Errorf("reading %s after write: %w", id.ID(), err).Error())
		return
	}

	hc.Response = armjson.AsMap(respBody)
	b.runHook(hookAfter(b.hooks, isNew), hc)
	if diags.HasError() {
		return
	}

	stateObj, fdiags := mapper.FlattenApplyInto(ctx, hc.Response, planObj, bt)
	if diags.Append(fdiags...); diags.HasError() {
		return
	}
	stateObj, sdiags := withString(ctx, stateObj, "id", id.ID())
	if diags.Append(sdiags...); diags.HasError() {
		return
	}
	// Apply results must be fully known — resolve any Optional+Computed unknowns
	// the response didn't populate to null.
	stateObj = mapper.ResolveUnknowns(ctx, stateObj).(types.Object)
	diags.Append(setState(ctx, state, stateObj)...)
}

func (b *Base) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if b.provider == nil {
		resp.Diagnostics.AddError("Provider not configured", "the azapi provider was not configured")
		return
	}
	stateObj, diags := b.rawToObject(ctx, req.State.Raw)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	bt, err := b.bodyGraph()
	if err != nil {
		resp.Diagnostics.AddError("Failed to load resource schema", err.Error())
		return
	}

	id, err := parse.ResourceIDWithResourceType(attrString(stateObj, "id"), b.typeAndVersion())
	if err != nil {
		resp.Diagnostics.AddError("Error parsing ID", err.Error())
		return
	}

	def := azwise.TimeoutDefault(b.desc.ARMType, b.desc.APIVersion, "read", 5*time.Minute)
	var to timeouts.Value
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("timeouts"), &to)...)
	if resp.Diagnostics.HasError() {
		return
	}
	timeout, tdiags := to.Read(ctx, def)
	if resp.Diagnostics.Append(tdiags...); resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	respBody, err := b.provider.ResourceClient.Get(ctx, id.AzureResourceId, id.ApiVersion, clients.DefaultRequestOptions())
	if err != nil {
		if utils.ResponseErrorWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to retrieve resource", err.Error())
		return
	}

	hc := &CrudCtx{Ctx: ctx, Client: b.provider, ID: id, State: stateObj, Response: armjson.AsMap(respBody), Diags: &resp.Diagnostics}
	if b.hooks != nil && b.hooks.AfterRead != nil {
		b.hooks.AfterRead(hc)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	newState, fdiags := mapper.FlattenInto(ctx, hc.Response, stateObj, bt)
	if resp.Diagnostics.Append(fdiags...); resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(setState(ctx, &resp.State, newState)...)
}

func (b *Base) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if b.hooks != nil && b.hooks.SkipARMDelete {
		// No ARM delete operation for this resource (e.g. a singleton child like
		// blobServices/default); the framework removes it from state on return.
		return
	}
	if b.provider == nil {
		resp.Diagnostics.AddError("Provider not configured", "the azapi provider was not configured")
		return
	}
	stateObj, diags := b.rawToObject(ctx, req.State.Raw)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := parse.ResourceIDWithResourceType(attrString(stateObj, "id"), b.typeAndVersion())
	if err != nil {
		resp.Diagnostics.AddError("Error parsing ID", err.Error())
		return
	}

	def := azwise.TimeoutDefault(b.desc.ARMType, b.desc.APIVersion, "delete", 30*time.Minute)
	var to timeouts.Value
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("timeouts"), &to)...)
	if resp.Diagnostics.HasError() {
		return
	}
	timeout, tdiags := to.Delete(ctx, def)
	if resp.Diagnostics.Append(tdiags...); resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if b.hooks != nil && b.hooks.BeforeDelete != nil {
		hc := &CrudCtx{Ctx: ctx, Client: b.provider, ID: id, State: stateObj, Diags: &resp.Diagnostics}
		b.hooks.BeforeDelete(hc)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if _, err := b.provider.ResourceClient.Delete(ctx, id.AzureResourceId, id.ApiVersion, clients.DefaultRequestOptions()); err != nil && !utils.ResponseErrorWasNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete resource", fmt.Errorf("deleting %s: %w", id.ID(), err).Error())
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func (b *Base) rawToObject(ctx context.Context, raw tftypes.Value) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics
	objType := b.objectType(ctx)
	v, err := objType.ValueFromTerraform(ctx, raw)
	if err != nil {
		diags.AddError("Failed to decode configuration", err.Error())
		return types.ObjectNull(objType.AttributeTypes()), diags
	}
	obj, ok := v.(types.Object)
	if !ok {
		diags.AddError("Failed to decode configuration", "root value is not an object")
		return types.ObjectNull(objType.AttributeTypes()), diags
	}
	return obj, diags
}

func (b *Base) runHook(h func(*CrudCtx), hc *CrudCtx) {
	if h != nil {
		h(hc)
	}
}

func hookBefore(h *Hooks, isNew bool) func(*CrudCtx) {
	if h == nil {
		return nil
	}
	if isNew {
		return h.BeforeCreate
	}
	return h.BeforeUpdate
}

func hookAfter(h *Hooks, isNew bool) func(*CrudCtx) {
	if h == nil {
		return nil
	}
	if isNew {
		return h.AfterCreate
	}
	return h.AfterUpdate
}

func attrString(obj types.Object, name string) string {
	v, ok := obj.Attributes()[name]
	if !ok {
		return ""
	}
	s, ok := v.(types.String)
	if !ok || s.IsNull() || s.IsUnknown() {
		return ""
	}
	return s.ValueString()
}

func withString(ctx context.Context, obj types.Object, name, val string) (types.Object, diag.Diagnostics) {
	attrTypes := obj.AttributeTypes(ctx)
	values := make(map[string]attr.Value, len(attrTypes))
	for k, v := range obj.Attributes() {
		values[k] = v
	}
	values[name] = types.StringValue(val)
	return types.ObjectValue(attrTypes, values)
}

func setState(ctx context.Context, st *tfsdk.State, obj types.Object) diag.Diagnostics {
	var diags diag.Diagnostics
	raw, err := obj.ToTerraformValue(ctx)
	if err != nil {
		diags.AddError("Failed to encode state", err.Error())
		return diags
	}
	st.Raw = raw
	return diags
}
