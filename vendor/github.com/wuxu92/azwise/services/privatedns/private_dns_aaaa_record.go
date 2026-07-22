package privatedns

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PrivateDnsAaaaRecord provides resource knowledge for Microsoft.Network/privateDnsZones/AAAA
// (private AAAA record sets; ARM type segment is the record type "AAAA").
//
// Sources:
//   - internal/services/privatedns/private_dns_aaaa_record_resource.go (schema + CreateUpdate)
//   - vendor/.../privatedns/2024-06-01/privatedns/model_recordsetproperties.go (AaaaRecords, Ttl)
//   - vendor/.../privatedns/2024-06-01/privatedns/id_recordtype.go (.../privateDnsZones/{zone}/AAAA/{name})
//
// Notes:
//   - `records` is a Set of IPv6 strings → properties.aaaaRecords[*].ipv6Address.
type PrivateDnsAaaaRecord struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PrivateDnsAaaaRecord)(nil)

func NewPrivateDnsAaaaRecord() *PrivateDnsAaaaRecord {
	return &PrivateDnsAaaaRecord{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/privateDnsZones/AAAA",
			ApiVersions:  []string{"2024-06-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"}, // relative record set name (ForceNew)
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ComputedFields: []string{
				"properties.fqdn",
				"properties.isAutoRegistered",
			},
			RequiredFields: []string{
				"properties.ttl",         // ttl (Required)
				"properties.aaaaRecords", // records (Required)
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewPrivateDnsAaaaRecord()) }
