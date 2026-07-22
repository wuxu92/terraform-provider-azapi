package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkManagerRoutingConfiguration provides resource knowledge for
// Microsoft.Network/networkManagers/routingConfigurations.
//
// Mirrors azurerm_network_manager_routing_configuration. name and the parent
// network_manager_id are envelope-owned (Required + RequiresReplace).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_manager_routing_configuration_resource.go
//   - go-azure-sdk resource-manager/network/2025-01-01/networkmanagerroutingconfigurations:
//     model_networkmanagerroutingconfigurationpropertiesformat.go, constants.go
type NetworkManagerRoutingConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkManagerRoutingConfiguration)(nil)

func NewNetworkManagerRoutingConfiguration() *NetworkManagerRoutingConfiguration {
	return &NetworkManagerRoutingConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkManagers/routingConfigurations",
			ApiVersions:  []string{"2025-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name validation (empty PropertyPath = the name attribute).
					Regex:     `^[a-zA-Z0-9\_\.\-]{1,64}$`,
					MinLength: 1,
					MaxLength: 64,
					Message:   "must be 1-64 characters and contain only letters, numbers, underscores, periods and hyphens",
				},
				{
					PropertyPath:  "properties.routeTableUsageMode",
					AllowedValues: []string{"ManagedOnly", "UseExisting"},
					Message:       "must be ManagedOnly or UseExisting",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.routeTableUsageMode", Value: "ManagedOnly"},
			},
		},
	}
}

func init() { azwise.Register(NewNetworkManagerRoutingConfiguration()) }
