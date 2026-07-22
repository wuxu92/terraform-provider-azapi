package privatedns

import (
	"math"
	"time"

	"github.com/wuxu92/azwise"
)

// PrivateDnsCNameRecord provides resource knowledge for Microsoft.Network/privateDnsZones/CNAME
// (private CNAME record sets; ARM type segment is the record type "CNAME").
//
// Sources:
//   - internal/services/privatedns/private_dns_cname_record_resource.go (schema + CreateUpdate)
//   - vendor/.../privatedns/2024-06-01/privatedns/model_cnamerecord.go (Cname)
//   - vendor/.../privatedns/2024-06-01/privatedns/id_recordtype.go (.../privateDnsZones/{zone}/CNAME/{name})
//
// Notes:
//   - `record` is a single string → properties.cnameRecord.cname (Required).
//   - `ttl` is validated IntBetween(0, math.MaxInt32).
type PrivateDnsCNameRecord struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PrivateDnsCNameRecord)(nil)

func NewPrivateDnsCNameRecord() *PrivateDnsCNameRecord {
	return &PrivateDnsCNameRecord{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/privateDnsZones/CNAME",
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
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.ttl",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(math.MaxInt32)),
					Message:      "ttl must be between 0 and 2147483647",
				},
			},
			ComputedFields: []string{
				"properties.fqdn",
				"properties.isAutoRegistered",
			},
			RequiredFields: []string{
				"properties.ttl",                // ttl (Required)
				"properties.cnameRecord.cname",  // record (Required)
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewPrivateDnsCNameRecord()) }
