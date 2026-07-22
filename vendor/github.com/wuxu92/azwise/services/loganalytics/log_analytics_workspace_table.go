// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package loganalytics

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LogAnalyticsWorkspaceTable provides resource knowledge for
// Microsoft.OperationalInsights/workspaces/tables.
//
// Contributing Terraform resources (all map to the one ARM tables type):
//   - azurerm_log_analytics_workspace_table             (built-in table plan/retention tuning)
//   - azurerm_log_analytics_workspace_table_custom_log  (custom "*_CL" DataCollectionRule tables)
//   - azurerm_log_analytics_workspace_table_microsoft   (built-in "*_CF" Microsoft tables)
//
// Sources:
//   - terraform-provider-azurerm internal/services/loganalytics/log_analytics_workspace_table_resource.go
//     (Arguments L52-87, Create body L126-143)
//   - terraform-provider-azurerm internal/services/loganalytics/log_analytics_workspace_table_custom_log_resource.go
//     (Arguments L62-118, Create body L179-201)
//   - terraform-provider-azurerm internal/services/loganalytics/log_analytics_workspace_table_microsoft_resource.go
//     (Arguments L69-168, Create body L237-293)
//   - go-azure-sdk resource-manager/operationalinsights/2022-10-01/tables:
//     model_tableproperties.go, model_schema.go, constants.go (TablePlanEnum, ColumnTypeEnum)
//
// Notes:
//   - only knowledge true for EVERY contributing table shape is unioned. The three siblings
//     have DIFFERENT name rules (custom_log requires "_CL" suffix + ForceNew; microsoft is a
//     fixed enum + ForceNew; workspace_table has no name validation and no ForceNew), so no
//     name rule (PropertyPath "") or name ForceNew is emitted — that would corrupt validation
//     for the other kinds.
//   - plan (properties.plan) and the retention ranges are universal: they only fire when the
//     field is present, so they are safe to union.
//   - column type enums live under properties.schema.columns[*].type (array-element paths)
//     and are intentionally not emitted.
type LogAnalyticsWorkspaceTable struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LogAnalyticsWorkspaceTable)(nil)

// NewLogAnalyticsWorkspaceTable returns knowledge for the workspaces/tables resource.
func NewLogAnalyticsWorkspaceTable() *LogAnalyticsWorkspaceTable {
	return &LogAnalyticsWorkspaceTable{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.OperationalInsights/workspaces/tables",
			ApiVersions:  []string{"2022-10-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.plan",
					AllowedValues: []string{"Analytics", "Basic"},
					Message:       "plan must be Analytics or Basic",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.retentionInDays",
					MinValue:     azwise.Ptr(int64(4)),
					MaxValue:     azwise.Ptr(int64(730)),
					Message:      "retention_in_days must be between 4 and 730",
				},
				{
					// AzureRM allows 4..730 plus the discrete set 1095,1460,1826,2191,2556,
					// 2922,3288,3653,4018,4383; expressed as a loose 4..4383 bound guard.
					PropertyPath: "properties.totalRetentionInDays",
					MinValue:     azwise.Ptr(int64(4)),
					MaxValue:     azwise.Ptr(int64(4383)),
					Message:      "total_retention_in_days must be between 4 and 730, or one of 1095, 1460, 1826, 2191, 2556, 2922, 3288, 3653, 4018, 4383",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.plan", Value: "Analytics"},
			},
		},
	}
}

func init() { azwise.Register(NewLogAnalyticsWorkspaceTable()) }
