// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package loganalytics

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LogAnalyticsStorageInsights provides resource knowledge for
// Microsoft.OperationalInsights/workspaces/storageInsightConfigs.
//
// Contributing Terraform resource: azurerm_log_analytics_storage_insights.
//
// Sources:
//   - terraform-provider-azurerm internal/services/loganalytics/log_analytics_storage_insights_resource.go
//     (schema L176-227, body L86-98)
//   - terraform-provider-azurerm internal/services/loganalytics/validate/log_analytics_storage_insights_name.go,
//     internal.go (logAnalyticsGenericName)
//   - go-azure-sdk resource-manager/operationalinsights/2020-08-01/storageinsights:
//     model_storageinsightproperties.go, model_storageaccount.go
//
// Notes:
//   - name is envelope-owned and ForceNew; the generic-name regex/length is emitted with
//     PropertyPath "".
//   - workspace_id/resource_group_name are envelope/parent references.
//   - storage_account_id maps to properties.storageAccount.id;
//     storage_account_key maps to the sensitive properties.storageAccount.key.
//   - blob_container_names -> properties.containers, table_names -> properties.tables.
type LogAnalyticsStorageInsights struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LogAnalyticsStorageInsights)(nil)

// NewLogAnalyticsStorageInsights returns knowledge for the
// workspaces/storageInsightConfigs resource.
func NewLogAnalyticsStorageInsights() *LogAnalyticsStorageInsights {
	return &LogAnalyticsStorageInsights{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.OperationalInsights/workspaces/storageInsightConfigs",
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
					Message:      "storage insights name may contain only letters, numbers and hyphens (not leading/trailing), 4-63 characters",
				},
			},
			SensitiveFields: []string{
				"properties.storageAccount.key",
			},
			RequiredFields: []string{
				"properties.storageAccount.id",
				"properties.storageAccount.key",
			},
		},
	}
}

func init() { azwise.Register(NewLogAnalyticsStorageInsights()) }
