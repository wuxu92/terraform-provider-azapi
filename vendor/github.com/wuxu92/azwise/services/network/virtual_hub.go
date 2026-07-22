package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualHub provides resource knowledge for Microsoft.Network/virtualHubs.
//
// This ARM type is the shared body for two AzureRM TF resources:
//   - azurerm_virtual_hub       (primary; full property surface)
//   - azurerm_route_server      (a Standard-sku virtualHub; contributed by NetSecurity)
//
// Only knowledge universal to EVERY virtualHubs body is encoded. Per-kind rules
// (route_server's restrictive 1-80 name regex, hardcoded "Standard" sku, its
// ForceNew on ipConfigurations/subnet) are intentionally NOT unioned — they would
// corrupt validation for azurerm_virtual_hub bodies. hub_routing_preference is a
// genuine virtualHubs property shared by both kinds, so its enum is safe.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/virtual_hub_resource.go
//     (schema 54-155, expand 186-214, timeouts 47-52)
//   - internal/services/network/validate/virtual_hub_name.go:11-14 (name 1-256)
//   - internal/services/network/route_server_resource.go (route_server contributor;
//     hub_routing_preference default ExpressRoute, timeouts 60m)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualwans:
//     model_virtualhubproperties.go, model_virtualrouterautoscaleconfiguration.go,
//     constants.go (HubRoutingPreference 414-417)
type VirtualHub struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualHub)(nil)

// NewVirtualHub returns knowledge for the virtualHubs resource.
func NewVirtualHub() *VirtualHub {
	return &VirtualHub{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/virtualHubs",
			ApiVersions:  []string{"2025-01-01"},
			// location, address_prefix, sku, virtual_wan_id all replace the hub.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "properties.addressPrefix"},
				{PropertyPath: "properties.sku"},
				{PropertyPath: "properties.virtualWan.id"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// name: universal for both kinds (route_server's 1-80 is a subset).
					MinLength: 1,
					MaxLength: 256,
					Message:   "name must be 1-256 characters",
				},
				{
					PropertyPath:  "properties.sku",
					AllowedValues: []string{"Basic", "Standard"},
					Message:       "sku must be Basic or Standard",
				},
				{
					PropertyPath:  "properties.hubRoutingPreference",
					AllowedValues: []string{"ExpressRoute", "VpnGateway", "ASPath"},
					Message:       "hub_routing_preference must be ExpressRoute, VpnGateway or ASPath",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.virtualRouterAutoScaleConfiguration.minCapacity",
					MinValue:     azwise.Ptr(int64(2)),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.allowBranchToBranchTraffic", Value: false},
				{PropertyPath: "properties.hubRoutingPreference", Value: "ExpressRoute"},
				{PropertyPath: "properties.virtualRouterAutoScaleConfiguration.minCapacity", Value: float64(2)},
			},
		},
	}
}

func init() { azwise.Register(NewVirtualHub()) }
