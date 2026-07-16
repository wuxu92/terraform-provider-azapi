package kusto

import (
	"context"
	"fmt"
	"strings"

	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func init() {
	nativeresource.RegisterHooks(KustoCluster.Name, &nativeresource.Hooks{
		ValidateConfig: kustoClusterValidateConfig,
	})
}

// kustoClusterValidateConfig enforces the two Microsoft.Kusto/clusters SKU cross-field
// rules ARM applies but the per-field schema validators cannot express, so the
// practitioner sees them at plan time instead of as an opaque ARM error at apply:
//
//   - Tier is determined by the SKU name. AzureRM derives it from the name's first
//     underscore-separated token (expandKustoClusterSku): a name beginning with
//     "Standard" is the Standard tier; the two "Dev(No SLA)_*" names are the Basic
//     tier. This holds the user-supplied sku.tier to that same mapping.
//   - Capacity (instance count) bounds depend on the tier: a Dev/Basic cluster is
//     single-instance (exactly 1); a Standard cluster scales between 2 and 1000. ARM
//     rejects an out-of-range count ("Invalid SKU capacity provided, the SKU
//     Standard_L8s_v3 should be between 2 and 1000"); this reproduces that check
//     client-side, tightening the schema's coarse Between(1, 1000) to the SKU's range.
//
// ValidateConfig runs on the raw config where computed-only values the user omitted
// are null (not unknown), so each attribute is guarded for null/unknown before use.
func kustoClusterValidateConfig(ctx context.Context, req fwresource.ValidateConfigRequest, resp *fwresource.ValidateConfigResponse) {
	skuPath := path.Root("sku")

	var name types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, skuPath.AtName("name"), &name)...)
	if resp.Diagnostics.HasError() || name.IsNull() || name.IsUnknown() {
		return
	}
	skuName := name.ValueString()
	expectedTier, minCapacity, maxCapacity := kustoSkuConstraints(skuName)

	// Rule 1: tier must match the name-derived tier.
	var tier types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, skuPath.AtName("tier"), &tier)...)
	if !resp.Diagnostics.HasError() && !tier.IsNull() && !tier.IsUnknown() && tier.ValueString() != expectedTier {
		resp.Diagnostics.AddAttributeError(
			skuPath.AtName("tier"),
			"Invalid Kusto cluster SKU tier",
			fmt.Sprintf("sku.tier %q does not match sku.name %q; the %q SKU is the %q tier.", tier.ValueString(), skuName, skuName, expectedTier),
		)
	}

	// Rule 2: capacity (instance count) must be within the tier's bounds.
	var capacity types.Int64
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, skuPath.AtName("capacity"), &capacity)...)
	if !resp.Diagnostics.HasError() && !capacity.IsNull() && !capacity.IsUnknown() {
		if c := capacity.ValueInt64(); c < minCapacity || c > maxCapacity {
			resp.Diagnostics.AddAttributeError(
				skuPath.AtName("capacity"),
				"Invalid Kusto cluster SKU capacity",
				fmt.Sprintf("Invalid SKU capacity provided, the SKU %s should be between %d and %d.", skuName, minCapacity, maxCapacity),
			)
		}
	}
}

// kustoSkuConstraints returns the tier and the [min, max] instance-count bounds implied
// by a Kusto cluster SKU name. The two "Dev(No SLA)_*" SKUs are single-instance Basic
// clusters; every other (Standard-prefixed) SKU is a Standard cluster scaling from 2 to
// 1000 instances. The sku.name OneOf validator restricts input to those two families,
// so the Standard-prefix test cleanly partitions the enum.
func kustoSkuConstraints(skuName string) (tier string, minCapacity, maxCapacity int64) {
	if strings.HasPrefix(skuName, "Standard") {
		return "Standard", 2, 1000
	}
	return "Basic", 1, 1
}
