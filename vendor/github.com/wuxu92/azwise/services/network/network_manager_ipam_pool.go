package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkManagerIpamPool provides resource knowledge for
// Microsoft.Network/networkManagers/ipamPools.
//
// Mirrors azurerm_network_manager_ipam_pool. name and the parent network_manager_id are
// envelope-owned (Required + RequiresReplace).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_manager_ipam_pool_resource.go
//   - go-azure-sdk resource-manager/network/2025-01-01/ipampools:
//     model_ipampoolproperties.go
//
// Not encoded (deliberate):
//   - address_prefixes elements validate with validation.IsCIDR — a per-element semantic
//     rule inside an array; belongs on an azapin customizer, noted only.
//   - ipAddressType is not exposed by AzureRM but IS present in the Create model
//     (IPamPoolProperties.IPAddressType), so it is NOT a ComputedField.
type NetworkManagerIpamPool struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkManagerIpamPool)(nil)

func NewNetworkManagerIpamPool() *NetworkManagerIpamPool {
	return &NetworkManagerIpamPool{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkManagers/ipamPools",
			ApiVersions:  []string{"2025-01-01"},
			// location, address_prefixes and parent_pool_name are all ForceNew in AzureRM.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "properties.addressPrefixes"},
				{PropertyPath: "properties.parentPoolName"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name validation (empty PropertyPath = the name attribute).
					Regex:     `^[a-zA-Z0-9\_\.\-]{1,64}$`,
					MinLength: 1,
					MaxLength: 64,
					Message:   "must be 1-64 characters and contain only letters, numbers, underscores, periods and hyphens",
				},
				{
					PropertyPath: "properties.displayName",
					Regex:        `^[a-zA-Z0-9\_\.\-]{1,64}$`,
					MinLength:    1,
					MaxLength:    64,
					Message:      "must be 1-64 characters and contain only letters, numbers, underscores, periods and hyphens",
				},
				{
					PropertyPath: "properties.parentPoolName",
					Regex:        `^[a-zA-Z0-9\_\.\-]{1,64}$`,
					MinLength:    1,
					MaxLength:    64,
					Message:      "must be 1-64 characters and contain only letters, numbers, underscores, periods and hyphens",
				},
			},
			// address_prefixes is Required.
			RequiredFields: []string{"properties.addressPrefixes"},
		},
	}
}

func init() { azwise.Register(NewNetworkManagerIpamPool()) }
