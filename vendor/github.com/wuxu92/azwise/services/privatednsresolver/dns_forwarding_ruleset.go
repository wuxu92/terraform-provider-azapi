package privatednsresolver

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DnsForwardingRuleset provides resource knowledge for
// Microsoft.Network/dnsForwardingRulesets.
//
// TF resource: azurerm_private_dns_resolver_dns_forwarding_ruleset
//
// This is a TOP-LEVEL ARM type (a sibling of dnsResolvers), NOT a child of
// dnsResolvers — confirmed via
// vendor/.../dnsresolver/2022-07-01/dnsforwardingrulesets/id_dnsforwardingruleset.go
// (Segments: providers/Microsoft.Network/dnsForwardingRulesets/{name}).
//
// Sources:
//   - internal/services/privatednsresolver/private_dns_resolver_dns_forwarding_ruleset_resource.go
//     (schema lines 45-69; Create/Update/Read/Delete timeouts lines 77/123/170/214)
//   - vendor/.../dnsforwardingrulesets/model_dnsforwardingrulesetproperties.go
//
// Notes:
//   - private_dns_resolver_outbound_endpoint_ids maps to
//     properties.dnsResolverOutboundEndpoints ([]SubResource), Required.
//   - name is ForceNew; location/resource_group_name are envelope fields.
type DnsForwardingRuleset struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DnsForwardingRuleset)(nil)

func NewDnsForwardingRuleset() *DnsForwardingRuleset {
	return &DnsForwardingRuleset{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/dnsForwardingRulesets",
			ApiVersions:  []string{"2022-07-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.dnsResolverOutboundEndpoints",
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.resourceGuid",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDnsForwardingRuleset()) }
