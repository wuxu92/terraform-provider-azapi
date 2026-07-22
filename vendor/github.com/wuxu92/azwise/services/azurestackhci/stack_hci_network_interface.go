package azurestackhci

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StackHCINetworkInterface provides resource knowledge for
// Microsoft.AzureStackHCI/networkInterfaces.
//
// Mirrors azurerm_stack_hci_network_interface. name / resource_group_name / location live on
// the operational envelope; location is surfaced here as ForceNew (commonschema.Location).
// Every settable argument on this resource is ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/azurestackhci/stack_hci_network_interface_resource.go:62-129
//     (schema: ForceNew, validators, MaxItems, ip_configuration block, computed gateway/prefix)
//   - terraform-provider-azurerm internal/services/azurestackhci/stack_hci_network_interface_resource.go:159-180,301-324
//     (create/expand mapping: ipConfigurations, dnsSettings.dnsServers, macAddress, extendedLocation)
//   - go-azure-sdk resource-manager/azurestackhci/2024-01-01/networkinterfaces:
//     model_networkinterfaceproperties.go (ipConfigurations, dnsSettings, macAddress),
//     model_ipconfigurationproperties.go, model_interfacednssettings.go
//
// Not encoded (deliberate):
//   - custom_location_id carries a resource-ID semantic validator (customlocations) that
//     belongs in an azapin customizer (validators.AzureResourceID) attached to
//     extendedLocation.name, not a StringRule.
//   - dns_servers is a list of validation.IsIPv4Address elements on
//     properties.dnsSettings.dnsServers[*]; azwise cannot resolve a path through an array
//     element, so it is documented, not emitted.
//   - Every ip_configuration-level rule (subnet_id resource ID, private_ip_address
//     IsIPv4Address) and the computed gateway / prefix_length attributes live under the
//     properties.ipConfigurations[*] array element. azwise/azapin cannot lower or resolve a
//     path through an array element, so these are skipped here (documented, not emitted).
type StackHCINetworkInterface struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StackHCINetworkInterface)(nil)

// NewStackHCINetworkInterface returns knowledge for the Azure Stack HCI networkInterfaces resource.
func NewStackHCINetworkInterface() *StackHCINetworkInterface {
	return &StackHCINetworkInterface{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AzureStackHCI/networkInterfaces",
			ApiVersions:  []string{"2024-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "extendedLocation.name"},
				{PropertyPath: "properties.ipConfigurations"},
				{PropertyPath: "properties.dnsSettings.dnsServers"},
				{PropertyPath: "properties.macAddress"},
			},
			RequiredFields: []string{
				"properties.ipConfigurations",
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
					// name: begin/end alphanumeric, 2-80 chars, alphanumeric + - . _ inside.
					Regex:   `^[a-zA-Z0-9][\-\.\_a-zA-Z0-9]{0,78}[a-zA-Z0-9]$`,
					Message: "name must begin and end with an alphanumeric character, be between 2 and 80 characters in length and can only contain alphanumeric characters, hyphens, periods or underscores",
				},
			},
		},
	}
}

func init() { azwise.Register(NewStackHCINetworkInterface()) }
