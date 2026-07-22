package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualNetwork provides resource knowledge for Microsoft.Network/virtualNetworks.
//
// Mirrors azurerm_virtual_network. The network name and resource group live on the
// operational envelope (Required + RequiresReplace by construction), so their
// ForceNew is not repeated here; only body/envelope-path knowledge is.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/virtual_network_resource.go
//     (schema resourceVirtualNetworkSchema, expandVirtualNetworkProperties, CRUD timeouts)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualnetworks:
//     model_virtualnetworkpropertiesformat.go, model_addressspace.go,
//     model_virtualnetworkencryption.go, model_virtualnetworkbgpcommunities.go,
//     constants.go (enum values)
//   - bicep types.json network/microsoft.network/2025-05-01 (body path + flag verification)
//
// Not encoded (deliberate):
//   - resourceGuid / provisioningState are already bicep ReadOnly (flags=2), so the
//     generator marks them Computed and strips them from the PUT body; no azwise
//     ComputedFields entry is needed.
//   - dns_servers is Optional+Computed in AzureRM (dhcpOptions.dnsServers) — a
//     user-settable field, not read-only; left to the mechanical bicep flag.
//   - Every subnet-level rule (address_prefixes MinItems, default_outbound_access_enabled
//     default, private_endpoint_network_policies enum/default, service_delegation name
//     enum, route_table_id/service_endpoint_policy_ids resource IDs) lives under the
//     subnets[*] array element. azwise/azapin cannot lower or resolve a path through an
//     array element, so these are skipped here (documented, not emitted).
//   - bgp_community carries AzureRM's custom "asn:community" validator; it is a
//     semantic rule ported as a network CustomValidator and attached in the resource
//     customizer (internal/native/generator/customizers/virtual_network.go), not here.
type VirtualNetwork struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualNetwork)(nil)

// NewVirtualNetwork returns knowledge for the virtualNetworks resource.
func NewVirtualNetwork() *VirtualNetwork {
	return &VirtualNetwork{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/virtualNetworks",
			ApiVersions:  []string{"2025-05-01"},
			// location (commonschema.Location) and edge_zone (extendedLocation,
			// EdgeZoneOptionalForceNew) both replace the network on change. name and
			// resource_group are envelope-owned and already Required+RequiresReplace.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "extendedLocation"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.encryption.enforcement",
					AllowedValues: []string{"DropUnencrypted", "AllowUnencrypted"},
					Message:       "must be DropUnencrypted or AllowUnencrypted",
				},
				{
					PropertyPath:  "properties.privateEndpointVNetPolicies",
					AllowedValues: []string{"Basic", "Disabled"},
					Message:       "must be Basic or Disabled",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.flowTimeoutInMinutes",
					MinValue:     azwise.Ptr(int64(4)),
					MaxValue:     azwise.Ptr(int64(30)),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.privateEndpointVNetPolicies", Value: "Disabled"},
			},
			// address_space and ip_address_pool are mutually exclusive (AzureRM
			// ExactlyOneOf); both expand under properties.addressSpace as distinct
			// object-level arrays, so the constraint has a representable ARM shape.
			ExactlyOneOf: []azwise.RelationalRule{
				{
					Paths: []string{
						"properties.addressSpace.addressPrefixes",
						"properties.addressSpace.ipamPoolPrefixAllocations",
					},
					Message: "exactly one of address_space or ip_address_pool must be set",
				},
			},
		},
	}
}

func init() { azwise.Register(NewVirtualNetwork()) }
