package databricks

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualNetworkPeering provides resource knowledge for
// Microsoft.Databricks/workspaces/virtualNetworkPeerings.
//
// Contributing TF resource: azurerm_databricks_virtual_network_peering.
//
// Sources:
//   - AzureRM internal/services/databricks/databricks_virtual_network_peering_resource.go
//     (schema :50-122, Create props :167-186, Read :205-261)
//   - AzureRM internal/services/databricks/validate/databricks_virtual_network_peering_name.go
//   - go-azure-sdk .../databricks/2026-01-01/vnetpeering:
//     model_virtualnetworkpeeringpropertiesformat.go, model_addressspace.go,
//     model_virtualnetworkpeeringpropertiesformatremotevirtualnetwork.go (ARM body paths)
type VirtualNetworkPeering struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualNetworkPeering)(nil)

func NewVirtualNetworkPeering() *VirtualNetworkPeering {
	return &VirtualNetworkPeering{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Databricks/workspaces/virtualNetworkPeerings",
			ApiVersions:  []string{"2026-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.remoteAddressSpace.addressPrefixes"},
				{PropertyPath: "properties.remoteVirtualNetwork.id"},
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). validate.DatabricksVirtualNetworkPeeringName:
					// 2-80 chars, begin with letter/number, end with letter/number/underscore,
					// may contain letters, numbers, underscores, periods, or hyphens.
					Regex:     `^[a-zA-Z\d][a-zA-Z\d._-]{0,78}[a-zA-Z\d_]$`,
					MinLength: 2,
					MaxLength: 80,
					Message:   "must be 2-80 characters, begin with a letter or number, end with a letter, number or underscore, and contain only letters, numbers, underscores, periods, or hyphens",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.allowVirtualNetworkAccess", Value: true},
				{PropertyPath: "properties.allowForwardedTraffic", Value: false},
				{PropertyPath: "properties.allowGatewayTransit", Value: false},
				{PropertyPath: "properties.useRemoteGateways", Value: false},
			},
			RequiredFields: []string{
				"properties.remoteVirtualNetwork.id",
				"properties.remoteAddressSpace.addressPrefixes",
			},
			// peeringState/provisioningState are Azure-populated status fields.
			// databricksAddressSpace/databricksVirtualNetwork are RP-computed but are
			// present in the create body, so they are intentionally not stripped here.
			ComputedFields: []string{
				"properties.peeringState",
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewVirtualNetworkPeering()) }
