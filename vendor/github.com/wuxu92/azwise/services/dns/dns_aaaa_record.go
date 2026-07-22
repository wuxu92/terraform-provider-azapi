package dns

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DnsAaaaRecord provides resource knowledge for Microsoft.Network/dnsZones/AAAA
// (AAAA record sets; ARM type segment is the record type "AAAA").
//
// Sources:
//   - internal/services/dns/dns_aaaa_record_resource.go (schema + CreateUpdate)
//   - vendor/.../dns/2018-05-01/recordsets/model_recordsetproperties.go
//   - vendor/.../dns/2018-05-01/recordsets/id_recordtype.go (.../dnsZones/{zone}/AAAA/{name})
//
// Notes:
//   - `records` (properties.AAAARecords) and `target_resource_id`
//     (properties.targetResource) are mutually exclusive (AzureRM ConflictsWith).
type DnsAaaaRecord struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DnsAaaaRecord)(nil)

func NewDnsAaaaRecord() *DnsAaaaRecord {
	return &DnsAaaaRecord{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/dnsZones/AAAA",
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
					Paths:   []string{"properties.AAAARecords", "properties.targetResource"},
					Message: "`records` and `target_resource_id` cannot be set together",
				},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDnsAaaaRecord()) }
