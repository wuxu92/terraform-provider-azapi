package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkManagerIpamPoolStaticCidr provides resource knowledge for
// Microsoft.Network/networkManagers/ipamPools/staticCidrs.
//
// Mirrors azurerm_network_manager_ipam_pool_static_cidr. name and the parent ipam_pool_id
// are envelope-owned (Required + RequiresReplace).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_manager_ipam_pool_static_cidr_resource.go
//   - terraform-provider-azurerm internal/services/network/validate/number_of_ip_addresses.go
//   - go-azure-sdk resource-manager/network/2025-01-01/staticcidrs:
//     model_staticcidrproperties.go
//
// Not encoded (deliberate):
//   - address_prefixes elements validate with validation.IsCIDR — a per-element semantic
//     rule inside an array; belongs on an azapin customizer, noted only.
//   - number_of_ip_addresses_to_allocate validates with validate.NumberOfIpAddresses (a
//     positive-integer-string check); a semantic rule that belongs on an azapin customizer,
//     noted only.
//   - totalNumberOfIPAddresses is server-computed but shares the single StaticCidrProperties
//     model used for Create, so it is not listed as a ComputedField.
type NetworkManagerIpamPoolStaticCidr struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkManagerIpamPoolStaticCidr)(nil)

func NewNetworkManagerIpamPoolStaticCidr() *NetworkManagerIpamPoolStaticCidr {
	return &NetworkManagerIpamPoolStaticCidr{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkManagers/ipamPools/staticCidrs",
			ApiVersions:  []string{"2025-01-01"},
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
			},
			// AzureRM ExactlyOneOf on address_prefixes / number_of_ip_addresses_to_allocate.
			ExactlyOneOf: []azwise.RelationalRule{
				{
					Paths: []string{
						"properties.addressPrefixes",
						"properties.numberOfIPAddressesToAllocate",
					},
					Message: "exactly one of `address_prefixes` or `number_of_ip_addresses_to_allocate` must be set",
				},
			},
		},
	}
}

func init() { azwise.Register(NewNetworkManagerIpamPoolStaticCidr()) }
