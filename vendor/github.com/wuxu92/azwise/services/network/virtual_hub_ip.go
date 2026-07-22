package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualHubIP provides resource knowledge for
// Microsoft.Network/virtualHubs/ipConfigurations.
//
// Mirrors azurerm_virtual_hub_ip. name, virtual_hub_id (parent) are
// envelope/parent references.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/virtual_hub_ip_resource.go
//     (schema 47-91, expand 122-143, timeouts 34-39)
//   - go-azure-helpers commonids/virtual_hub_ip_configuration.go:118-121 (segment casing)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualwans:
//     model_hubipconfigurationpropertiesformat.go, constants.go (IPAllocationMethod 457-460)
type VirtualHubIP struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualHubIP)(nil)

// NewVirtualHubIP returns knowledge for the ipConfigurations resource.
func NewVirtualHubIP() *VirtualHubIP {
	return &VirtualHubIP{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/virtualHubs/ipConfigurations",
			ApiVersions:  []string{"2025-01-01"},
			// public_ip_address_id and subnet_id are ForceNew; only the private IP
			// fields are updatable.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.publicIPAddress.id"},
				{PropertyPath: "properties.subnet.id"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.privateIPAllocationMethod",
					AllowedValues: []string{"Dynamic", "Static"},
					Message:       "private_ip_allocation_method must be Dynamic or Static",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.privateIPAllocationMethod", Value: "Dynamic"},
			},
			RequiredFields: []string{
				"properties.subnet.id",
			},
		},
	}
}

func init() { azwise.Register(NewVirtualHubIP()) }
