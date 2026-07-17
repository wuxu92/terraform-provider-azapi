package validators

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// storageAccountIPRulePattern matches an IPv4 address with an optional CIDR
// prefix of 0-30 bits (e.g. "23.45.1.0" or "23.45.1.0/30"). It mirrors the
// pattern AzureRM uses for storage account network_acls ip_rules values.
var storageAccountIPRulePattern = regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}(/([0-9]|[1-2][0-9]|30))?$`)

type storageAccountIPRule struct{}

func (storageAccountIPRule) Description(context.Context) string {
	return "must be a public IPv4 address or CIDR range with a /0-30 prefix (e.g. 23.45.1.0/30); private ranges (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16) are not allowed"
}

func (v storageAccountIPRule) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (storageAccountIPRule) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	str := req.ConfigValue
	if str.IsUnknown() || str.IsNull() {
		return
	}
	value := str.ValueString()

	if !storageAccountIPRulePattern.MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid IP ACL rule",
			fmt.Sprintf("must be an IPv4 address with an optional slash and prefix length (0-30), e.g. 23.45.1.0/30, but got %q", value),
		)
		return
	}

	// Reject private (non-public) IPv4 ranges, matching AzureRM: 10.0.0.0/8,
	// 172.16.0.0/12 and 192.168.0.0/16. The CIDR suffix, if any, rides on the
	// last octet, so the first two octets are safe to inspect.
	parts := strings.Split(value, ".")
	first := parts[0]
	second, _ := strconv.Atoi(parts[1])
	if first == "10" || (first == "172" && second >= 16 && second <= 31) || (first == "192" && second == 168) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Private IP address not allowed",
			fmt.Sprintf("must be a public IP address but got %q", value),
		)
	}
}

// StorageAccountIPRule returns a string validator for storage account
// network_acls ip_rules values: a public IPv4 address or CIDR range (prefix
// 0-30), rejecting the private ranges. Ported from AzureRM's StorageAccountIpRule.
// Resource-specific to Microsoft.Storage/storageAccounts.
func StorageAccountIPRule() validator.String {
	return storageAccountIPRule{}
}
