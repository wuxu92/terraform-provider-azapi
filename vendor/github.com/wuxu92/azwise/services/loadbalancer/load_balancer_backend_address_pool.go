package loadbalancer

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LoadBalancerBackendAddressPool provides resource knowledge for
// Microsoft.Network/loadBalancers/backendAddressPools.
//
// This is a genuine ARM child type: the SDK exposes a dedicated id parser
// (id_loadbalancerbackendaddresspool.go, ParseLoadBalancerBackendAddressPoolID) and a
// dedicated PUT endpoint (method_loadbalancerbackendaddresspoolscreateorupdate.go,
// LoadBalancerBackendAddressPoolsCreateOrUpdate), so it is emitted as its own resource type
// rather than folded into the parent LB body.
//
// Contributing Terraform resource: azurerm_lb_backend_address_pool.
//
// Sources:
//   - terraform-provider-azurerm internal/services/loadbalancer/lb_backend_address_pool_resource.go
//     (schema L50-155, synchronous_mode L65-71, tunnel_interface L73-116,
//     virtual_network_id L118-122, timeouts L43-48)
//   - go-azure-sdk resource-manager/network/2023-09-01/loadbalancers:
//     model_backendaddresspoolpropertiesformat.go, constants.go (SyncMode L1203-1208)
//
// Notes:
//   - name is envelope-owned; loadbalancer_id is the parent reference (both ForceNew in AzureRM,
//     neither is a body property, so not emitted).
//   - tunnel_interface maps to properties.tunnelInterfaces (array); its type/protocol/identifier/port
//     sub-fields are array-element paths and carry no declarative rules here.
//   - backend_ip_configurations, inbound_nat_rules, load_balancing_rules, outbound_rules are
//     read-only reference lists (Computed) populated by the service.
//   - lb_backend_address_pool_address is an element within a pool's
//     properties.loadBalancerBackendAddresses array (not a distinct ARM type); folded here, no rule.
type LoadBalancerBackendAddressPool struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LoadBalancerBackendAddressPool)(nil)

// NewLoadBalancerBackendAddressPool returns knowledge for the backendAddressPools child resource.
func NewLoadBalancerBackendAddressPool() *LoadBalancerBackendAddressPool {
	return &LoadBalancerBackendAddressPool{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/loadBalancers/backendAddressPools",
			ApiVersions:  []string{"2023-09-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.syncMode"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.syncMode",
					AllowedValues: []string{"Automatic", "Manual"},
					Message:       "synchronous_mode must be Automatic or Manual",
				},
			},
			// synchronous_mode RequiredWith virtual_network_id (lb_backend_address_pool_resource.go L70).
			RequiredWith: []azwise.RelationalRule{
				{Paths: []string{"properties.syncMode", "properties.virtualNetwork.id"}},
			},
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewLoadBalancerBackendAddressPool()) }
