package paloalto

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkVirtualAppliance provides resource knowledge for
// Microsoft.Network/networkVirtualAppliances.
//
// azurerm_palo_alto_virtual_network_appliance is the sole TF contributor to this ARM
// type; it is a thin wrapper that provisions an NVA delegated to
// "PaloAltoNetworks.Cloudngfw/firewalls" inside a Virtual Hub. Because
// networkVirtualAppliances is a general Network ARM type that a future generic
// resource could also target, only UNIVERSAL knowledge is emitted here: the name
// constraint and timeouts. The palo-alto-specific body (properties.delegation.serviceName,
// properties.virtualHub.id) is NOT captured as RequiredFields/DefaultValues, since those
// are not universal to every networkVirtualAppliance and would corrupt validation for
// any other contributor.
//
// Sources:
//   - terraform-provider-azurerm internal/services/paloalto/palo_alto_virtual_network_appliance_resource.go
//     Arguments() L43-59 (name ForceNew, network/validate.VirtualHubName ^.{1,256}$;
//     virtual_hub_id ForceNew, envelope/parent reference), Create() L65-128
//     (Delegation.ServiceName="PaloAltoNetworks.Cloudngfw/firewalls",
//     VirtualHub.Id — palo-alto-specific, not universal), timeouts 30m/5m/-/30m
//   - go-azure-sdk resource-manager/network/2025-01-01/networkvirtualappliances
//     id_networkvirtualappliance.go (Microsoft.Network/networkVirtualAppliances)
type NetworkVirtualAppliance struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkVirtualAppliance)(nil)

// NewNetworkVirtualAppliance returns knowledge for the networkVirtualAppliances resource.
func NewNetworkVirtualAppliance() *NetworkVirtualAppliance {
	return &NetworkVirtualAppliance{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkVirtualAppliances",
			ApiVersions:  []string{"2025-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "", // resource name
					MinLength:    1,
					MaxLength:    256,
					Message:      "name must be 1-256 characters",
				},
			},
		},
	}
}

func init() { azwise.Register(NewNetworkVirtualAppliance()) }
