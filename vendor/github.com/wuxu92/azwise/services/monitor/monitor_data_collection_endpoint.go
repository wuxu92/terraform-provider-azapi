package monitor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DataCollectionEndpoint provides resource knowledge for
// Microsoft.Insights/dataCollectionEndpoints.
//
// Mirrors azurerm_monitor_data_collection_endpoint.
//
// Sources:
//   - terraform-provider-azurerm internal/services/monitor/monitor_data_collection_endpoint_resource.go
//     Arguments (43-77): name Required+ForceNew (StringIsNotEmpty); resource_group_name +
//     location ForceNew (commonschema); public_network_access_enabled Optional Default true →
//     properties.networkAcls.publicNetworkAccess; description Optional; kind Optional enum
//     (Linux/Windows). Attributes (79-101): configuration_access_endpoint, immutable_id,
//     logs_ingestion_endpoint, metrics_ingestion_endpoint all Computed. Create (115-160)
//     maps public_network_access bool → Enabled/Disabled. Timeouts Create/Update/Delete 30m,
//     Read 5m.
//   - go-azure-sdk resource-manager/insights/2023-03-11/datacollectionendpoints:
//     DataCollectionEndpoint (model_datacollectionendpoint.go) fields
//     configurationAccess/logsIngestion/metricsIngestion/immutableId/provisioningState
//     read-only; NetworkRuleSet.publicNetworkAccess enum
//     KnownPublicNetworkAccessOptions{Disabled,Enabled,SecuredByPerimeter};
//     KnownDataCollectionEndpointResourceKind{Linux,Windows};
//     id_datacollectionendpoint.go type Microsoft.Insights/dataCollectionEndpoints.
type DataCollectionEndpoint struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DataCollectionEndpoint)(nil)

// NewDataCollectionEndpoint returns knowledge for the dataCollectionEndpoints resource.
func NewDataCollectionEndpoint() *DataCollectionEndpoint {
	return &DataCollectionEndpoint{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Insights/dataCollectionEndpoints",
			ApiVersions:  []string{"2023-03-11"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name: StringIsNotEmpty.
					MinLength: 1,
					Message:   "name must not be empty",
				},
				{
					PropertyPath:  "kind",
					AllowedValues: []string{"Linux", "Windows"},
					Message:       "kind must be one of Linux, Windows",
				},
				{
					PropertyPath:  "properties.networkAcls.publicNetworkAccess",
					AllowedValues: []string{"Disabled", "Enabled", "SecuredByPerimeter"},
					Message:       "publicNetworkAccess must be one of Disabled, Enabled, SecuredByPerimeter",
				},
			},
			// public_network_access_enabled defaults true → networkAcls.publicNetworkAccess "Enabled".
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.networkAcls.publicNetworkAccess", Value: "Enabled"},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.configurationAccess",
				"properties.logsIngestion",
				"properties.metricsIngestion",
				"properties.immutableId",
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewDataCollectionEndpoint()) }
