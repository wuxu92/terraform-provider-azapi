package storage

import (
	"context"

	"github.com/Azure/terraform-provider-azapi/internal/azure/azwise"
	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Storage account hooks add the conditional replacement rules that a static schema
// plan modifier cannot express. Generated schema already carries the declarative
// ForceNew rules from azwise as schema-level RequiresReplace; this hook handles the
// value-dependent decisions encoded by azwise.CheckForceNew (SKU zone migration,
// account_kind migration, large-file-share disablement).
func init() {
	nativeresource.RegisterHooks(StorageAccount.Name, &nativeresource.Hooks{
		ModifyPlan: storageAccountModifyPlan,
	})
}

// condReplaceRule pairs the attribute whose change is evaluated with the attribute
// path marked as forcing replacement, plus a wrapper that lifts the read value into
// an isolated ARM body so only this rule's logic fires in azwise.CheckForceNew.
type condReplaceRule struct {
	attr   path.Path
	marker path.Path
	body   func(value string) map[string]interface{}
}

func storageAccountModifyPlan(ctx context.Context, req fwresource.ModifyPlanRequest, resp *fwresource.ModifyPlanResponse) {
	// Only relevant for updates (both prior state and plan present).
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	rules := []condReplaceRule{
		{
			// SKU zone migration (e.g. Standard_LRS <-> Standard_ZRS). The whole sku
			// block is marked because account_replication_type is encoded in sku.name.
			attr:   path.Root("sku").AtName("name"),
			marker: path.Root("sku"),
			body:   func(v string) map[string]interface{} { return map[string]interface{}{"sku": map[string]interface{}{"name": v}} },
		},
		{
			// account_kind migration — only Storage -> StorageV2 is allowed in place.
			attr:   path.Root("kind"),
			marker: path.Root("kind"),
			body:   func(v string) map[string]interface{} { return map[string]interface{}{"kind": v} },
		},
		{
			// large_file_shares_state cannot be disabled once enabled.
			attr:   path.Root("properties").AtName("large_file_shares_state"),
			marker: path.Root("properties").AtName("large_file_shares_state"),
			body: func(v string) map[string]interface{} {
				return map[string]interface{}{"properties": map[string]interface{}{"largeFileSharesState": v}}
			},
		},
	}

	for _, r := range rules {
		var oldVal, newVal types.String
		// A read failure means the attribute isn't meaningfully present in this
		// config shape (e.g. a null parent block); skip rather than abort the plan.
		if req.State.GetAttribute(ctx, r.attr, &oldVal).HasError() {
			continue
		}
		if req.Plan.GetAttribute(ctx, r.attr, &newVal).HasError() {
			continue
		}
		if oldVal.IsUnknown() || newVal.IsUnknown() {
			continue
		}
		oldBody := r.body(oldVal.ValueString())
		newBody := r.body(newVal.ValueString())
		if azwise.CheckForceNew(StorageAccount.ARMType, StorageAccount.APIVersion, oldBody, newBody) {
			resp.RequiresReplace = append(resp.RequiresReplace, r.marker)
		}
	}
}
