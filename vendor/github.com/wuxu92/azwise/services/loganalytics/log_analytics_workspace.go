// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package loganalytics

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LogAnalyticsWorkspace provides resource knowledge for
// Microsoft.OperationalInsights/workspaces.
//
// Contributing Terraform resource: azurerm_log_analytics_workspace.
//
// Sources:
//   - terraform-provider-azurerm internal/services/loganalytics/log_analytics_workspace_resource.go
//     (schema L64-185, CustomizeDiff L208-248, Create body L290-395)
//   - terraform-provider-azurerm internal/services/loganalytics/validate/log_analytics_workspace_name.go
//   - go-azure-sdk resource-manager/operationalinsights/2023-09-01/workspaces:
//     model_workspaceproperties.go, model_workspacesku.go, model_workspacefeatures.go,
//     model_workspacecapping.go, constants.go (WorkspaceSkuNameEnum, PublicNetworkAccessType,
//     CapacityReservationLevel)
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; the name regex/length is
//     emitted with PropertyPath "".
//   - sku is only conditionally ForceNew (AzureRM's CustomizeDiff ignores LACluster
//     linkage and CapacityReservation<->PerGB2018 transitions). Documented and still
//     listed on properties.sku.name so unrelated tier changes are caught.
//   - internet_ingestion_enabled/internet_query_enabled are booleans mapped to the
//     PublicNetworkAccessType enum ("Enabled"/"Disabled") on the ARM body.
//   - local_authentication_enabled maps to the inverted properties.features.disableLocalAuth.
//   - allow_resource_only_permissions maps to
//     properties.features.enableLogAccessUsingOnlyResourcePermissions.
//   - reservation_capacity_in_gb_per_day maps to properties.sku.capacityReservationLevel,
//     an int enum accepting only discrete values (100..50000); expressed as a loose
//     int range guard since azwise IntRule has no discrete-set constraint.
//   - primary_shared_key/secondary_shared_key are read-only shared keys (separate
//     dataplane), not body properties.
//   - Log Analytics workspaces support soft-delete/recovery.
type LogAnalyticsWorkspace struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LogAnalyticsWorkspace)(nil)

// NewLogAnalyticsWorkspace returns knowledge for the workspaces resource.
func NewLogAnalyticsWorkspace() *LogAnalyticsWorkspace {
	return &LogAnalyticsWorkspace{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.OperationalInsights/workspaces",
			ApiVersions:  []string{"2023-09-01"},
			SoftDelete:   true,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				// Conditional ForceNew (see CustomizeDiff): unrelated tier changes force
				// replacement, but LACluster linkage and CapacityReservation<->PerGB2018
				// transitions do not.
				{PropertyPath: "properties.sku.name"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[A-Za-z0-9][A-Za-z0-9-]+[A-Za-z0-9]$`,
					MinLength:    4,
					MaxLength:    63,
					Message:      "workspace name may contain only letters, numbers and hyphens (not leading/trailing), 4-63 characters",
				},
				{
					PropertyPath: "properties.sku.name",
					// Full ARM SDK enum set (AzureRM restricts creation to a subset).
					AllowedValues: []string{
						"CapacityReservation", "Free", "LACluster", "PerGB2018",
						"PerNode", "Premium", "Standalone", "Standard",
					},
					Message: "sku name must be one of CapacityReservation, Free, LACluster, PerGB2018, PerNode, Premium, Standalone, Standard",
				},
				{
					PropertyPath:  "properties.publicNetworkAccessForIngestion",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.publicNetworkAccessForQuery",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be Enabled or Disabled",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.retentionInDays",
					MinValue:     azwise.Ptr(int64(30)),
					MaxValue:     azwise.Ptr(int64(730)),
					Message:      "retention_in_days must be between 30 and 730",
				},
				{
					// Discrete enum (100,200,300,400,500,1000,2000,5000,10000,25000,50000);
					// expressed as a loose bound guard.
					PropertyPath: "properties.sku.capacityReservationLevel",
					MinValue:     azwise.Ptr(int64(100)),
					MaxValue:     azwise.Ptr(int64(50000)),
					Message:      "capacity reservation level must be one of 100, 200, 300, 400, 500, 1000, 2000, 5000, 10000, 25000, 50000",
				},
			},
			FloatRules: []azwise.FloatRule{
				{
					PropertyPath: "properties.workspaceCapping.dailyQuotaGb",
					MinValue:     azwise.Ptr(-1.0),
					Message:      "daily_quota_gb must be -1 (unlimited) or greater",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.sku.name", Value: "PerGB2018"},
				{PropertyPath: "properties.publicNetworkAccessForIngestion", Value: "Enabled"},
				{PropertyPath: "properties.publicNetworkAccessForQuery", Value: "Enabled"},
				{PropertyPath: "properties.features.enableLogAccessUsingOnlyResourcePermissions", Value: true},
				{PropertyPath: "properties.features.disableLocalAuth", Value: false},
				{PropertyPath: "properties.workspaceCapping.dailyQuotaGb", Value: -1.0},
			},
		},
	}
}

func init() { azwise.Register(NewLogAnalyticsWorkspace()) }
