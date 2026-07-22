package privatedns

import (
	"math"
	"time"

	"github.com/wuxu92/azwise"
)

// PrivateDnsPtrRecord provides resource knowledge for Microsoft.Network/privateDnsZones/PTR
// (private PTR record sets; ARM type segment is the record type "PTR").
//
// Sources:
//   - internal/services/privatedns/private_dns_ptr_record_resource.go (schema + CreateUpdate)
//   - vendor/.../privatedns/2024-06-01/privatedns/model_ptrrecord.go (Ptrdname)
//   - vendor/.../privatedns/2024-06-01/privatedns/id_recordtype.go (.../privateDnsZones/{zone}/PTR/{name})
//
// Notes:
//   - `records` is a Set of strings → properties.ptrRecords[*].ptrdname.
//   - `ttl` is validated IntBetween(1, math.MaxInt32).
type PrivateDnsPtrRecord struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PrivateDnsPtrRecord)(nil)

func NewPrivateDnsPtrRecord() *PrivateDnsPtrRecord {
	return &PrivateDnsPtrRecord{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/privateDnsZones/PTR",
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
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(math.MaxInt32)),
					Message:      "ttl must be between 1 and 2147483647",
				},
			},
			ComputedFields: []string{
				"properties.fqdn",
				"properties.isAutoRegistered",
			},
			RequiredFields: []string{
				"properties.ttl",        // ttl (Required)
				"properties.ptrRecords", // records (Required)
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewPrivateDnsPtrRecord()) }
