// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package loganalytics

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LogAnalyticsSolution provides resource knowledge for
// Microsoft.OperationsManagement/solutions.
//
// Contributing Terraform resource: azurerm_log_analytics_solution.
//
// Sources:
//   - terraform-provider-azurerm internal/services/loganalytics/log_analytics_solution_resource.go
//     (Arguments L74-132, Create body L172-185, plan expand L317-336)
//   - go-azure-sdk resource-manager/operationsmanagement/2015-11-01-preview/solution:
//     model_solutionproperties.go, model_solutionplan.go, id_solution.go
//
// Notes:
//   - solution_name/workspace_name/location/resource_group_name are envelope/composite name
//     inputs; AzureRM composes the ARM name as "SolutionName(WorkspaceName)".
//   - workspace_resource_id is Required + ForceNew and maps to
//     properties.workspaceResourceId.
//   - plan is Required (MaxItems 1): plan.publisher and plan.product are Required + ForceNew;
//     plan.name is server-computed (AzureRM overwrites it with the solution name), and
//     plan.promotion_code is Optional + ForceNew.
type LogAnalyticsSolution struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LogAnalyticsSolution)(nil)

// NewLogAnalyticsSolution returns knowledge for the solutions resource.
func NewLogAnalyticsSolution() *LogAnalyticsSolution {
	return &LogAnalyticsSolution{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.OperationsManagement/solutions",
			ApiVersions:  []string{"2015-11-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.workspaceResourceId"},
				{PropertyPath: "plan.publisher"},
				{PropertyPath: "plan.product"},
				{PropertyPath: "plan.promotionCode"},
			},
			RequiredFields: []string{
				"properties.workspaceResourceId",
				"plan.publisher",
				"plan.product",
			},
		},
	}
}

func init() { azwise.Register(NewLogAnalyticsSolution()) }
