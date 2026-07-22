// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package sentinel

import (
	"time"

	"github.com/wuxu92/azwise"
)

// WatchlistItems provides resource knowledge for
// Microsoft.SecurityInsights/watchlists/watchlistItems (a child of a watchlist,
// itself an extension resource on a Microsoft.OperationalInsights/workspaces scope).
//
// Contributing Terraform resource: azurerm_sentinel_watchlist_item.
//
// Sources:
//   - terraform-provider-azurerm internal/services/sentinel/sentinel_watchlist_item_resource.go
//     (schema L30-53, timeouts L72-124)
//   - go-azure-sdk resource-manager/securityinsights/2022-11-01/watchlistitems:
//     id_watchlistitem.go (Microsoft.SecurityInsights/watchlists/watchlistItems),
//     model_watchlistitemproperties.go (itemsKeyValue)
//
// Notes:
//   - name is optional/computed (server-generated) and validated as a UUID.
//   - the user-supplied `properties` map maps to properties.itemsKeyValue.
type WatchlistItems struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*WatchlistItems)(nil)

// NewWatchlistItems returns knowledge for the watchlistItems resource type.
func NewWatchlistItems() *WatchlistItems {
	return &WatchlistItems{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SecurityInsights/watchlists/watchlistItems",
			ApiVersions:  []string{"2022-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{"properties.itemsKeyValue"},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
					Message:      "watchlist item name must be a valid UUID",
				},
			},
		},
	}
}

func init() { azwise.Register(NewWatchlistItems()) }
