package networkfunction

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkFunctionCollectorPolicy provides resource knowledge for
// Microsoft.NetworkFunction/azureTrafficCollectors/collectorPolicies.
//
// Mirrors azurerm_network_function_collector_policy.
//
// Sources:
//   - terraform-provider-azurerm internal/services/networkfunction/network_function_collector_policy_resource.go
//     Arguments (59-124: name StringMatch regex + ForceNew; location ForceNew; traffic_collector_id
//     parent ref ForceNew; ipfx_emission MaxItems 1 Required ForceNew; ipfx_ingestion MaxItems 1
//     Required ForceNew; destination_types StringInSlice DestinationType; source_resource_ids
//     ValidateResourceID), Create (130-179: EmissionPolicies + IngestionPolicy with hardcoded
//     IngestionType IPFIX), timeouts Create 30m / Read 5m / Update 30m / Delete 30m.
//   - go-azure-sdk resource-manager/networkfunction/2022-11-01/collectorpolicies
//     CollectorPolicyPropertiesFormat (emissionPolicies/ingestionPolicy settable; provisioningState
//     read-only), IngestionPolicyPropertiesFormat (ingestionSources/ingestionType),
//     EmissionPoliciesPropertiesFormat (emissionDestinations/emissionType), constants.go
//     DestinationType [AzureMonitor] / IngestionType [IPFIX] / EmissionType [IPFIX],
//     id_collectorpolicy.go Segments (azureTrafficCollectors / collectorPolicies casing).
//
// Note: destination_types enum (DestinationType: AzureMonitor) lives at the array-element path
// properties.emissionPolicies[*].emissionDestinations[*].destinationType, and source_resource_ids
// at properties.ingestionPolicy.ingestionSources[*].resourceId; azwise cannot resolve array-element
// paths so these are documented but not emitted as rules.
type NetworkFunctionCollectorPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkFunctionCollectorPolicy)(nil)

// NewNetworkFunctionCollectorPolicy returns knowledge for the collectorPolicies resource.
func NewNetworkFunctionCollectorPolicy() *NetworkFunctionCollectorPolicy {
	return &NetworkFunctionCollectorPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.NetworkFunction/azureTrafficCollectors/collectorPolicies",
			ApiVersions:  []string{"2022-11-01"},
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
				{PropertyPath: "properties.emissionPolicies"},
				{PropertyPath: "properties.ingestionPolicy.ingestionSources"},
			},
			RequiredFields: []string{
				"properties.emissionPolicies",
				"properties.ingestionPolicy.ingestionSources",
			},
			StringRules: []azwise.StringRule{
				{
					// name StringMatch (resource.go:65-68).
					PropertyPath: "name",
					MaxLength:    80,
					Regex:        "^[a-zA-Z0-9]([-._a-zA-Z0-9]{0,78}[a-zA-Z0-9_])?$",
					Message:      "name may contain letters, numbers, periods, hyphens and underscores, up to 80 chars, begin with a letter or number and end with a letter, number or underscore",
				},
			},
			ArrayRules: []azwise.ArrayRule{
				{
					// ipfx_emission MaxItems 1 (resource.go:84).
					PropertyPath: "properties.emissionPolicies",
					MaxItems:     1,
					Message:      "only one emission policy is supported",
				},
			},
			// Server-populated, read-only ARM property.
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewNetworkFunctionCollectorPolicy()) }
