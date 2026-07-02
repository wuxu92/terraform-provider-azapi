package storage

import (
	"context"

	"github.com/Azure/terraform-provider-azapi/internal/azure/azwise"
	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Storage account hooks add the conditional SKU zone-migration rule that a
// static schema plan modifier cannot express. Generated schema already carries
// declarative ForceNew rules from azwise; this hook only handles the dynamic
// replacement decision encoded by azwise.CheckForceNew.
func init() {
	nativeresource.RegisterHooks(StorageAccount.Name, &nativeresource.Hooks{
		ModifyPlan: storageAccountModifyPlan,
	})
}

func storageAccountModifyPlan(ctx context.Context, req fwresource.ModifyPlanRequest, resp *fwresource.ModifyPlanResponse) {
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

	if azwise.CheckForceNew(StorageAccount.ARMType, StorageAccount.APIVersion, oldBody, newBody) {
		resp.RequiresReplace = append(resp.RequiresReplace, path.Root("sku"))
	}
}
