package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// RouteFilter provides resource knowledge for Microsoft.Network/routeFilters.
//
// Mirrors azurerm_route_filter. name and resource_group are envelope-owned;
// location forces replacement.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/route_filter_resource.go
//     (schema, expandRouteFilterRules, CRUD timeouts 30/5/30/30)
//   - go-azure-sdk resource-manager/network/2025-01-01/routefilters:
//     model_routefilterpropertiesformat.go, id_routefilter.go (segment casing "routeFilters")
//
// Not encoded (deliberate):
//   - the rule block folds into properties.rules[*]; its enums (access = Allow,
//     rule_type = Community) live under an array element which azwise/azapin cannot
//     lower through "[*]". Skipped, not emitted.
type RouteFilter struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*RouteFilter)(nil)

// NewRouteFilter returns knowledge for the routeFilters resource.
func NewRouteFilter() *RouteFilter {
	return &RouteFilter{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/routeFilters",
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

func init() { azwise.Register(NewRouteFilter()) }
