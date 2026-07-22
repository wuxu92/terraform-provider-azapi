package dns

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DnsSrvRecord provides resource knowledge for Microsoft.Network/dnsZones/SRV
// (SRV record sets; ARM type segment is the record type "SRV").
//
// Sources:
//   - internal/services/dns/dns_srv_record_resource.go (schema + Create/Update)
//   - vendor/.../dns/2018-05-01/recordsets/model_srvrecord.go
//   - vendor/.../dns/2018-05-01/recordsets/id_recordtype.go (.../dnsZones/{zone}/SRV/{name})
type DnsSrvRecord struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DnsSrvRecord)(nil)

func NewDnsSrvRecord() *DnsSrvRecord {
	return &DnsSrvRecord{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/dnsZones/SRV",
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
				"properties.SRVRecords", // record (Required)
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDnsSrvRecord()) }
