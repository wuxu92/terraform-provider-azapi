package privatednsresolver

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DnsResolver provides resource knowledge for Microsoft.Network/dnsResolvers.
//
// TF resource: azurerm_private_dns_resolver
//
// Sources:
//   - internal/services/privatednsresolver/private_dns_resolver_resource.go
//     (schema lines 45-67; Create/Update/Read/Delete timeouts lines 75/119/158/202)
//   - vendor/.../go-azure-sdk/resource-manager/dnsresolver/2022-07-01/dnsresolvers/model_dnsresolverproperties.go
//   - vendor/.../dnsresolvers/id_dnsresolver.go (ARM type segment "dnsResolvers")
//
// Notes:
//   - virtual_network_id maps to properties.virtualNetwork.id (SubResource), Required + ForceNew.
//   - name/location/resource_group_name are envelope fields (also ForceNew) and are not
//     emitted as body ForceNew rules.
//   - dnsResolverState is settable in the ARM model (pointer), so it is NOT a computed field.
type DnsResolver struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DnsResolver)(nil)

func NewDnsResolver() *DnsResolver {
	return &DnsResolver{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/dnsResolvers",
			ApiVersions:  []string{"2022-07-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.virtualNetwork.id"}, // virtual_network_id (ForceNew)
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.virtualNetwork.id",
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.resourceGuid",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDnsResolver()) }
