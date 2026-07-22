package costmanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CostManagementView provides resource knowledge for
// Microsoft.CostManagement/views.
//
// This is a MERGED file: AzureRM exposes one ARM type (views) as two typed
// Terraform resources that differ ONLY by the parent scope (resource group /
// subscription). The scope is the ID scope, not a body property, so body knowledge
// is identical and is unioned here.
//
// Contributing Terraform resources:
//   - azurerm_resource_group_cost_management_view
//   - azurerm_subscription_cost_management_view
//
// Field -> ARM body mapping (from view_resource_base.go arguments + per-scope Create
// funcs building views.View):
//   - display_name        -> properties.displayName
//   - chart_type          -> properties.chart (ChartType)
//   - accumulated         -> properties.accumulated (AccumulatedType "true"/"false"; ForceNew)
//   - report_type         -> properties.query.type (ReportType; hardcoded Usage)
//   - timeframe           -> properties.query.timeframe (ReportTimeframeType)
//   - dataset.granularity -> properties.query.dataSet.granularity (ReportGranularityType)
//
// Not lowered to rules (array-element / map-key paths — unsupported by azwise):
//   - dataset.aggregation (name/column_name, both ForceNew) -> properties.query.dataSet.aggregation
//     is a map[string]ReportConfigAggregation; map-key rules cannot be expressed.
//   - dataset.sorting[*].direction -> properties.query.dataSet.sorting[*].direction (ReportConfigSortingType)
//   - dataset.grouping[*].type     -> properties.query.dataSet.grouping[*].type (QueryColumnType)
//   - kpi[*].type                  -> properties.kpis[*].type (KpiTypeType)
//   - pivot[*].type                -> properties.pivots[*].type (PivotTypeType)
//
// Sources:
//   - terraform-provider-azurerm internal/services/costmanagement/view_resource_base.go
//     (arguments 53-194 enums; deleteFunc 30m; expandDatasetFromModel 222-266)
//   - internal/services/costmanagement/resource_group_cost_management_view_resource.go
//     (Create 75-129: accumulated/chart/query mapping, report_type hardcoded Usage; 30m/5m timeouts)
//   - internal/services/costmanagement/subscription_cost_management_view_resource.go (identical body)
//   - go-azure-sdk resource-manager/costmanagement/2023-08-01/views
//     (model_view*.go, model_reportconfig*.go, constants.go)
type CostManagementView struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CostManagementView)(nil)

// NewCostManagementView returns knowledge for the views resource.
func NewCostManagementView() *CostManagementView {
	return &CostManagementView{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.CostManagement/views",
			ApiVersions:  []string{"2023-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// accumulated is unconditionally ForceNew in AzureRM.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.accumulated"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.chart",
					AllowedValues: []string{"Area", "GroupedColumn", "Line", "StackedColumn", "Table"},
					Message:       "chart type must be one of Area, GroupedColumn, Line, StackedColumn, Table",
				},
				{
					PropertyPath:  "properties.accumulated",
					AllowedValues: []string{"true", "false"},
					Message:       "accumulated must be \"true\" or \"false\"",
				},
				{
					PropertyPath:  "properties.query.type",
					AllowedValues: []string{"Usage"},
					Message:       "report type must be Usage",
				},
				{
					PropertyPath:  "properties.query.timeframe",
					AllowedValues: []string{"Custom", "MonthToDate", "WeekToDate", "YearToDate"},
					Message:       "timeframe must be one of Custom, MonthToDate, WeekToDate, YearToDate",
				},
				{
					PropertyPath:  "properties.query.dataSet.granularity",
					AllowedValues: []string{"Daily", "Monthly"},
					Message:       "dataset granularity must be Daily or Monthly",
				},
			},
			// ARM body properties required for a valid view creation.
			RequiredFields: []string{
				"properties.displayName",
				"properties.chart",
				"properties.accumulated",
				"properties.query.type",
				"properties.query.timeframe",
				"properties.query.dataSet.granularity",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewCostManagementView()) }
