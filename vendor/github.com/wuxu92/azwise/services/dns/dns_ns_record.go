package dns

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DnsNsRecord provides resource knowledge for Microsoft.Network/dnsZones/NS
// (NS record sets; ARM type segment is the record type "NS").
//
// Sources:
//   - internal/services/dns/dns_ns_record_resource.go (schema + Create/Update)
//   - vendor/.../dns/2018-05-01/recordsets/model_nsrecord.go
//   - vendor/.../dns/2018-05-01/recordsets/id_recordtype.go (.../dnsZones/{zone}/NS/{name})
type DnsNsRecord struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DnsNsRecord)(nil)

func NewDnsNsRecord() *DnsNsRecord {
	return &DnsNsRecord{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/dnsZones/NS",
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
				"properties.TTL",       // ttl (Required)
				"properties.NSRecords", // records (Required)
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDnsNsRecord()) }
