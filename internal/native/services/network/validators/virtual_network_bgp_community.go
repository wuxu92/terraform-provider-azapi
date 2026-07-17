package validators

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

type virtualNetworkBgpCommunity struct{}

func (virtualNetworkBgpCommunity) Description(context.Context) string {
	return `must be a BGP community in "asn:community" notation, where each value is in the range (0, 65535) (e.g. "12076:20000")`
}

func (v virtualNetworkBgpCommunity) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (virtualNetworkBgpCommunity) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	str := req.ConfigValue
	if str.IsUnknown() || str.IsNull() {
		return
	}
	value := str.ValueString()

	segments := strings.Split(value, ":")
	if len(segments) != 2 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid BGP community",
			fmt.Sprintf(`invalid notation of bgp community: expected "asn:community" but got %q`, value),
		)
		return
	}

	asn, err := strconv.Atoi(segments[0])
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid BGP community",
			fmt.Sprintf("converting asn %q: %v", segments[0], err),
		)
		return
	}
	if asn <= 0 || asn >= 65535 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid BGP community",
			fmt.Sprintf("asn %d exceeds range: (0, 65535)", asn),
		)
		return
	}

	comm, err := strconv.Atoi(segments[1])
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid BGP community",
			fmt.Sprintf("converting community value %q: %v", segments[1], err),
		)
		return
	}
	if comm <= 0 || comm >= 65535 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid BGP community",
			fmt.Sprintf("community value %d exceeds range: (0, 65535)", comm),
		)
	}
}

// VirtualNetworkBgpCommunity returns a string validator for a virtual network's
// bgp_community (properties.bgpCommunities.virtualNetworkCommunity): "asn:community"
// notation with each value in the open range (0, 65535). Ported from AzureRM's
// validate.VirtualNetworkBgpCommunity. Resource-specific to
// Microsoft.Network/virtualNetworks.
func VirtualNetworkBgpCommunity() validator.String {
	return virtualNetworkBgpCommunity{}
}
