package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualHubRouteTable provides resource knowledge for
// Microsoft.Network/virtualHubs/hubRouteTables.
//
// Mirrors azurerm_virtual_hub_route_table. name, virtual_hub_id (parent) are
// envelope/parent references.
//
// azurerm_virtual_hub_route_table_route is NOT a separate ARM resource: it folds
// into this type's body (properties.routes[*]). It manages individual entries of
// the same hubRouteTables routes array, so it contributes no distinct ARM type and
// no separate knowledge file — its per-route rules (name, destinations,
// destinations_type CIDR/ResourceId/Service, next_hop, next_hop_type ResourceId)
// live under the routes[*] array element and cannot be lowered by azwise/azapin
// (array-element paths are unsupported), so they are documented here, not emitted.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/virtual_hub_route_table_resource.go
//     (schema 49-120, timeouts 36-41)
//   - internal/services/network/virtual_hub_route_table_route_resource.go (folded-in)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualwans:
//     id_hubroutetable.go:110 (segment casing), model_hubroutetable.go
type VirtualHubRouteTable struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualHubRouteTable)(nil)

// NewVirtualHubRouteTable returns knowledge for the hubRouteTables resource.
func NewVirtualHubRouteTable() *VirtualHubRouteTable {
	return &VirtualHubRouteTable{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/virtualHubs/hubRouteTables",
			ApiVersions:  []string{"2025-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewVirtualHubRouteTable()) }
