package privatedns

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PrivateDnsARecord provides resource knowledge for Microsoft.Network/privateDnsZones/A
// (private A record sets; ARM type segment is the record type "A").
//
// Sources:
//   - internal/services/privatedns/private_dns_a_record_resource.go (schema + CreateUpdate)
//   - vendor/.../privatedns/2024-06-01/privatedns/model_recordsetproperties.go (ARecords, Ttl)
//   - vendor/.../privatedns/2024-06-01/privatedns/id_recordtype.go (.../privateDnsZones/{zone}/A/{name})
//
// Notes:
//   - `name`/`private_dns_zone_id` are URL/parent segments; only the record-set
//     name is modeled as ForceNew here.
//   - `records` is a Set of IPv4 strings → properties.aRecords[*].ipv4Address;
//     MaxItems(20) applies to the array itself (properties.aRecords).
type PrivateDnsARecord struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PrivateDnsARecord)(nil)

func NewPrivateDnsARecord() *PrivateDnsARecord {
	return &PrivateDnsARecord{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/privateDnsZones/A",
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
			ArrayRules: []azwise.ArrayRule{
				{PropertyPath: "properties.aRecords", MaxItems: 20, Message: "at most 20 A records are allowed"},
			},
			ComputedFields: []string{
				"properties.fqdn",
				"properties.isAutoRegistered",
			},
			RequiredFields: []string{
				"properties.ttl",      // ttl (Required)
				"properties.aRecords", // records (Required)
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewPrivateDnsARecord()) }
