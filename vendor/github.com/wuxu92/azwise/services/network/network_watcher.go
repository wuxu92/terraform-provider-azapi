package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkWatcher provides resource knowledge for Microsoft.Network/networkWatchers.
//
// Mirrors azurerm_network_watcher. The body carries only Location + Tags; name
// and resource_group are envelope-owned.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_watcher_resource.go
//     (schema, CRUD timeouts 30/5/30/30)
//   - go-azure-sdk resource-manager/network/2025-01-01/networkwatchers:
//     id_networkwatcher.go (segment casing "networkWatchers")
type NetworkWatcher struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkWatcher)(nil)

// NewNetworkWatcher returns knowledge for the networkWatchers resource.
func NewNetworkWatcher() *NetworkWatcher {
	return &NetworkWatcher{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkWatchers",
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

func init() { azwise.Register(NewNetworkWatcher()) }
