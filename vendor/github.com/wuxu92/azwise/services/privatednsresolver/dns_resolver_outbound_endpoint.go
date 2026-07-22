package privatednsresolver

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DnsResolverOutboundEndpoint provides resource knowledge for
// Microsoft.Network/dnsResolvers/outboundEndpoints.
//
// TF resource: azurerm_private_dns_resolver_outbound_endpoint
//
// Sources:
//   - internal/services/privatednsresolver/private_dns_resolver_outbound_endpoint_resource.go
//     (schema lines 47-74; Create/Update/Read/Delete timeouts lines 82/130/169/211)
//   - vendor/.../dnsresolver/2022-07-01/outboundendpoints/model_outboundendpointproperties.go
//   - vendor/.../outboundendpoints/id_outboundendpoint.go
//     (child of dnsResolvers; segment "outboundEndpoints")
//
// Notes:
//   - private_dns_resolver_id is the parent-resource reference (envelope), not a body field.
//   - subnet_id maps to properties.subnet.id (SubResource), Required + ForceNew.
type DnsResolverOutboundEndpoint struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DnsResolverOutboundEndpoint)(nil)

func NewDnsResolverOutboundEndpoint() *DnsResolverOutboundEndpoint {
	return &DnsResolverOutboundEndpoint{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/dnsResolvers/outboundEndpoints",
			ApiVersions:  []string{"2022-07-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.subnet.id"}, // subnet_id (ForceNew)
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.subnet.id",
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.resourceGuid",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDnsResolverOutboundEndpoint()) }
