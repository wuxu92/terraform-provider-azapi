package streamanalytics

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StreamAnalyticsManagedPrivateEndpoint provides resource knowledge for
// Microsoft.StreamAnalytics/clusters/privateEndpoints.
//
// Mirrors azurerm_stream_analytics_managed_private_endpoint. Despite the "managed
// private endpoint" name, the ARM parent is a Stream Analytics *cluster*, not the
// job — verified via privateendpoints/id_privateendpoint.go
// (clusters/{clusterName}/privateEndpoints/{privateEndpointName}).
//
// Sources:
//   - terraform-provider-azurerm internal/services/streamanalytics/stream_analytics_managed_private_endpoint_resource.go:46-191
//   - Timeouts: Create 30m (:86) / Read 5m (:135) / Delete 5m (:176)
//   - go-azure-sdk resource-manager/streamanalytics/2020-03-01/privateendpoints
//     model_privateendpointproperties.go / model_privatelinkserviceconnection.go
//   - target_resource_id (-> privateLinkServiceId) and subresource_name (-> groupIds)
//     live inside the manualPrivateLinkServiceConnections[*] array element, so their
//     value constraints are array-element paths and are skipped (no declarative form).
type StreamAnalyticsManagedPrivateEndpoint struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StreamAnalyticsManagedPrivateEndpoint)(nil)

// NewStreamAnalyticsManagedPrivateEndpoint returns knowledge for the managed private endpoint resource.
func NewStreamAnalyticsManagedPrivateEndpoint() *StreamAnalyticsManagedPrivateEndpoint {
	return &StreamAnalyticsManagedPrivateEndpoint{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StreamAnalytics/clusters/privateEndpoints",
			ApiVersions:  []string{"2020-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 5 * time.Minute,
			},
			// The connection array must be present for a valid create. The per-element
			// privateLinkServiceId / groupIds are Required within each element (array
			// paths, not enforced declaratively here).
			RequiredFields: []string{
				"properties.manualPrivateLinkServiceConnections",
			},
			StringRules: []azwise.StringRule{
				// name: validation.StringIsNotEmpty.
				{PropertyPath: "", MinLength: 1, Message: "managed private endpoint name must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewStreamAnalyticsManagedPrivateEndpoint()) }
