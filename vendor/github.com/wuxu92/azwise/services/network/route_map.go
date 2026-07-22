package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// RouteMap provides resource knowledge for Microsoft.Network/virtualHubs/routeMaps.
//
// Mirrors azurerm_route_map. name is ForceNew; virtual_hub_id is the envelope
// parent reference (also ForceNew).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/route_map_resource.go
//     (typed schema, expandRules, Create/Update 60m, Read 5m, Delete 60m)
//   - terraform-provider-azurerm internal/services/network/validate/route_map_name.go (name regex)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualwans:
//     model_routemapproperties.go, id_routemap.go (segment casing "virtualHubs/routeMaps")
//
// Not encoded (deliberate):
//   - the rule block folds into properties.rules[*]; every enum
//     (next_step_if_matched Continue/Terminate/Unknown, action type
//     Add/Drop/Remove/Replace/Unknown, match_criteria match_condition
//     Contains/Equals/NotContains/NotEquals/Unknown) lives under array elements
//     which azwise/azapin cannot lower through "[*]". Skipped, not emitted.
//   - the CustomizeDiff requiring `parameter` when action type != Drop is a
//     value-conditional per-array-element rule with no declarative equivalent.
type RouteMap struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*RouteMap)(nil)

// NewRouteMap returns knowledge for the virtualHubs/routeMaps resource.
func NewRouteMap() *RouteMap {
	return &RouteMap{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/virtualHubs/routeMaps",
			ApiVersions:  []string{"2025-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					Regex:   `^[a-zA-Z0-9][a-zA-Z0-9_.-]+[a-zA-Z_0-9]$`,
					Message: "name must begin with a letter or number, end with a letter, number or underscore, and contain only letters, numbers, underscores, periods or hyphens",
				},
			},
		},
	}
}

func init() { azwise.Register(NewRouteMap()) }
