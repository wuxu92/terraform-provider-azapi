// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package loganalytics

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LogAnalyticsCluster provides resource knowledge for
// Microsoft.OperationalInsights/clusters.
//
// Contributing Terraform resources:
//   - azurerm_log_analytics_cluster
//   - azurerm_log_analytics_cluster_customer_managed_key (mutates the same cluster body's
//     properties.keyVaultProperties)
//
// Sources:
//   - terraform-provider-azurerm internal/services/loganalytics/log_analytics_cluster_resource.go
//     (Arguments L40-64, Create body L121-130, timeouts L89/L144/L196/L237)
//   - terraform-provider-azurerm internal/services/loganalytics/log_analytics_cluster_customer_managed_key_resource.go
//     (schema L48-61, body L112-121)
//   - terraform-provider-azurerm internal/services/loganalytics/validate/log_analytics_cluster_name.go,
//     internal.go (logAnalyticsGenericName)
//   - go-azure-sdk resource-manager/operationalinsights/2022-10-01/clusters:
//     model_clusterproperties.go, model_clustersku.go, model_keyvaultproperties.go,
//     constants.go (ClusterSkuNameEnum, Capacity, BillingType)
//
// Notes:
//   - name/location are envelope-owned and ForceNew; the name regex is emitted with
//     PropertyPath "" and location is listed as a body ForceNew path.
//   - identity is Required + ForceNew (SystemOrUserAssignedIdentityRequiredForceNew).
//   - sku.name is hardcoded to "CapacityReservation" by AzureRM (the only ARM enum value).
//   - size_gb maps to properties.sku.capacity, an int enum accepting only discrete values
//     (100..50000); expressed as a loose int range guard since azwise IntRule has no
//     discrete-set constraint.
//   - the customer_managed_key resource sets a composite Key Vault key URI that ARM splits
//     into properties.keyVaultProperties.{keyName,keyVaultUri,keyVersion}; that composite
//     has no single ARM body field, so it is non-mappable and no declarative rule is emitted.
type LogAnalyticsCluster struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LogAnalyticsCluster)(nil)

// NewLogAnalyticsCluster returns knowledge for the clusters resource.
func NewLogAnalyticsCluster() *LogAnalyticsCluster {
	return &LogAnalyticsCluster{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.OperationalInsights/clusters",
			ApiVersions:  []string{"2022-10-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 6 * time.Hour,
				Read:   5 * time.Minute,
				Update: 6 * time.Hour,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "identity"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[A-Za-z0-9][A-Za-z0-9-]+[A-Za-z0-9]$`,
					MinLength:    4,
					MaxLength:    63,
					Message:      "cluster name may contain only letters, numbers and hyphens (not leading/trailing), 4-63 characters",
				},
				{
					PropertyPath:  "properties.sku.name",
					AllowedValues: []string{"CapacityReservation"},
					Message:       "sku name must be CapacityReservation",
				},
				{
					PropertyPath:  "properties.billingType",
					AllowedValues: []string{"Cluster", "Workspaces"},
					Message:       "billing type must be Cluster or Workspaces",
				},
			},
			IntRules: []azwise.IntRule{
				{
					// Discrete enum (100,200,300,400,500,1000,2000,5000,10000,25000,50000);
					// expressed as a loose bound guard.
					PropertyPath: "properties.sku.capacity",
					MinValue:     azwise.Ptr(int64(100)),
					MaxValue:     azwise.Ptr(int64(50000)),
					Message:      "size_gb must be one of 100, 200, 300, 400, 500, 1000, 2000, 5000, 10000, 25000, 50000",
				},
			},
			RequiredFields: []string{
				// AzureRM hardcodes sku.name to CapacityReservation on create.
				"properties.sku.name",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.sku.name", Value: "CapacityReservation"},
				{PropertyPath: "properties.sku.capacity", Value: 100},
			},
		},
	}
}

func init() { azwise.Register(NewLogAnalyticsCluster()) }
