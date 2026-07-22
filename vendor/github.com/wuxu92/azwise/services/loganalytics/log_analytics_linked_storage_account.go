// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package loganalytics

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LogAnalyticsLinkedStorageAccount provides resource knowledge for
// Microsoft.OperationalInsights/workspaces/linkedStorageAccounts.
//
// Contributing Terraform resource: azurerm_log_analytics_linked_storage_account.
//
// Sources:
//   - terraform-provider-azurerm internal/services/loganalytics/log_analytics_linked_storage_account_resource.go
//     (schema L49-101, body L137-141)
//   - go-azure-sdk resource-manager/operationalinsights/2020-08-01/linkedstorageaccounts:
//     model_linkedstorageaccountsproperties.go, constants.go (DataSourceType),
//     id_datasourcetype.go
//
// Notes:
//   - the resource name IS the dataSourceType (a ConstantSegment in the ARM ID), so the
//     DataSourceType enum is emitted with PropertyPath "" (the name attribute).
//   - workspace_id/resource_group_name are envelope/parent references.
//   - storage_account_ids maps to properties.storageAccountIds (Required, min 1 item).
type LogAnalyticsLinkedStorageAccount struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LogAnalyticsLinkedStorageAccount)(nil)

// NewLogAnalyticsLinkedStorageAccount returns knowledge for the
// workspaces/linkedStorageAccounts resource.
func NewLogAnalyticsLinkedStorageAccount() *LogAnalyticsLinkedStorageAccount {
	return &LogAnalyticsLinkedStorageAccount{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.OperationalInsights/workspaces/linkedStorageAccounts",
			ApiVersions:  []string{"2020-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "",
					AllowedValues: []string{"Alerts", "AzureWatson", "CustomLogs", "Ingestion", "Query"},
					Message:       "data source type must be one of Alerts, AzureWatson, CustomLogs, Ingestion, Query",
				},
			},
			RequiredFields: []string{
				"properties.storageAccountIds",
			},
		},
	}
}

func init() { azwise.Register(NewLogAnalyticsLinkedStorageAccount()) }
