package loadbalancer

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LoadBalancer provides resource knowledge for Microsoft.Network/loadBalancers.
//
// Contributing Terraform resource: azurerm_lb.
//
// Sources:
//   - terraform-provider-azurerm internal/services/loadbalancer/lb_resource.go
//     (schema L55-215, sku/sku_tier L68-89, edge_zone L66, frontend CustomizeDiff
//     ForceNewIf L217-245, timeouts L48-53)
//   - go-azure-sdk resource-manager/network/2023-09-01/loadbalancers:
//     model_loadbalancer.go, model_loadbalancerpropertiesformat.go,
//     model_loadbalancersku.go, constants.go
//     (LoadBalancerSkuName L393-399, LoadBalancerSkuTier L437-442)
//
// Notes:
//   - name/location/resource_group_name/tags are envelope-owned; not emitted as body rules.
//   - edge_zone maps to the top-level extendedLocation.name (EdgeZoneOptionalForceNew) and
//     is unconditionally ForceNew.
//   - frontend_ip_configuration is properties.frontendIPConfigurations (array). AzureRM
//     applies a *conditional* ForceNew (ForceNewIf, lb_resource.go L217-245): replacement is
//     forced only when the last frontend is removed (>0 -> 0 configs, Azure rejects an LB with
//     no frontend) or when a frontend config's `zones` change. That predicate is not a simple
//     old!=new on a single path, so it is intentionally NOT emitted as a declarative ForceNew
//     rule; its array-element sub-fields (subnet_id, private_ip_address, etc.) are array-element
//     paths and carry no declarative rules here.
//   - private_ip_address / private_ip_addresses are computed and derived from the frontend
//     configs; no direct settable ARM path.
//   - child properties (backendAddressPools, inboundNatRules) are managed as distinct ARM child
//     resource types (Microsoft.Network/loadBalancers/backendAddressPools and .../inboundNatRules)
//     with their own knowledge files; loadBalancingRules, probes, outboundRules and inboundNatPools
//     are array-element-only bodies of this LB with no dedicated ARM child type (see skips note).
type LoadBalancer struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LoadBalancer)(nil)

// NewLoadBalancer returns knowledge for the loadBalancers resource.
func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/loadBalancers",
			ApiVersions:  []string{"2023-09-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "sku.name"},
				{PropertyPath: "sku.tier"},
				{PropertyPath: "extendedLocation.name"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Basic", "Gateway", "Standard"},
					Message:       "sku must be Basic, Gateway or Standard",
				},
				{
					PropertyPath:  "sku.tier",
					AllowedValues: []string{"Global", "Regional"},
					Message:       "sku_tier must be Global or Regional",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "sku.name", Value: "Standard"},
				{PropertyPath: "sku.tier", Value: "Regional"},
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.resourceGuid",
			},
		},
	}
}

func init() { azwise.Register(NewLoadBalancer()) }
