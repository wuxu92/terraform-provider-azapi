package azurestackhci

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StackHCILogicalNetwork provides resource knowledge for Microsoft.AzureStackHCI/logicalNetworks.
//
// Mirrors azurerm_stack_hci_logical_network. name / resource_group_name / location live on the
// operational envelope; location is surfaced here as ForceNew (commonschema.Location). Every
// argument on this resource is ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/azurestackhci/stack_hci_logical_network_resource.go:73-202
//     (schema: ForceNew, validators, MaxItems, subnet block)
//   - terraform-provider-azurerm internal/services/azurestackhci/stack_hci_logical_network_resource.go:232-247,371-471
//     (create/expand mapping: vmSwitchName, subnets, dhcpOptions.dnsServers, extendedLocation)
//   - go-azure-sdk resource-manager/azurestackhci/2024-01-01/logicalnetworks:
//     model_logicalnetworkproperties.go (vmSwitchName, subnets, dhcpOptions),
//     model_subnetpropertiesformat.go, model_logicalnetworkpropertiesdhcpoptions.go
//
// Not encoded (deliberate):
//   - custom_location_id carries customlocations.ValidateCustomLocationID — a resource-ID
//     semantic validator that belongs in an azapin customizer (validators.AzureResourceID)
//     attached to extendedLocation.name, not a StringRule.
//   - dns_servers is a list of validation.IsIPv4Address elements — a per-element validator on
//     properties.dhcpOptions.dnsServers[*]; azwise cannot resolve a path through an array
//     element, so it is documented, not emitted.
//   - Every subnet-level rule (ip_allocation_method enum IPAllocationMethodEnum, address_prefix
//     IsCIDR, ip_pool start/end IsIPv4Address, route address_prefix IsCIDR / next_hop_ip_address
//     IsIPv4Address / name regex, vlan_id IntAtLeast(0)) lives under the properties.subnets[*]
//     array element. azwise/azapin cannot lower or resolve a path through an array element, so
//     these are skipped here (documented, not emitted).
type StackHCILogicalNetwork struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StackHCILogicalNetwork)(nil)

// NewStackHCILogicalNetwork returns knowledge for the Azure Stack HCI logicalNetworks resource.
func NewStackHCILogicalNetwork() *StackHCILogicalNetwork {
	return &StackHCILogicalNetwork{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AzureStackHCI/logicalNetworks",
			ApiVersions:  []string{"2024-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "extendedLocation.name"},
				{PropertyPath: "properties.vmSwitchName"},
				{PropertyPath: "properties.subnets"},
				{PropertyPath: "properties.dhcpOptions.dnsServers"},
			},
			RequiredFields: []string{
				"properties.vmSwitchName",
				"properties.subnets",
				// AzureRM always sends the custom location as the extended location.
				"extendedLocation.name",
				"extendedLocation.type",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// name: begin/end alphanumeric, 2-64 chars, alphanumeric + - . _ inside.
					Regex:   `^[a-zA-Z0-9][\-\.\_a-zA-Z0-9]{0,62}[a-zA-Z0-9]$`,
					Message: "name must begin and end with an alphanumeric character, be between 2 and 64 characters in length and can only contain alphanumeric characters, hyphens, periods or underscores",
				},
			},
			ArrayRules: []azwise.ArrayRule{
				{
					PropertyPath: "properties.subnets",
					MaxItems:     1,
					Message:      "only a single subnet is supported",
				},
			},
		},
	}
}

func init() { azwise.Register(NewStackHCILogicalNetwork()) }
