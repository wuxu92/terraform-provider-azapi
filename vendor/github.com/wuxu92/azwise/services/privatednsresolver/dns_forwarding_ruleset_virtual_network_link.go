package privatednsresolver

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DnsForwardingRulesetVirtualNetworkLink provides resource knowledge for
// Microsoft.Network/dnsForwardingRulesets/virtualNetworkLinks.
//
// TF resource: azurerm_private_dns_resolver_virtual_network_link
//
// This is a child of dnsForwardingRulesets (NOT of dnsResolvers) — confirmed via
// vendor/.../dnsresolver/2022-07-01/virtualnetworklinks/id_virtualnetworklink.go
// (Segments: .../dnsForwardingRulesets/{ruleset}/virtualNetworkLinks/{name}).
//
// Sources:
//   - internal/services/privatednsresolver/private_dns_resolver_virtual_network_link_resource.go
//     (schema lines 43-74; Create/Update/Read/Delete timeouts lines 82/129/174/216)
//   - vendor/.../virtualnetworklinks/model_virtualnetworklinkproperties.go
//
// Notes:
//   - dns_forwarding_ruleset_id is the parent-resource reference (envelope), not a body field.
//   - virtual_network_id maps to properties.virtualNetwork.id (SubResource), Required + ForceNew.
//   - metadata maps to properties.metadata (map, no key rules emitted).
type DnsForwardingRulesetVirtualNetworkLink struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DnsForwardingRulesetVirtualNetworkLink)(nil)

func NewDnsForwardingRulesetVirtualNetworkLink() *DnsForwardingRulesetVirtualNetworkLink {
	return &DnsForwardingRulesetVirtualNetworkLink{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/dnsForwardingRulesets/virtualNetworkLinks",
			ApiVersions:  []string{"2022-07-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.virtualNetwork.id"}, // virtual_network_id (ForceNew)
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.virtualNetwork.id",
			},
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDnsForwardingRulesetVirtualNetworkLink()) }
