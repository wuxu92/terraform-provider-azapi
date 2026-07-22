package privatednsresolver

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DnsResolverInboundEndpoint provides resource knowledge for
// Microsoft.Network/dnsResolvers/inboundEndpoints.
//
// TF resource: azurerm_private_dns_resolver_inbound_endpoint
//
// Sources:
//   - internal/services/privatednsresolver/private_dns_resolver_inbound_endpoint_resource.go
//     (schema lines 53-105; Create/Update/Read/Delete timeouts lines 113/166/205/248)
//   - vendor/.../dnsresolver/2022-07-01/inboundendpoints/model_inboundendpointproperties.go
//   - vendor/.../inboundendpoints/model_ipconfiguration.go
//   - vendor/.../inboundendpoints/id_inboundendpoint.go
//     (child of dnsResolvers; segment "inboundEndpoints")
//
// Notes:
//   - private_dns_resolver_id is the parent-resource reference (envelope), not a body field.
//   - ip_configurations (whole list) is Required + ForceNew -> properties.ipConfigurations.
//   - The per-element enum private_ip_allocation_method
//     (properties.ipConfigurations[*].privateIpAllocationMethod, AllowedValues
//     [Dynamic, Static], AzureRM default Dynamic) is an array-element path and is not
//     expressible as a declarative rule here; likewise the element privateIpAddress
//     (Optional+Computed) and subnet.id are array-element paths — skipped.
type DnsResolverInboundEndpoint struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DnsResolverInboundEndpoint)(nil)

func NewDnsResolverInboundEndpoint() *DnsResolverInboundEndpoint {
	return &DnsResolverInboundEndpoint{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/dnsResolvers/inboundEndpoints",
			ApiVersions:  []string{"2022-07-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.ipConfigurations"}, // ip_configurations (ForceNew)
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.ipConfigurations",
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.resourceGuid",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDnsResolverInboundEndpoint()) }
