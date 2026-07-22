// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package loganalytics

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LogAnalyticsLinkedService provides resource knowledge for
// Microsoft.OperationalInsights/workspaces/linkedServices.
//
// Contributing Terraform resource: azurerm_log_analytics_linked_service.
//
// Sources:
//   - terraform-provider-azurerm internal/services/loganalytics/log_analytics_linked_service_resource.go
//     (schema L212-242, body L109-119)
//   - go-azure-sdk resource-manager/operationalinsights/2020-08-01/linkedservices:
//     model_linkedserviceproperties.go
//
// Notes:
//   - the resource name (linkedServiceName) is derived by AzureRM ("Automation" or
//     "Cluster") from which of read_access_id / write_access_id is set; it is not a
//     user-validated string, so no name rule is emitted.
//   - read_access_id maps to properties.resourceId (Automation Account),
//     write_access_id maps to properties.writeAccessResourceId (Log Analytics Cluster).
//     AzureRM's ExactlyOneOf(read_access_id, write_access_id) is a schema-level constraint
//     on distinct ARM paths and is expressed as an ExactlyOneOf relational rule.
type LogAnalyticsLinkedService struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LogAnalyticsLinkedService)(nil)

// NewLogAnalyticsLinkedService returns knowledge for the workspaces/linkedServices resource.
func NewLogAnalyticsLinkedService() *LogAnalyticsLinkedService {
	return &LogAnalyticsLinkedService{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.OperationalInsights/workspaces/linkedServices",
			ApiVersions:  []string{"2020-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ExactlyOneOf: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.resourceId", "properties.writeAccessResourceId"},
					Message: "exactly one of read_access_id (properties.resourceId) or write_access_id (properties.writeAccessResourceId) must be set",
				},
			},
		},
	}
}

func init() { azwise.Register(NewLogAnalyticsLinkedService()) }
