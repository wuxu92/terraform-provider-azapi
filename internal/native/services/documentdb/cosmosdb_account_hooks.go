package documentdb

import (
	"context"

	"github.com/Azure/terraform-provider-azapi/internal/azure/azwise"
	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Cosmos DB account hooks add the value-conditional replacements a static schema
// plan modifier cannot express. The generated schema already carries the declarative
// ForceNew rules from azwise as schema-level RequiresReplace; this hook handles the
// change-dependent decisions encoded by azwise.CheckForceNew (analytical storage
// cannot be disabled in place; the backup policy can only move Periodic->Continuous,
// so a Continuous->Periodic change forces replacement).
func init() {
	nativeresource.RegisterHooks(CosmosdbAccount.Name, &nativeresource.Hooks{
		ModifyPlan: cosmosDBModifyPlan,
	})
}

func cosmosDBModifyPlan(ctx context.Context, req fwresource.ModifyPlanRequest, resp *fwresource.ModifyPlanResponse) {
	// Only relevant for updates (both prior state and plan present).
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	// Rule 1: analytical storage cannot be disabled (true -> false) in place.
	asPath := path.Root("properties").AtName("enable_analytical_storage")
	var oldAS, newAS types.Bool
	if !req.State.GetAttribute(ctx, asPath, &oldAS).HasError() &&
		!req.Plan.GetAttribute(ctx, asPath, &newAS).HasError() &&
		!oldAS.IsUnknown() && !newAS.IsUnknown() {
		oldBody := map[string]interface{}{"properties": map[string]interface{}{"enableAnalyticalStorage": oldAS.ValueBool()}}
		newBody := map[string]interface{}{"properties": map[string]interface{}{"enableAnalyticalStorage": newAS.ValueBool()}}
		if azwise.CheckForceNew(CosmosdbAccount.ARMType, CosmosdbAccount.APIVersion, oldBody, newBody) {
			resp.RequiresReplace = append(resp.RequiresReplace, asPath)
		}
	}

	// Rule 2: backup policy can only migrate Periodic -> Continuous; the reverse
	// forces replacement. The discriminator (properties.backupPolicy.type) is not a
	// settable attribute — it is synthesized from whichever variant block is set — so
	// derive it from the presence of the periodic/continuous blocks in state vs plan.
	oldType := cosmosDBBackupType(ctx, req.State.GetAttribute)
	newType := cosmosDBBackupType(ctx, req.Plan.GetAttribute)
	if oldType != "" && newType != "" {
		backupPath := path.Root("properties").AtName("backup_policy")
		oldBody := map[string]interface{}{"properties": map[string]interface{}{"backupPolicy": map[string]interface{}{"type": oldType}}}
		newBody := map[string]interface{}{"properties": map[string]interface{}{"backupPolicy": map[string]interface{}{"type": newType}}}
		if azwise.CheckForceNew(CosmosdbAccount.ARMType, CosmosdbAccount.APIVersion, oldBody, newBody) {
			resp.RequiresReplace = append(resp.RequiresReplace, backupPath)
		}
	}
}

// getAttrFunc is the shared shape of tfsdk.State.GetAttribute / tfsdk.Plan.GetAttribute.
type getAttrFunc func(context.Context, path.Path, interface{}) diag.Diagnostics

// cosmosDBBackupType derives the ARM backupPolicy discriminator value ("Continuous"
// / "Periodic") from whichever variant block under properties.backup_policy is set.
// Returns "" when neither is configured (Azure's default policy applies and there is
// no explicit change to evaluate).
func cosmosDBBackupType(ctx context.Context, get getAttrFunc) string {
	base := path.Root("properties").AtName("backup_policy")
	var continuous, periodic types.Object
	if !get(ctx, base.AtName("continuous"), &continuous).HasError() && !continuous.IsNull() && !continuous.IsUnknown() {
		return "Continuous"
	}
	if !get(ctx, base.AtName("periodic"), &periodic).HasError() && !periodic.IsNull() && !periodic.IsUnknown() {
		return "Periodic"
	}
	return ""
}
