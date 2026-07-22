package dns

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DnsCaaRecord provides resource knowledge for Microsoft.Network/dnsZones/CAA
// (CAA record sets; ARM type segment is the record type "CAA").
//
// Sources:
//   - internal/services/dns/dns_caa_record_resource.go (schema + CreateUpdate)
//   - vendor/.../dns/2018-05-01/recordsets/model_caarecord.go
//   - vendor/.../dns/2018-05-01/recordsets/id_recordtype.go (.../dnsZones/{zone}/CAA/{name})
//
// Notes:
//   - The `tag` StringInSlice enum (issue/issuewild/iodef/contactemail) lives at the
//     array-element path properties.caaRecords[*].tag, which azwise/azwise_validate
//     cannot lower — omitted per the array-element rule.
type DnsCaaRecord struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DnsCaaRecord)(nil)

func NewDnsCaaRecord() *DnsCaaRecord {
	return &DnsCaaRecord{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/dnsZones/CAA",
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
				"properties.caaRecords", // record (Required)
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDnsCaaRecord()) }
