package dns

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DnsARecord provides resource knowledge for Microsoft.Network/dnsZones/A
// (A record sets; ARM type segment is the record type "A").
//
// Sources:
//   - internal/services/dns/dns_a_record_resource.go (schema + CreateUpdate)
//   - vendor/.../dns/2018-05-01/recordsets/model_recordsetproperties.go
//   - vendor/.../dns/2018-05-01/recordsets/id_recordtype.go (.../dnsZones/{zone}/A/{name})
//
// Notes:
//   - `name`/`zone_name` are ID (URL) segments, ForceNew at the envelope; only the
//     record-set name is modeled as ForceNew here.
//   - `records` (properties.ARecords) and `target_resource_id` (properties.targetResource)
//     are mutually exclusive (AzureRM ConflictsWith); AzureRM also requires at least
//     one of them at runtime, but that is a code check, not a schema constraint.
type DnsARecord struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DnsARecord)(nil)

func NewDnsARecord() *DnsARecord {
	return &DnsARecord{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/dnsZones/A",
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
				"properties.TTL", // ttl (Required)
			},
			ConflictsWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.ARecords", "properties.targetResource"},
					Message: "`records` and `target_resource_id` cannot be set together",
				},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDnsARecord()) }
