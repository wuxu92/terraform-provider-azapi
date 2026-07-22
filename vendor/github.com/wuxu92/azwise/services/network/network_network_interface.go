package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkInterface provides resource knowledge for Microsoft.Network/networkInterfaces.
//
// Mirrors azurerm_network_interface. The ip_configuration block expands into
// properties.ipConfigurations (an array — element-level rules are not encoded);
// the top-level scalar options map under properties.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_interface_resource.go
//     (resourceNetworkInterface schema lines 36-216, expand lines 218-314)
//   - go-azure-sdk resource-manager/network/2025-01-01/networkinterfaces:
//     id via commonids NetworkInterfaceId (segment "networkInterfaces"),
//     model_networkinterfacepropertiesformat.go / model_networkinterfacednssettings.go
//     (enableIPForwarding, enableAcceleratedNetworking, auxiliaryMode, auxiliarySku
//     json tags), constants.go (NetworkInterfaceAuxiliaryMode,
//     NetworkInterfaceAuxiliarySku)
//
// Not encoded (deliberate):
//   - The five association resources — network_interface_application_gateway_backend_address_pool_association,
//     network_interface_application_security_group_association,
//     network_interface_backend_address_pool_association,
//     network_interface_nat_rule_association and
//     network_interface_security_group_association — all manage the SAME
//     networkInterfaces resource: they fold their references into
//     properties.ipConfigurations[*] (backend pools / ASGs / NAT rules) or
//     properties.networkSecurityGroup on the parent body. None is a distinct ARM
//     resource type, so no separate knowledge files are emitted.
//   - ip_configuration sub-fields (private_ip_address_version enum,
//     private_ip_address_allocation enum) live under the ipConfigurations[*]
//     array element; azwise cannot resolve a path through an array element, so
//     they are skipped (documented, not emitted).
//   - mac_address, private_ip_address(es), virtual_machine_id, applied_dns_servers,
//     internal_domain_name_suffix are Computed read-only; already bicep ReadOnly.
type NetworkInterface struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkInterface)(nil)

// NewNetworkInterface returns knowledge for the networkInterfaces resource.
func NewNetworkInterface() *NetworkInterface {
	return &NetworkInterface{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkInterfaces",
			ApiVersions:  []string{"2025-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "extendedLocation"}, // edge_zone
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.auxiliaryMode",
					AllowedValues: []string{"AcceleratedConnections", "Floating", "MaxConnections", "None"},
					Message:       "must be AcceleratedConnections, Floating, MaxConnections or None",
				},
				{
					PropertyPath:  "properties.auxiliarySku",
					AllowedValues: []string{"A8", "A4", "A1", "A2", "None"},
					Message:       "must be one of A1, A2, A4, A8 or None",
				},
			},
			// auxiliary_mode and auxiliary_sku are mutually RequiredWith in AzureRM
			// (bidirectional); each must accompany the other.
			RequiredWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.auxiliaryMode", "properties.auxiliarySku"},
					Message: "auxiliary_mode and auxiliary_sku must be set together",
				},
				{
					Paths:   []string{"properties.auxiliarySku", "properties.auxiliaryMode"},
					Message: "auxiliary_mode and auxiliary_sku must be set together",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.enableIPForwarding", Value: false},
				{PropertyPath: "properties.enableAcceleratedNetworking", Value: false},
			},
		},
	}
}

func init() { azwise.Register(NewNetworkInterface()) }
