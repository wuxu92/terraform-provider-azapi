// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package loganalytics

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LogAnalyticsDataExport provides resource knowledge for
// Microsoft.OperationalInsights/workspaces/dataExports.
//
// Contributing Terraform resource: azurerm_log_analytics_data_export_rule.
//
// Sources:
//   - terraform-provider-azurerm internal/services/loganalytics/log_analytics_data_export_rule_resource.go
//     (schema L55-99, Create body L138-168)
//   - terraform-provider-azurerm internal/services/loganalytics/validate/log_analytics_data_export_name.go,
//     internal.go (logAnalyticsGenericName)
//   - go-azure-sdk resource-manager/operationalinsights/2020-08-01/dataexport:
//     model_dataexportproperties.go, model_destination.go
//
// Notes:
//   - name is envelope-owned and ForceNew; the generic-name regex/length is emitted with
//     PropertyPath "".
//   - workspace_resource_id/resource_group_name are envelope/parent references, not body.
//   - destination_resource_id maps to properties.destination.resourceId; when the target is
//     an Event Hub, AzureRM additionally derives properties.destination.metaData.eventHubName.
type LogAnalyticsDataExport struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LogAnalyticsDataExport)(nil)

// NewLogAnalyticsDataExport returns knowledge for the workspaces/dataExports resource.
func NewLogAnalyticsDataExport() *LogAnalyticsDataExport {
	return &LogAnalyticsDataExport{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.OperationalInsights/workspaces/dataExports",
			ApiVersions:  []string{"2020-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[A-Za-z0-9][A-Za-z0-9-]+[A-Za-z0-9]$`,
					MinLength:    4,
					MaxLength:    63,
					Message:      "data export rule name may contain only letters, numbers and hyphens (not leading/trailing), 4-63 characters",
				},
			},
			RequiredFields: []string{
				"properties.destination.resourceId",
				"properties.tableNames",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.enable", Value: false},
			},
		},
	}
}

func init() { azwise.Register(NewLogAnalyticsDataExport()) }
