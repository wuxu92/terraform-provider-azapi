package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkSecurityPerimeterProfile provides resource knowledge for
// Microsoft.Network/networkSecurityPerimeters/profiles.
//
// Mirrors azurerm_network_security_perimeter_profile. name is ForceNew;
// network_security_perimeter_id is the envelope parent reference. The create
// body has no settable properties beyond the name.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_security_perimeter_profile_resource.go
//     (typed schema, Create 30m / Read 5m timeouts)
//   - go-azure-sdk resource-manager/network/2025-01-01/networksecurityperimeterprofiles:
//     id_profile.go (segment casing "networkSecurityPerimeters/profiles")
type NetworkSecurityPerimeterProfile struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkSecurityPerimeterProfile)(nil)

// NewNetworkSecurityPerimeterProfile returns knowledge for the
// networkSecurityPerimeters/profiles resource.
func NewNetworkSecurityPerimeterProfile() *NetworkSecurityPerimeterProfile {
	return &NetworkSecurityPerimeterProfile{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkSecurityPerimeters/profiles",
			ApiVersions:  []string{"2025-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					Regex:   `(^[a-zA-Z0-9]+[a-zA-Z0-9_.-]{0,78}[a-zA-Z0-9_]+$)|(^[a-zA-Z0-9]$)`,
					Message: "name must be 1-80 chars, start with a letter or number, end with a letter, number or underscore, and contain only letters, numbers, underscores, periods, or hyphens",
				},
			},
		},
	}
}

func init() { azwise.Register(NewNetworkSecurityPerimeterProfile()) }
