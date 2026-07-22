package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkProfile provides resource knowledge for Microsoft.Network/networkProfiles.
//
// Mirrors azurerm_network_profile. The container_network_interface block is
// Required and expands into properties.containerNetworkInterfaceConfigurations
// (an array — element-level rules are not encoded).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_profile_resource.go
//     (resourceNetworkProfile schema lines 32-107, expand lines 109-163)
//   - go-azure-sdk resource-manager/network/2025-01-01/networkprofiles:
//     id_networkprofile.go (ARM type segment "networkProfiles"),
//     model_networkprofilepropertiesformat.go (containerNetworkInterfaceConfigurations
//     json tag)
//
// Not encoded (deliberate):
//   - container_network_interface_ids is Computed read-only
//     (properties.containerNetworkInterfaces); already bicep ReadOnly.
//   - container_network_interface sub-fields (name, ip_configuration.subnet_id)
//     live under the containerNetworkInterfaceConfigurations[*] array element;
//     azwise cannot resolve a path through an array element, so they are skipped.
type NetworkProfile struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkProfile)(nil)

// NewNetworkProfile returns knowledge for the networkProfiles resource.
func NewNetworkProfile() *NetworkProfile {
	return &NetworkProfile{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkProfiles",
			ApiVersions:  []string{"2025-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.containerNetworkInterfaceConfigurations",
			},
		},
	}
}

func init() { azwise.Register(NewNetworkProfile()) }
