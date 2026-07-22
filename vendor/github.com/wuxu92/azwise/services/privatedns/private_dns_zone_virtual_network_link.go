package privatedns

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PrivateDnsZoneVirtualNetworkLink provides resource knowledge for
// Microsoft.Network/privateDnsZones/virtualNetworkLinks (links a virtual network
// to a private DNS zone).
//
// Sources:
//   - internal/services/privatedns/private_dns_zone_virtual_network_link_resource.go (schema + CreateUpdate)
//   - vendor/.../privatedns/2024-06-01/virtualnetworklinks/model_virtualnetworklinkproperties.go
//   - vendor/.../privatedns/2024-06-01/virtualnetworklinks/constants.go (ResolutionPolicy)
//   - vendor/.../privatedns/2024-06-01/virtualnetworklinks/id_virtualnetworklink.go
//     (.../privateDnsZones/{zone}/virtualNetworkLinks/{name})
//
// Notes:
//   - `name` and `private_dns_zone_id` are URL/parent-reference segments; only the
//     link name is modeled here (parent zone id is envelope, excluded).
//   - `virtual_network_id` is Required + ForceNew → properties.virtualNetwork.id.
type PrivateDnsZoneVirtualNetworkLink struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PrivateDnsZoneVirtualNetworkLink)(nil)

func NewPrivateDnsZoneVirtualNetworkLink() *PrivateDnsZoneVirtualNetworkLink {
	return &PrivateDnsZoneVirtualNetworkLink{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/privateDnsZones/virtualNetworkLinks",
			ApiVersions:  []string{"2024-06-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},                          // link name (ForceNew, URL segment)
				{PropertyPath: "properties.virtualNetwork.id"},  // virtual_network_id (ForceNew)
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.resolutionPolicy",
					AllowedValues: []string{"Default", "NxDomainRedirect"}, // PossibleValuesForResolutionPolicy
					Message:       "resolution_policy must be Default or NxDomainRedirect",
				},
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.virtualNetworkLinkState",
			},
			RequiredFields: []string{
				"properties.virtualNetwork.id", // virtual_network_id (Required)
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.registrationEnabled", Value: false}, // registration_enabled default false
				{PropertyPath: "properties.resolutionPolicy", Value: nil},      // Optional+Computed; server decides
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewPrivateDnsZoneVirtualNetworkLink()) }
