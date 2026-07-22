package dns

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DnsZone provides resource knowledge for Microsoft.Network/dnsZones (public DNS zones).
//
// Sources:
//   - internal/services/dns/dns_zone_resource.go (schema + Create/Read/Update/Delete)
//   - vendor/.../go-azure-sdk/resource-manager/dns/2018-05-01/zones/model_zoneproperties.go
//   - vendor/.../go-azure-sdk/resource-manager/dns/2018-05-01/zones/constants.go (ZoneType)
//
// Notes:
//   - AzureRM only sends Location ("global") + Tags on the zone body; zoneType,
//     registrationVirtualNetworks and resolutionVirtualNetworks are not exposed
//     (private zones live under Microsoft.Network/privateDnsZones).
//   - The `soa_record` block is NOT part of the dnsZones body — AzureRM writes it
//     via a separate SOA record-set call (Microsoft.Network/dnsZones/SOA), so its
//     fields/validation belong to that sub-resource, not here.
type DnsZone struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DnsZone)(nil)

func NewDnsZone() *DnsZone {
	return &DnsZone{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/dnsZones",
			ApiVersions:  []string{"2018-05-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"}, // name (ForceNew)
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ComputedFields: []string{
				"properties.nameServers",
				"properties.numberOfRecordSets",
				"properties.maxNumberOfRecordSets",
				"properties.maxNumberOfRecordsPerRecordSet",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.zoneType", Value: "Public"}, // ZoneType default; private zones are a separate ARM type
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDnsZone()) }
