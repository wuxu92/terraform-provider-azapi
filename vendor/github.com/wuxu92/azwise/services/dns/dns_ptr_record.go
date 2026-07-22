package dns

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DnsPtrRecord provides resource knowledge for Microsoft.Network/dnsZones/PTR
// (PTR record sets; ARM type segment is the record type "PTR").
//
// Sources:
//   - internal/services/dns/dns_ptr_record_resource.go (schema + CreateUpdate)
//   - vendor/.../dns/2018-05-01/recordsets/model_ptrrecord.go
//   - vendor/.../dns/2018-05-01/recordsets/id_recordtype.go (.../dnsZones/{zone}/PTR/{name})
type DnsPtrRecord struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DnsPtrRecord)(nil)

func NewDnsPtrRecord() *DnsPtrRecord {
	return &DnsPtrRecord{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/dnsZones/PTR",
			ApiVersions:  []string{"2018-05-01"},
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
				"properties.provisioningState",
			},
			RequiredFields: []string{
				"properties.TTL",        // ttl (Required)
				"properties.PTRRecords", // records (Required)
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDnsPtrRecord()) }
