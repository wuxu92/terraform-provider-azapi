package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VpnGatewayNatRule provides resource knowledge for
// Microsoft.Network/vpnGateways/natRules.
//
// Mirrors azurerm_vpn_gateway_nat_rule. name and vpn_gateway_id (parent) are
// envelope/parent references.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/vpn_gateway_nat_rule_resource.go
//     (schema 41-126, expand 156-173, timeouts 29-34)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualwans:
//     id_natrule.go:110 (segment casing), model_vpngatewaynatruleproperties.go,
//     constants.go (VpnNatRuleMode 2311-2313, VpnNatRuleType 2352-2354)
type VpnGatewayNatRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VpnGatewayNatRule)(nil)

// NewVpnGatewayNatRule returns knowledge for the natRules resource.
func NewVpnGatewayNatRule() *VpnGatewayNatRule {
	return &VpnGatewayNatRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/vpnGateways/natRules",
			ApiVersions:  []string{"2025-01-01"},
			// mode and type are ForceNew; mappings and ip_configuration_id are updatable.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.mode"},
				{PropertyPath: "properties.type"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.mode",
					AllowedValues: []string{"EgressSnat", "IngressSnat"},
					Message:       "mode must be EgressSnat or IngressSnat",
				},
				{
					PropertyPath:  "properties.type",
					AllowedValues: []string{"Static", "Dynamic"},
					Message:       "type must be Static or Dynamic",
				},
				{
					PropertyPath:  "properties.ipConfigurationId",
					AllowedValues: []string{"Instance0", "Instance1"},
					Message:       "ip_configuration_id must be Instance0 or Instance1",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.mode", Value: "EgressSnat"},
				{PropertyPath: "properties.type", Value: "Static"},
			},
		},
	}
}

func init() { azwise.Register(NewVpnGatewayNatRule()) }
