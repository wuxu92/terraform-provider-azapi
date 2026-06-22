package resource

import (
	"context"

	"github.com/Azure/terraform-provider-azapi/internal/azure/azwise"
	"github.com/Azure/terraform-provider-azapi/internal/native/generated"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Storage account overlay. The generated schema already carries declarative
// ForceNew (RequiresReplace) from the azwise overlay; this hook adds the
// *conditional* SKU zone-migration rule that a static plan modifier can't
// express — replacement is required only when migrating between zonal and
// non-zonal SKUs, which azwise.CheckForceNew encodes.
func init() {
	RegisterHooks("azapi_storage_account", &Hooks{
		ModifyPlan: storageAccountModifyPlan,
	})
}

func storageAccountModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Only relevant for updates (both prior state and plan present).
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var oldName, newName types.String
	skuName := path.Root("sku").AtName("name")
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, skuName, &oldName)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, skuName, &newName)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Minimal bodies isolate the conditional SKU rule: the declarative ForceNew
	// paths (already enforced as schema-level RequiresReplace) compare nil == nil
	// here, so only the SKU zone-migration logic in the storage knowledge fires.
	oldBody := map[string]interface{}{"sku": map[string]interface{}{"name": oldName.ValueString()}}
	newBody := map[string]interface{}{"sku": map[string]interface{}{"name": newName.ValueString()}}

	// Look up the ARM type + API version from the shipped descriptor so this stays
	// correct when the generator rolls the resource forward to a newer version.
	d := generated.Registry["azapi_storage_account"]
	if azwise.CheckForceNew(d.ARMType, d.APIVersion, oldBody, newBody) {
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root("sku"))
	}
}
