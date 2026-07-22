package loadbalancer

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LoadBalancerInboundNatRule provides resource knowledge for
// Microsoft.Network/loadBalancers/inboundNatRules.
//
// This is a genuine ARM child type: the SDK exposes a dedicated id parser
// (id_inboundnatrule.go, ParseInboundNatRuleID) and a dedicated PUT endpoint
// (method_inboundnatrulescreateorupdate.go, InboundNatRulesCreateOrUpdate), so it is emitted as
// its own resource type rather than folded into the parent LB body.
//
// Contributing Terraform resource: azurerm_lb_nat_rule.
//
// Sources:
//   - terraform-provider-azurerm internal/services/loadbalancer/lb_nat_rule_resource.go
//     (schema L50-147, protocol L67-75, ports L77-136, ConflictsWith/RequiredWith L81-129,
//     idle_timeout IntBetween(4,30) Default 4 L131-136, timeouts L43-48)
//   - terraform-provider-azurerm helpers/validate/network.go (PortNumber 1-65535,
//     PortNumberOrZero 0-65535, L39-63)
//   - go-azure-sdk resource-manager/network/2023-09-01/loadbalancers:
//     model_inboundnatrulepropertiesformat.go, constants.go (TransportProtocol L1244-1250)
//
// Notes:
//   - name/resource_group_name are envelope-owned; loadbalancer_id is the parent reference
//     (all ForceNew in AzureRM, none a body property, so not emitted).
//   - frontend_ip_configuration_name is resolved to properties.frontendIPConfiguration.id (a
//     subresource id); it is Required so that id path is a required body field.
//   - frontend_ip_configuration_id (properties.frontendIPConfiguration.id) is echoed back as a
//     Computed field, but the value is user-supplied via the name, so it is a required body field
//     here, NOT a computed one. backend_ip_configuration_id is the genuinely read-only reference.
type LoadBalancerInboundNatRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LoadBalancerInboundNatRule)(nil)

// NewLoadBalancerInboundNatRule returns knowledge for the inboundNatRules child resource.
func NewLoadBalancerInboundNatRule() *LoadBalancerInboundNatRule {
	return &LoadBalancerInboundNatRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/loadBalancers/inboundNatRules",
			ApiVersions:  []string{"2023-09-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.protocol",
					AllowedValues: []string{"All", "Tcp", "Udp"},
					Message:       "protocol must be All, Tcp or Udp",
				},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.frontendPort", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(65535))},
				{PropertyPath: "properties.backendPort", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(65535))},
				{PropertyPath: "properties.frontendPortRangeStart", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(65535))},
				{PropertyPath: "properties.frontendPortRangeEnd", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(65535))},
				{PropertyPath: "properties.idleTimeoutInMinutes", MinValue: azwise.Ptr(int64(4)), MaxValue: azwise.Ptr(int64(30))},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.idleTimeoutInMinutes", Value: 4},
			},
			RequiredFields: []string{
				"properties.protocol",
				"properties.backendPort",
				"properties.frontendIPConfiguration.id",
			},
			// frontend_port ConflictsWith frontend_port_start/end + backend_address_pool_id;
			// backend_address_pool_id/frontend_port_start/frontend_port_end ConflictsWith frontend_port
			// (lb_nat_rule_resource.go L81, L111, L120, L128).
			ConflictsWith: []azwise.RelationalRule{
				{Paths: []string{"properties.frontendPort", "properties.frontendPortRangeStart", "properties.frontendPortRangeEnd", "properties.backendAddressPool.id"}},
				{Paths: []string{"properties.backendAddressPool.id", "properties.frontendPort"}},
				{Paths: []string{"properties.frontendPortRangeStart", "properties.frontendPort"}},
				{Paths: []string{"properties.frontendPortRangeEnd", "properties.frontendPort"}},
			},
			// backend_address_pool_id RequiredWith frontend_port_start + frontend_port_end, and each
			// range endpoint RequiredWith the pool id and the other endpoint
			// (lb_nat_rule_resource.go L112, L119, L127).
			RequiredWith: []azwise.RelationalRule{
				{Paths: []string{"properties.backendAddressPool.id", "properties.frontendPortRangeStart", "properties.frontendPortRangeEnd"}},
				{Paths: []string{"properties.frontendPortRangeStart", "properties.backendAddressPool.id", "properties.frontendPortRangeEnd"}},
				{Paths: []string{"properties.frontendPortRangeEnd", "properties.backendAddressPool.id", "properties.frontendPortRangeStart"}},
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.backendIPConfiguration.id",
			},
		},
	}
}

func init() { azwise.Register(NewLoadBalancerInboundNatRule()) }
