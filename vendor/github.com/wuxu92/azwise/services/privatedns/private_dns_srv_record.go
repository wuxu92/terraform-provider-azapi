package privatedns

import (
	"math"
	"time"

	"github.com/wuxu92/azwise"
)

// PrivateDnsSrvRecord provides resource knowledge for Microsoft.Network/privateDnsZones/SRV
// (private SRV record sets; ARM type segment is the record type "SRV").
//
// Sources:
//   - internal/services/privatedns/private_dns_srv_record_resource.go (schema + CreateUpdate)
//   - vendor/.../privatedns/2024-06-01/privatedns/model_srvrecord.go (Port, Priority, Target, Weight)
//   - vendor/.../privatedns/2024-06-01/privatedns/id_recordtype.go (.../privateDnsZones/{zone}/SRV/{name})
//
// Notes:
//   - `record` is a Set → properties.srvRecords[*].{priority,weight,port,target}.
//     The nested priority/weight IntBetween(0,65535), port IntBetween(1,65535) and
//     target StringIsNotEmpty validators live at array-element paths that azwise
//     cannot lower — omitted per the array-element rule.
//   - `ttl` is validated IntBetween(1, math.MaxInt32).
type PrivateDnsSrvRecord struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PrivateDnsSrvRecord)(nil)

func NewPrivateDnsSrvRecord() *PrivateDnsSrvRecord {
	return &PrivateDnsSrvRecord{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/privateDnsZones/SRV",
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
				"properties.srvRecords", // record (Required)
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewPrivateDnsSrvRecord()) }
