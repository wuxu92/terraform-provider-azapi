// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package loganalytics

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LogAnalyticsSavedSearch provides resource knowledge for
// Microsoft.OperationalInsights/workspaces/savedSearches.
//
// Contributing Terraform resource: azurerm_log_analytics_saved_search.
//
// Sources:
//   - terraform-provider-azurerm internal/services/loganalytics/log_analytics_saved_search_resource.go
//     (schema L46-104, body L131-150)
//   - go-azure-sdk resource-manager/operationalinsights/2020-08-01/savedsearches:
//     model_savedsearchproperties.go
//
// Notes:
//   - name is envelope-owned and ForceNew (no declarative validation in AzureRM).
//   - log_analytics_workspace_id/resource_group_name are envelope/parent references.
//   - every content field is ForceNew in AzureRM: category, display_name, query,
//     function_alias, function_parameters and tags. category/display_name/query are
//     Required and map to properties.{category,displayName,query}.
//   - function_parameters is a list joined by AzureRM into a single
//     properties.functionParameters string; the per-item StringMatch regex applies to the
//     joined representation and cannot map to a single ARM field, so it is not emitted.
type LogAnalyticsSavedSearch struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LogAnalyticsSavedSearch)(nil)

// NewLogAnalyticsSavedSearch returns knowledge for the workspaces/savedSearches resource.
func NewLogAnalyticsSavedSearch() *LogAnalyticsSavedSearch {
	return &LogAnalyticsSavedSearch{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.OperationalInsights/workspaces/savedSearches",
			ApiVersions:  []string{"2020-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.category"},
				{PropertyPath: "properties.displayName"},
				{PropertyPath: "properties.query"},
				{PropertyPath: "properties.functionAlias"},
				{PropertyPath: "properties.functionParameters"},
			},
			RequiredFields: []string{
				"properties.category",
				"properties.displayName",
				"properties.query",
			},
		},
	}
}

func init() { azwise.Register(NewLogAnalyticsSavedSearch()) }
