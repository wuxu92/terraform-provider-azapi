package privatedns

import (
	"math"
	"time"

	"github.com/wuxu92/azwise"
)

// PrivateDnsMxRecord provides resource knowledge for Microsoft.Network/privateDnsZones/MX
// (private MX record sets; ARM type segment is the record type "MX").
//
// Sources:
//   - internal/services/privatedns/private_dns_mx_record_resource.go (schema + CreateUpdate)
//   - vendor/.../privatedns/2024-06-01/privatedns/model_mxrecord.go (Exchange, Preference)
//   - vendor/.../privatedns/2024-06-01/privatedns/id_recordtype.go (.../privateDnsZones/{zone}/MX/{name})
//
// Notes:
//   - `name` is Optional with Default "@" in AzureRM, still ForceNew (URL segment).
//   - `record` is a Set → properties.mxRecords[*].{preference,exchange}. The nested
//     preference IntBetween(0,65535) and exchange StringIsNotEmpty validators live
//     at array-element paths that azwise cannot lower — omitted per the
//     array-element rule.
//   - `ttl` is validated IntBetween(1, math.MaxInt32).
type PrivateDnsMxRecord struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PrivateDnsMxRecord)(nil)

func NewPrivateDnsMxRecord() *PrivateDnsMxRecord {
	return &PrivateDnsMxRecord{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/privateDnsZones/MX",
			ApiVersions:  []string{"2024-06-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"}, // relative record set name (ForceNew, default "@")
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
				"properties.ttl",       // ttl (Required)
				"properties.mxRecords", // record (Required)
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewPrivateDnsMxRecord()) }
