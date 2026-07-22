package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkConnectionMonitor provides resource knowledge for
// Microsoft.Network/networkWatchers/connectionMonitors.
//
// Mirrors azurerm_network_connection_monitor. name, network_watcher_id and
// resource_group are envelope-owned; location is a top-level ARM property that
// forces replacement.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_connection_monitor_resource.go
//     (schema, expandNetworkConnectionMonitor*, CRUD timeouts 30/5/30/30)
//   - go-azure-sdk resource-manager/network/2025-01-01/connectionmonitors:
//     model_connectionmonitorparameters.go, id_connectionmonitor.go
//     (segment casing "networkWatchers/connectionMonitors")
//
// Not encoded (deliberate):
//   - endpoint, test_configuration and test_group are TypeSet blocks that expand
//     into properties.endpoints[*], properties.testConfigurations[*] and
//     properties.testGroups[*]. All their enums (endpoint coverage_level and
//     target_resource_type, test protocol Tcp/Http/Icmp, http method Get/Post,
//     preferred_ip_version IPv4/IPv6, destination_port_behavior None/ListenIfAvailable,
//     filter type/item type) live under array elements, which azwise/azapin cannot
//     lower through "[*]". Skipped, not emitted.
//   - properties.notes is the only free-form top-level body string; it has no
//     value constraint.
type NetworkConnectionMonitor struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkConnectionMonitor)(nil)

// NewNetworkConnectionMonitor returns knowledge for the networkWatchers/connectionMonitors resource.
func NewNetworkConnectionMonitor() *NetworkConnectionMonitor {
	return &NetworkConnectionMonitor{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkWatchers/connectionMonitors",
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
		},
	}
}

func init() { azwise.Register(NewNetworkConnectionMonitor()) }
