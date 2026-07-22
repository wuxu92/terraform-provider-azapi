package storage

import (
	"context"
	"strings"

	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
	"github.com/Azure/terraform-provider-azapi/internal/native/services"
	"github.com/hashicorp/terraform-plugin-framework/path"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Storage account hooks add the conditional replacement rules that a static schema
// plan modifier cannot express. Generated schema already carries the declarative
// ForceNew rules as schema-level RequiresReplace; this hook handles the
// value-dependent decisions (SKU zone migration, account_kind migration,
// large-file-share disablement) mirroring AzureRM's storage account CustomizeDiff.
func init() {
	nativeresource.RegisterHooks(StorageAccount.Name, &nativeresource.Hooks{
		ModifyPlan: storageAccountModifyPlan,
		// Customer-managed key encryption requires the account to carry the referenced
		// user-assigned identity.
		Relational: []services.RelationalConstraint{
			{Kind: services.RequiredWith, Paths: []string{"properties.encryption.identity.user_assigned_identity", "identity.user_assigned_identities"}, Message: "customer-managed key encryption requires the account to carry the referenced user-assigned identity"},
		},
	})
}

// condReplaceRule pairs the attribute whose change is evaluated with the attribute
// path marked as forcing replacement, plus a predicate deciding whether the value
// transition requires replacement.
type condReplaceRule struct {
	attr    path.Path
	marker  path.Path
	changed func(oldVal, newVal string) bool
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
			attr:    path.Root("sku").AtName("name"),
			marker:  path.Root("sku"),
			changed: skuZoneMigration,
		},
		{
			// account_kind migration — only Storage -> StorageV2 is allowed in place.
			attr:    path.Root("kind"),
			marker:  path.Root("kind"),
			changed: accountKindRequiresReplace,
		},
		{
			// large_file_shares_state cannot be disabled once enabled.
			attr:    path.Root("properties").AtName("large_file_shares_state"),
			marker:  path.Root("properties").AtName("large_file_shares_state"),
			changed: largeFileShareDisabled,
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
		if r.changed(oldVal.ValueString(), newVal.ValueString()) {
			resp.RequiresReplace = append(resp.RequiresReplace, r.marker)
		}
	}
}

// skuZoneMigration reports whether the account_replication_type (the suffix of
// sku.name, e.g. "ZRS" in "Standard_ZRS") crosses the zonal/non-zonal boundary,
// which requires recreation. Matches AzureRM ForceNewIfChange("account_replication_type")
// on the replication suffix alone — tier-agnostic, so Premium_LRS <-> Premium_ZRS is
// covered as well as Standard.
func skuZoneMigration(oldSku, newSku string) bool {
	oldSku = strings.ToUpper(oldSku)
	newSku = strings.ToUpper(newSku)
	if oldSku == "" || newSku == "" || oldSku == newSku {
		return false
	}
	oldRep := replicationType(oldSku)
	newRep := replicationType(newSku)
	zonal := map[string]bool{"ZRS": true, "GZRS": true, "RAGZRS": true}
	nonZonal := map[string]bool{"LRS": true, "GRS": true, "RAGRS": true}
	return (zonal[oldRep] && nonZonal[newRep]) || (nonZonal[oldRep] && zonal[newRep])
}

// replicationType returns the redundancy suffix of an upper-cased sku.name, e.g.
// "STANDARD_RAGZRS" -> "RAGZRS", "PREMIUMV2_LRS" -> "LRS".
func replicationType(sku string) string {
	if i := strings.LastIndex(sku, "_"); i >= 0 {
		return sku[i+1:]
	}
	return sku
}

// accountKindRequiresReplace reports whether an account_kind change forces
// replacement. AzureRM permits exactly one in-place migration, Storage -> StorageV2;
// every other kind change requires recreation (storage_account_resource.go:1103).
func accountKindRequiresReplace(oldKind, newKind string) bool {
	if oldKind == "" || oldKind == newKind {
		return false
	}
	return oldKind != "Storage" && newKind != "StorageV2"
}

// largeFileShareDisabled reports whether large file shares are being turned off.
// Once enabled the feature cannot be disabled in place; Enabled -> anything else
// forces replacement (storage_account_resource.go:1116).
func largeFileShareDisabled(oldLFS, newLFS string) bool {
	return strings.EqualFold(oldLFS, "Enabled") && !strings.EqualFold(newLFS, "Enabled")
}
