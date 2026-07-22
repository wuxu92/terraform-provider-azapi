package streamanalytics

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StreamAnalyticsCluster provides resource knowledge for
// Microsoft.StreamAnalytics/clusters.
//
// Mirrors azurerm_stream_analytics_cluster.
//
// Sources:
//   - terraform-provider-azurerm internal/services/streamanalytics/stream_analytics_cluster_resource.go:49-215
//   - Timeouts: Create 90m (:81) / Read 5m (:126) / Update 90m (:185) / Delete 90m (:166)
//   - go-azure-sdk resource-manager/streamanalytics/2020-03-01/clusters
//     model_cluster.go / model_clustersku.go (Sku is top-level, not under properties) / constants.go
//   - ARM type casing verified via clusters/id_cluster.go StaticSegment("staticClusters", "clusters").
//   - streaming_capacity uses validation.All(IntBetween(36,216), IntDivisibleBy(36)); the
//     divisible-by-36 constraint has no declarative azwise form and is noted here only.
type StreamAnalyticsCluster struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StreamAnalyticsCluster)(nil)

// NewStreamAnalyticsCluster returns knowledge for the Stream Analytics cluster resource.
func NewStreamAnalyticsCluster() *StreamAnalyticsCluster {
	return &StreamAnalyticsCluster{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StreamAnalytics/clusters",
			ApiVersions:  []string{"2020-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 90 * time.Minute,
				Delete: 90 * time.Minute,
			},
			// sku.name is hardcoded to "Default" by AzureRM; sku.capacity is Required.
			RequiredFields: []string{
				"sku.name",
				"sku.capacity",
			},
			DefaultValues: []azwise.DefaultValue{
				// sku.name is always "Default" (only allowed ClusterSkuName value).
				{PropertyPath: "sku.name", Value: "Default"},
			},
			IntRules: []azwise.IntRule{
				// streaming_capacity: IntBetween(36, 216) (also must be divisible by 36).
				{PropertyPath: "sku.capacity", MinValue: azwise.Ptr(int64(36)), MaxValue: azwise.Ptr(int64(216))},
			},
			StringRules: []azwise.StringRule{
				// name: validation.StringIsNotEmpty.
				{PropertyPath: "", MinLength: 1, Message: "cluster name must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewStreamAnalyticsCluster()) }
