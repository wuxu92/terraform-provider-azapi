package dns

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DnsTxtRecord provides resource knowledge for Microsoft.Network/dnsZones/TXT
// (TXT record sets; ARM type segment is the record type "TXT").
//
// Sources:
//   - internal/services/dns/dns_txt_record_resource.go (schema + CreateUpdate)
//   - vendor/.../dns/2018-05-01/recordsets/model_txtrecord.go
//   - vendor/.../dns/2018-05-01/recordsets/id_recordtype.go (.../dnsZones/{zone}/TXT/{name})
//
// Notes:
//   - The `value` StringLenBetween(1, 4096) constraint applies to each string at
//     the array-element path properties.TXTRecords[*].value, which azwise cannot
//     lower — omitted per the array-element rule.
type DnsTxtRecord struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DnsTxtRecord)(nil)

func NewDnsTxtRecord() *DnsTxtRecord {
	return &DnsTxtRecord{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/dnsZones/TXT",
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
				"properties.TXTRecords", // record (Required)
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDnsTxtRecord()) }
