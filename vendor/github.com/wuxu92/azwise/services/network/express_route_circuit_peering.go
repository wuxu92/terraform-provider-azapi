package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ExpressRouteCircuitPeering provides resource knowledge for
// Microsoft.Network/expressRouteCircuits/peerings.
//
// Mirrors azurerm_express_route_circuit_peering. resource_group_name,
// express_route_circuit_name and peering_type live on the operational envelope
// (peering_type is the resource name segment, mapped into properties.peeringType).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/express_route_circuit_peering_resource.go
//     (schema resourceExpressRouteCircuitPeering, Create body expand lines 283-341, CRUD timeouts)
//   - go-azure-sdk resource-manager/network/2025-01-01/expressroutecircuitpeerings:
//     id_expressroutecircuit.go (ARM type casing "expressRouteCircuits/peerings"),
//     model_expressroutecircuitpeeringpropertiesformat.go, constants.go (ExpressRoutePeeringType)
//
// Not encoded (deliberate):
//   - route_filter_id / ipv6.route_filter_id carry a resource-ID validator (semantic,
//     not declarative) — ported as azapin customizer validators, not here.
//   - customer_asn (Default 0) and routing_registry_name (Default "NONE") live inside
//     the optional microsoftPeeringConfig sub-object; filling them unconditionally
//     would inject the block, so they are not encoded as DefaultValues.
//   - azure_asn and gateway_manager_etag are Computed in AzureRM but ARE written into
//     the create body, so they are NOT read-only and are excluded from ComputedFields.
type ExpressRouteCircuitPeering struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ExpressRouteCircuitPeering)(nil)

// NewExpressRouteCircuitPeering returns knowledge for the peerings resource.
func NewExpressRouteCircuitPeering() *ExpressRouteCircuitPeering {
	return &ExpressRouteCircuitPeering{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/expressRouteCircuits/peerings",
			ApiVersions:  []string{"2025-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.peeringType",
					AllowedValues: []string{"AzurePrivatePeering", "AzurePublicPeering", "MicrosoftPeering"},
					Message:       "must be AzurePrivatePeering, AzurePublicPeering or MicrosoftPeering",
				},
				{
					PropertyPath: "properties.sharedKey",
					MinLength:    1,
					MaxLength:    25,
					Message:      "shared_key must be between 1 and 25 characters",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// ipv4_enabled Default:true -> properties.state "Enabled".
				{PropertyPath: "properties.state", Value: "Enabled"},
			},
			// vlan_id is Required in AzureRM.
			RequiredFields: []string{
				"properties.vlanId",
			},
			SensitiveFields: []string{
				"properties.sharedKey",
			},
			ComputedFields: []string{
				"properties.primaryAzurePort",
				"properties.secondaryAzurePort",
			},
		},
	}
}

func init() { azwise.Register(NewExpressRouteCircuitPeering()) }
