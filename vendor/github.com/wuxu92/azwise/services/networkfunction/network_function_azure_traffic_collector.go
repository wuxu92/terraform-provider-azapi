package networkfunction

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkFunctionAzureTrafficCollector provides resource knowledge for
// Microsoft.NetworkFunction/azureTrafficCollectors.
//
// Mirrors azurerm_network_function_azure_traffic_collector.
//
// Sources:
//   - terraform-provider-azurerm internal/services/networkfunction/network_function_azure_traffic_collector_resource.go
//     Arguments (46-64: name StringMatch regex + ForceNew; location/resource_group_name envelope),
//     Attributes (66-84: collector_policy_ids/virtual_hub_id Computed), Create (86-122: empty
//     PropertiesFormat), timeouts Create 30m / Read 5m / Update 30m / Delete 30m.
//   - go-azure-sdk resource-manager/networkfunction/2022-11-01/azuretrafficcollectors
//     AzureTrafficCollectorPropertiesFormat (collectorPolicies/provisioningState/virtualHub all
//     read-only), id_azuretrafficcollector.go Segments (Microsoft.NetworkFunction /
//     azureTrafficCollectors casing).
type NetworkFunctionAzureTrafficCollector struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkFunctionAzureTrafficCollector)(nil)

// NewNetworkFunctionAzureTrafficCollector returns knowledge for the azureTrafficCollectors resource.
func NewNetworkFunctionAzureTrafficCollector() *NetworkFunctionAzureTrafficCollector {
	return &NetworkFunctionAzureTrafficCollector{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.NetworkFunction/azureTrafficCollectors",
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
			},
			StringRules: []azwise.StringRule{
				{
					// name StringMatch (resource_resource.go:52-55).
					PropertyPath: "name",
					MaxLength:    80,
					Regex:        "^[a-zA-Z0-9]([-._a-zA-Z0-9]{0,78}[a-zA-Z0-9_])?$",
					Message:      "name may contain letters, numbers, periods, hyphens and underscores, up to 80 chars, begin with a letter or number and end with a letter, number or underscore",
				},
			},
			// Server-populated, read-only ARM properties (Create sends empty PropertiesFormat).
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.collectorPolicies",
				"properties.virtualHub",
			},
		},
	}
}

func init() { azwise.Register(NewNetworkFunctionAzureTrafficCollector()) }
