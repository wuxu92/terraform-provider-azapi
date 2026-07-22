package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkSecurityPerimeter provides resource knowledge for
// Microsoft.Network/networkSecurityPerimeters.
//
// Mirrors azurerm_network_security_perimeter. The body carries only Location +
// Tags; name and resource_group are envelope-owned. name is ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_security_perimeter_resource.go
//     (typed schema, Create/Update/Delete 30m, Read 5m)
//   - go-azure-sdk resource-manager/network/2025-01-01/networksecurityperimeters:
//     id_networksecurityperimeter.go (segment casing "networkSecurityPerimeters")
type NetworkSecurityPerimeter struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkSecurityPerimeter)(nil)

// NewNetworkSecurityPerimeter returns knowledge for the networkSecurityPerimeters resource.
func NewNetworkSecurityPerimeter() *NetworkSecurityPerimeter {
	return &NetworkSecurityPerimeter{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkSecurityPerimeters",
			ApiVersions:  []string{"2025-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// resource name attribute (PropertyPath == "")
					Regex:   `(^[a-zA-Z0-9]+[a-zA-Z0-9_.-]{0,78}[a-zA-Z0-9_]+$)|(^[a-zA-Z0-9]$)`,
					Message: "name must be 1-80 chars, start with a letter or number, end with a letter, number or underscore, and contain only letters, numbers, underscores, periods, or hyphens",
				},
			},
		},
	}
}

func init() { azwise.Register(NewNetworkSecurityPerimeter()) }
