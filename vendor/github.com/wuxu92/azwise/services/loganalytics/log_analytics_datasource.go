// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package loganalytics

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LogAnalyticsDataSource provides resource knowledge for
// Microsoft.OperationalInsights/workspaces/dataSources.
//
// Contributing Terraform resources (discriminated by the ARM body "kind"):
//   - azurerm_log_analytics_datasource_windows_event                 (kind WindowsEvent)
//   - azurerm_log_analytics_datasource_windows_performance_counter   (kind WindowsPerformanceCounter)
//
// Sources:
//   - terraform-provider-azurerm internal/services/loganalytics/log_analytics_datasource_windows_event_resource.go
//     (schema L50-85, body L122-128)
//   - terraform-provider-azurerm internal/services/loganalytics/log_analytics_datasource_windows_performance_counter_resource.go
//     (schema L50-91, body)
//   - go-azure-sdk resource-manager/operationalinsights/2020-08-01/datasources:
//     model_datasource.go, constants.go (DataSourceKind)
//
// Notes:
//   - name/workspace_name/resource_group_name are envelope/parent references. AzureRM
//     validates name with StringIsNotEmpty (no declarative constraint), so no name rule.
//   - the ARM "properties" payload is a free-form object keyed by "kind"; the per-kind
//     shapes (eventLogName/eventTypes[*] vs counterName/intervalSeconds/...) differ, so only
//     the universal "kind" enum is emitted. Kind-specific value constraints (event type
//     enum, interval range) are array-element / kind-specific and are intentionally omitted.
type LogAnalyticsDataSource struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LogAnalyticsDataSource)(nil)

// NewLogAnalyticsDataSource returns knowledge for the workspaces/dataSources resource.
func NewLogAnalyticsDataSource() *LogAnalyticsDataSource {
	return &LogAnalyticsDataSource{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.OperationalInsights/workspaces/dataSources",
			ApiVersions:  []string{"2020-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Full ARM SDK DataSourceKind enum set.
					PropertyPath: "kind",
					AllowedValues: []string{
						"ApplicationInsights", "AzureActivityLog", "AzureAuditLog",
						"ChangeTrackingContentLocation", "ChangeTrackingCustomPath",
						"ChangeTrackingDataTypeConfiguration", "ChangeTrackingDefaultRegistry",
						"ChangeTrackingLinuxPath", "ChangeTrackingPath", "ChangeTrackingRegistry",
						"ChangeTrackingServices", "CustomLog", "CustomLogCollection",
						"DnsAnalytics", "GenericDataSource", "IISLogs", "ImportComputerGroup",
						"Itsm", "LinuxChangeTrackingPath", "LinuxPerformanceCollection",
						"LinuxPerformanceObject", "LinuxSyslog", "LinuxSyslogCollection",
						"NetworkMonitoring", "Office365",
						"SecurityCenterSecurityWindowsBaselineConfiguration",
						"SecurityEventCollectionConfiguration",
						"SecurityInsightsSecurityEventCollectionConfiguration",
						"SecurityWindowsBaselineConfiguration", "SqlDataClassification",
						"WindowsEvent", "WindowsPerformanceCounter", "WindowsTelemetry",
					},
					Message: "kind must be a valid Log Analytics data source kind",
				},
			},
			RequiredFields: []string{
				"kind",
				"properties",
			},
		},
	}
}

func init() { azwise.Register(NewLogAnalyticsDataSource()) }
