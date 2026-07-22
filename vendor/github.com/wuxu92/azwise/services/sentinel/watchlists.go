// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package sentinel

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Watchlists provides resource knowledge for
// Microsoft.SecurityInsights/watchlists (an extension resource on a
// Microsoft.OperationalInsights/workspaces scope).
//
// Contributing Terraform resource: azurerm_sentinel_watchlist.
//
// Sources:
//   - terraform-provider-azurerm internal/services/sentinel/sentinel_watchlist_resource.go
//     (schema L35-83, body L131-149)
//   - go-azure-sdk resource-manager/securityinsights/2022-11-01/watchlists:
//     id_watchlist.go (Microsoft.SecurityInsights/watchlists),
//     model_watchlistproperties.go (displayName, itemsSearchKey, provider, defaultDuration)
//
// Notes:
//   - Every settable field is ForceNew — the watchlist is replace-only in AzureRM.
//   - properties.provider is hardcoded to "Microsoft" by AzureRM.
type Watchlists struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Watchlists)(nil)

// NewWatchlists returns knowledge for the watchlists resource type.
func NewWatchlists() *Watchlists {
	return &Watchlists{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SecurityInsights/watchlists",
			ApiVersions:  []string{"2022-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.displayName"},
				{PropertyPath: "properties.itemsSearchKey"},
				{PropertyPath: "properties.description"},
				{PropertyPath: "properties.labels"},
				{PropertyPath: "properties.defaultDuration"},
			},
			RequiredFields: []string{
				"properties.displayName",
				"properties.itemsSearchKey",
				"properties.provider",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					MinLength:    1,
					Message:      "watchlist name must not be empty",
				},
				{
					PropertyPath: "properties.displayName",
					MinLength:    1,
					Message:      "display_name must not be empty",
				},
				{
					PropertyPath: "properties.itemsSearchKey",
					MinLength:    1,
					Message:      "item_search_key must not be empty",
				},
			},
		},
	}
}

func init() { azwise.Register(NewWatchlists()) }
