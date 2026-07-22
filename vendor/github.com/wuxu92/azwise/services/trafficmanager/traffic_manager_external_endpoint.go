package trafficmanager

import (
	"time"

	"github.com/wuxu92/azwise"
)

// TrafficManagerExternalEndpoint provides resource knowledge for
// Microsoft.Network/trafficManagerProfiles/ExternalEndpoints.
//
// Mirrors azurerm_traffic_manager_external_endpoint. The endpoint type is a
// ConstantSegment in the SDK ID parser, so azure/external/nested endpoints are
// three DISTINCT ARM child types (verified via id_endpointtype.go Segments()).
// name and profile_id (parent) live on the envelope.
//
// Sources:
//   - terraform-provider-azurerm internal/services/trafficmanager/traffic_manager_external_endpoint_resource.go
//     (schema lines 50-155, Create body lines 184-224, CRUD timeouts 43-48)
//   - go-azure-sdk resource-manager/trafficmanager/2022-04-01/trafficmanagers:
//     id_endpointtype.go (ConstantSegment "ExternalEndpoints" under
//     "Microsoft.Network/trafficManagerProfiles"), model_endpointproperties.go,
//     constants.go (EndpointType, EndpointStatus)
//
// Not encoded (deliberate):
//   - subnet.first/last (IPv4Address) and subnet.scope (IntBetween 0-32) are
//     array-element paths (properties.subnets[*].*) which azwise cannot resolve.
type TrafficManagerExternalEndpoint struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*TrafficManagerExternalEndpoint)(nil)

// NewTrafficManagerExternalEndpoint returns knowledge for the ExternalEndpoints resource.
func NewTrafficManagerExternalEndpoint() *TrafficManagerExternalEndpoint {
	return &TrafficManagerExternalEndpoint{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/trafficManagerProfiles/ExternalEndpoints",
			ApiVersions:  []string{"2022-04-01"},
			// subnet is ForceNew (schema line 133) -> properties.subnets; name and
			// profile_id are envelope/parent refs.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.subnets"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.weight",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(1000)),
					Message:      "weight must be between 1 and 1000",
				},
				{
					PropertyPath: "properties.priority",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(1000)),
					Message:      "priority must be between 1 and 1000",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// enabled Default:true -> endpointStatus "Enabled".
				{PropertyPath: "properties.endpointStatus", Value: "Enabled"},
				// weight Default:1.
				{PropertyPath: "properties.weight", Value: int64(1)},
				// priority is Optional+Computed (server increments per endpoint count).
				{PropertyPath: "properties.priority"},
				// endpoint_location is Optional+Computed.
				{PropertyPath: "properties.endpointLocation"},
			},
			// target is Required (schema lines 65-68).
			RequiredFields: []string{
				"properties.target",
			},
			// endpointMonitorStatus is populated by Azure (read-only).
			ComputedFields: []string{
				"properties.endpointMonitorStatus",
			},
		},
	}
}

func init() { azwise.Register(NewTrafficManagerExternalEndpoint()) }
