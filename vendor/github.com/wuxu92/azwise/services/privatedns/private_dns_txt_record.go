package privatedns

import (
	"math"
	"time"

	"github.com/wuxu92/azwise"
)

// PrivateDnsTxtRecord provides resource knowledge for Microsoft.Network/privateDnsZones/TXT
// (private TXT record sets; ARM type segment is the record type "TXT").
//
// Sources:
//   - internal/services/privatedns/private_dns_txt_record_resource.go (schema + CreateUpdate)
//   - vendor/.../privatedns/2024-06-01/privatedns/model_txtrecord.go (Value)
//   - vendor/.../privatedns/2024-06-01/privatedns/id_recordtype.go (.../privateDnsZones/{zone}/TXT/{name})
//
// Notes:
//   - `record` is a Set → properties.txtRecords[*].value[*]. The nested
//     value StringLenBetween(1,1024) validator lives at an array-element path that
//     azwise cannot lower — omitted per the array-element rule.
//   - `ttl` is validated IntBetween(1, math.MaxInt32).
type PrivateDnsTxtRecord struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PrivateDnsTxtRecord)(nil)

func NewPrivateDnsTxtRecord() *PrivateDnsTxtRecord {
	return &PrivateDnsTxtRecord{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/privateDnsZones/TXT",
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
				"properties.txtRecords", // record (Required)
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewPrivateDnsTxtRecord()) }
