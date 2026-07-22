package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NatGateway provides resource knowledge for Microsoft.Network/natGateways.
//
// Mirrors azurerm_nat_gateway. sku_name expands into the top-level sku.name;
// idle_timeout_in_minutes into properties.idleTimeoutInMinutes.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/nat_gateway_resource.go
//     (resourceNatGatewaySchema lines 69-118, expand lines 143-157, timeouts)
//   - go-azure-sdk resource-manager/network/2025-01-01/natgateways:
//     id_natgateway.go (ARM type segment "natGateways"),
//     model_natgatewaypropertiesformat.go (idleTimeoutInMinutes json tag),
//     constants.go (NatGatewaySkuName: Standard, StandardV2)
//
// Not encoded (deliberate):
//   - azurerm_nat_gateway_public_ip_association and
//     azurerm_nat_gateway_public_ip_prefix_association manage the SAME natGateways
//     resource: they fold public IP / prefix references into
//     properties.publicIpAddresses / properties.publicIpPrefixes on the parent
//     body. They are not distinct ARM resource types, so no separate knowledge
//     files are emitted.
//   - zones is ForceNew but lives on the top-level "zones" envelope array
//     (commonschema), already RequiresReplace by construction.
type NatGateway struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NatGateway)(nil)

// NewNatGateway returns knowledge for the natGateways resource.
func NewNatGateway() *NatGateway {
	return &NatGateway{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/natGateways",
			ApiVersions:  []string{"2025-01-01"},
			// location and sku_name (sku.name) replace the gateway on change;
			// zones is also ForceNew (envelope array).
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "sku.name"},
				{PropertyPath: "zones"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Standard", "StandardV2"},
					Message:       "must be Standard or StandardV2",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.idleTimeoutInMinutes",
					MinValue:     azwise.Ptr(int64(4)),
					MaxValue:     azwise.Ptr(int64(120)),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "sku.name", Value: "Standard"},
				{PropertyPath: "properties.idleTimeoutInMinutes", Value: float64(4)},
			},
		},
	}
}

func init() { azwise.Register(NewNatGateway()) }
