package costmanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CostManagementExport provides resource knowledge for
// Microsoft.CostManagement/exports.
//
// This is a MERGED file: AzureRM exposes one ARM type (exports) as three typed
// Terraform resources that differ ONLY by the parent scope the export is created
// under (billing account / resource group / subscription). The scope lives on the
// operational envelope (it is the ID scope, not a body property), so the body
// knowledge is identical across all three and is unioned here.
//
// Contributing Terraform resources:
//   - azurerm_billing_account_cost_management_export
//   - azurerm_resource_group_cost_management_export
//   - azurerm_subscription_cost_management_export
//
// Field -> ARM body mapping (from export_resource_base.go arguments + the per-scope
// Create funcs, all building exports.Export):
//   - active                         -> properties.schedule.status (Active/Inactive; bool default true -> Active)
//   - recurrence_type                -> properties.schedule.recurrence (RecurrenceType)
//   - recurrence_period_start_date   -> properties.schedule.recurrencePeriod.from
//   - recurrence_period_end_date     -> properties.schedule.recurrencePeriod.to
//   - file_format                    -> properties.format (FormatType; default Csv)
//   - export_data_options.type       -> properties.definition.type (ExportType)
//   - export_data_options.time_frame -> properties.definition.timeframe (TimeframeType)
//   - export_data_storage_location.container_id      -> one-to-many:
//       properties.deliveryInfo.destination.resourceId (parent storage account) +
//       properties.deliveryInfo.destination.container (container name)  [both ForceNew]
//   - export_data_storage_location.root_folder_path  -> properties.deliveryInfo.destination.rootFolderPath [ForceNew]
//
// Note: AzureRM's time_frame ValidateFunc adds "TheLast7Days", which is not yet in
// the SDK TimeframeType enum (SDK TODO referencing azure-rest-api-specs#23707) but is
// accepted by the API — it is included in AllowedValues so valid configs are not
// rejected.
//
// Sources:
//   - terraform-provider-azurerm internal/services/costmanagement/export_resource_base.go
//     (arguments 33-136 enums/defaults; deleteFunc 30m; expand helpers 162-244)
//   - internal/services/costmanagement/resource_group_cost_management_export_resource.go
//     (Create 74-142: schedule/status/format/definition/deliveryInfo mapping; 30m/5m timeouts)
//   - internal/services/costmanagement/{billing_account,subscription}_cost_management_export_resource.go
//     (identical body, scope-only differences)
//   - go-azure-sdk resource-manager/costmanagement/2023-08-01/exports
//     (model_export*.go, constants.go: RecurrenceType/FormatType/ExportType/TimeframeType/StatusType)
type CostManagementExport struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CostManagementExport)(nil)

// NewCostManagementExport returns knowledge for the exports resource.
func NewCostManagementExport() *CostManagementExport {
	return &CostManagementExport{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.CostManagement/exports",
			ApiVersions:  []string{"2023-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// The storage delivery destination (from container_id + root_folder_path)
			// is unconditionally ForceNew in AzureRM.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.deliveryInfo.destination.resourceId"},
				{PropertyPath: "properties.deliveryInfo.destination.container"},
				{PropertyPath: "properties.deliveryInfo.destination.rootFolderPath"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.schedule.recurrence",
					AllowedValues: []string{"Annually", "Daily", "Monthly", "Weekly"},
					Message:       "recurrence must be one of Annually, Daily, Monthly, Weekly",
				},
				{
					PropertyPath:  "properties.schedule.status",
					AllowedValues: []string{"Active", "Inactive"},
					Message:       "schedule status must be Active or Inactive",
				},
				{
					PropertyPath:  "properties.format",
					AllowedValues: []string{"Csv"},
					Message:       "format must be Csv",
				},
				{
					PropertyPath:  "properties.definition.type",
					AllowedValues: []string{"ActualCost", "AmortizedCost", "Usage"},
					Message:       "export type must be one of ActualCost, AmortizedCost, Usage",
				},
				{
					PropertyPath: "properties.definition.timeframe",
					// SDK TimeframeType set plus "TheLast7Days" (API-accepted, not yet in SDK).
					AllowedValues: []string{
						"BillingMonthToDate", "Custom", "MonthToDate",
						"TheLastBillingMonth", "TheLastMonth", "WeekToDate", "TheLast7Days",
					},
					Message: "timeframe must be a valid export timeframe",
				},
			},
			// active defaults true -> Active; file_format defaults Csv.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.schedule.status", Value: "Active"},
				{PropertyPath: "properties.format", Value: "Csv"},
			},
			// ARM body properties required for a valid export creation.
			RequiredFields: []string{
				"properties.schedule.recurrence",
				"properties.schedule.recurrencePeriod.from",
				"properties.definition.type",
				"properties.definition.timeframe",
				"properties.deliveryInfo.destination.resourceId",
				"properties.deliveryInfo.destination.container",
				"properties.deliveryInfo.destination.rootFolderPath",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewCostManagementExport()) }
