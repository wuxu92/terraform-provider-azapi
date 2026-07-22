package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Subnet provides resource knowledge for Microsoft.Network/virtualNetworks/subnets.
//
// Mirrors azurerm_subnet. name, virtual_network_name and resource_group are
// envelope-owned (name + virtual_network_name are RequiresReplace by
// construction); the subnet body carries no replacement-forcing property.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/subnet_resource.go
//     (schema, expandSubnet*, CRUD timeouts 30/5/30/30)
//   - go-azure-sdk resource-manager/network/2025-01-01/subnets:
//     model_subnetpropertiesformat.go, constants.go
//     (VirtualNetworkPrivateEndpointNetworkPolicies, SharingScope,
//     VirtualNetworkPrivateLinkServiceNetworkPolicies)
//   - go-azure-helpers commonids/subnet.go (segment casing "virtualNetworks/subnets")
//
// Not encoded (deliberate):
//   - azurerm_subnet_network_security_group_association,
//     azurerm_subnet_route_table_association and
//     azurerm_subnet_nat_gateway_association fold into this subnet body as
//     properties.networkSecurityGroup.id, properties.routeTable.id and
//     properties.natGateway.id respectively. They are separate TF resources but
//     the same ARM subnet PUT; their single id fields carry no extra declarative
//     value constraint. Skipped, not emitted (documented).
//   - address_prefixes, delegation, service_endpoints and ip_address_pool expand
//     into arrays (properties.addressPrefixes / delegations[*] /
//     serviceEndpoints[*] / ipamPoolPrefixAllocations[*]); their element-level
//     rules (delegation service_delegation name enum, action enum) live under
//     "[*]" which azwise/azapin cannot lower. Skipped.
//   - the CustomizeDiff forbidding sharing_scope when default_outbound_access_enabled
//     is true is a value-conditional cross-field rule with no declarative
//     equivalent; left to AzureRM/API enforcement.
type Subnet struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Subnet)(nil)

// NewSubnet returns knowledge for the virtualNetworks/subnets resource.
func NewSubnet() *Subnet {
	return &Subnet{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/virtualNetworks/subnets",
			ApiVersions:  []string{"2025-01-01", "2023-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "properties.privateEndpointNetworkPolicies",
					// Full ARM SDK set; AzureRM restricts but AzAPI sends raw ARM values.
					AllowedValues: []string{"Disabled", "Enabled", "NetworkSecurityGroupEnabled", "RouteTableEnabled"},
					Message:       "private_endpoint_network_policies must be Disabled, Enabled, NetworkSecurityGroupEnabled or RouteTableEnabled",
				},
				{
					PropertyPath: "properties.privateLinkServiceNetworkPolicies",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "private_link_service_network_policies must be Disabled or Enabled",
				},
				{
					PropertyPath: "properties.sharingScope",
					// Full ARM SDK set (AzureRM currently only exposes Tenant).
					AllowedValues: []string{"DelegatedServices", "Tenant"},
					Message:       "sharing_scope must be DelegatedServices or Tenant",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.defaultOutboundAccess", Value: true},
				{PropertyPath: "properties.privateEndpointNetworkPolicies", Value: "Disabled"},
				// private_link_service_network_policies_enabled default true → "Enabled".
				{PropertyPath: "properties.privateLinkServiceNetworkPolicies", Value: "Enabled"},
			},
		},
	}
}

func init() { azwise.Register(NewSubnet()) }
