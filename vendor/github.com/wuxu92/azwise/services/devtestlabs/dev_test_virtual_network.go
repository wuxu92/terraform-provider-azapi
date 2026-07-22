package devtestlabs

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DevTestVirtualNetwork provides resource knowledge for Microsoft.DevTestLab/labs/virtualNetworks.
//
// Contributing Terraform resource: azurerm_dev_test_virtual_network.
//
// Sources:
//   - terraform-provider-azurerm internal/services/devtestlabs/dev_test_virtual_network_resource.go
//     (schema L52-144, ValidateDevTestVirtualNetworkName L308-313, subnet expanders L315-375)
//   - go-azure-sdk resource-manager/devtestlab/2018-09-15/virtualnetworks:
//     model_virtualnetworkproperties.go, constants.go (TransportProtocol, UsagePermissionType),
//     id_virtualnetwork.go.
//
// Notes:
//   - name/lab_name are envelope-owned (ARM name segment + parent labs/{labName}) and ForceNew;
//     only the name regex is emitted (empty PropertyPath).
//   - description maps to properties.description.
//   - The subnet block maps to properties.subnetOverrides (an ARRAY). Its enum fields —
//     use_in_virtual_machine_creation / use_public_ip_address (UsagePermissionType:
//     Allow/Default/Deny) and shared_public_ip_address.allowed_ports.transport_protocol
//     (Tcp/Udp) — live under array-element paths (properties.subnetOverrides[*].*), which are
//     unsupported for rule lowering, so they are documented rather than emitted.
type DevTestVirtualNetwork struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DevTestVirtualNetwork)(nil)

func NewDevTestVirtualNetwork() *DevTestVirtualNetwork {
	return &DevTestVirtualNetwork{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DevTestLab/labs/virtualNetworks",
			ApiVersions:  []string{"2018-09-15"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// Resource name validation (empty PropertyPath validates the name).
				{Regex: `^[A-Za-z0-9_-]+$`, Message: "Virtual Network Name can only include alphanumeric characters, underscores, hyphens."},
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.uniqueIdentifier",
			},
		},
	}
}

func init() { azwise.Register(NewDevTestVirtualNetwork()) }
