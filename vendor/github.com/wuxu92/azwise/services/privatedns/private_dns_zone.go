package privatedns

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PrivateDnsZone provides resource knowledge for Microsoft.Network/privateDnsZones
// (private DNS zones). Distinct from the public Microsoft.Network/dnsZones type —
// privateDnsZones is a separate ARM provider path.
//
// Sources:
//   - internal/services/privatedns/private_dns_zone_resource.go (schema + Create/Read/Update/Delete)
//   - vendor/.../privatedns/2024-06-01/privatezones/model_privatezoneproperties.go
//   - vendor/.../privatedns/2024-06-01/privatezones/id_privatednszone.go (.../privateDnsZones/{name})
//
// Notes:
//   - AzureRM only sends Location ("global") + Tags on the zone body; every
//     PrivateZoneProperties field is server-computed (read-only) and absent from
//     the create payload.
//   - The `soa_record` block is NOT part of the privateDnsZones body — AzureRM
//     writes it via a separate SOA record-set call (privateDnsZones/{zone}/SOA/@),
//     so its fields/validation belong to that sub-resource, not here.
type PrivateDnsZone struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PrivateDnsZone)(nil)

func NewPrivateDnsZone() *PrivateDnsZone {
	return &PrivateDnsZone{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/privateDnsZones",
			ApiVersions:  []string{"2024-06-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"}, // name (ForceNew, URL segment)
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ComputedFields: []string{
				"properties.internalId",
				"properties.maxNumberOfRecordSets",
				"properties.maxNumberOfVirtualNetworkLinks",
				"properties.maxNumberOfVirtualNetworkLinksWithRegistration",
				"properties.numberOfRecordSets",
				"properties.numberOfVirtualNetworkLinks",
				"properties.numberOfVirtualNetworkLinksWithRegistration",
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewPrivateDnsZone()) }
