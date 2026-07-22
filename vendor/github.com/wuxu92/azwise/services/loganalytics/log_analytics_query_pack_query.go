// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package loganalytics

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LogAnalyticsQueryPackQuery provides resource knowledge for
// Microsoft.OperationalInsights/queryPacks/queries.
//
// Contributing Terraform resource: azurerm_log_analytics_query_pack_query.
//
// Sources:
//   - terraform-provider-azurerm internal/services/loganalytics/log_analytics_query_pack_query_resource.go
//     (Arguments L53-333, Create body L375-409)
//   - go-azure-sdk resource-manager/operationalinsights/2019-09-01/querypackqueries:
//     model_loganalyticsquerypackqueryproperties.go,
//     model_loganalyticsquerypackquerypropertiesrelated.go
//
// Notes:
//   - name is Optional+Computed + ForceNew and validated as a UUID; emitted as a name-level
//     (PropertyPath "") UUID regex. When omitted AzureRM generates a UUID server-side.
//   - query_pack_id is a parent reference (envelope), not a body field.
//   - body/display_name are Required and map to properties.{body,displayName}.
//   - description -> properties.description, additional_settings_json -> properties.properties.
//   - categories/resource_types/solutions map to array fields under properties.related; their
//     per-element StringInSlice enums are array-element paths and are intentionally not emitted.
type LogAnalyticsQueryPackQuery struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LogAnalyticsQueryPackQuery)(nil)

// NewLogAnalyticsQueryPackQuery returns knowledge for the queryPacks/queries resource.
func NewLogAnalyticsQueryPackQuery() *LogAnalyticsQueryPackQuery {
	return &LogAnalyticsQueryPackQuery{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.OperationalInsights/queryPacks/queries",
			ApiVersions:  []string{"2019-09-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
					Message:      "query name must be a valid UUID",
				},
			},
			RequiredFields: []string{
				"properties.body",
				"properties.displayName",
			},
		},
	}
}

func init() { azwise.Register(NewLogAnalyticsQueryPackQuery()) }
