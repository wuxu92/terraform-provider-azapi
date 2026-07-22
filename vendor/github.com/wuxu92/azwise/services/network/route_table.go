package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// RouteTable provides resource knowledge for Microsoft.Network/routeTables.
//
// Mirrors azurerm_route_table. name and resource_group are envelope-owned;
// location forces replacement.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/route_table_resource.go
//     (schema, expandRouteTableRoutes, CRUD timeouts 30/5/30/30)
//   - terraform-provider-azurerm internal/services/network/validate/route_table_name.go (name regex)
//   - go-azure-sdk resource-manager/network/2025-01-01/routetables:
//     model_routetablepropertiesformat.go, id_routetable.go (segment casing "routeTables")
//
// Not encoded (deliberate):
//   - azurerm_route folds into this resource as the properties.routes[*] array
//     (see route_table_resource.go expandRouteTableRoutes). Its per-route enum
//     (next_hop_type VirtualNetworkGateway/VnetLocal/Internet/VirtualAppliance/None)
//     and name regex live under an array element, which azwise/azapin cannot lower
//     through "[*]". Skipped, not emitted.
//   - bgp_route_propagation_enabled inverts into
//     properties.disableBgpRoutePropagation (AzureRM Default true → ARM default
//     false); encoded as a DefaultValue below.
type RouteTable struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*RouteTable)(nil)

// NewRouteTable returns knowledge for the routeTables resource.
func NewRouteTable() *RouteTable {
	return &RouteTable{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/routeTables",
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
			StringRules: []azwise.StringRule{
				{
					Regex:   `^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,78}[a-zA-Z0-9_]?$`,
					Message: "name must be 1-80 chars, start with an alphanumeric, end with an alphanumeric or underscore, and contain only alphanumerics, underscores, periods, and hyphens",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// bgp_route_propagation_enabled defaults true; ARM stores the inverse.
				{PropertyPath: "properties.disableBgpRoutePropagation", Value: false},
			},
		},
	}
}

func init() { azwise.Register(NewRouteTable()) }
