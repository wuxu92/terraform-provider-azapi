package dns

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DnsCnameRecord provides resource knowledge for Microsoft.Network/dnsZones/CNAME
// (CNAME record sets; ARM type segment is the record type "CNAME").
//
// Sources:
//   - internal/services/dns/dns_cname_record_resource.go (schema + Create/Update)
//   - vendor/.../dns/2018-05-01/recordsets/model_cnamerecord.go
//   - vendor/.../dns/2018-05-01/recordsets/id_recordtype.go (.../dnsZones/{zone}/CNAME/{name})
//
// Notes:
//   - `record` (properties.CNAMERecord) and `target_resource_id`
//     (properties.targetResource) are ExactlyOneOf in AzureRM: exactly one must be set.
type DnsCnameRecord struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DnsCnameRecord)(nil)

func NewDnsCnameRecord() *DnsCnameRecord {
	return &DnsCnameRecord{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/dnsZones/CNAME",
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
			ExactlyOneOf: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.CNAMERecord", "properties.targetResource"},
					Message: "exactly one of `record` or `target_resource_id` must be set",
				},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDnsCnameRecord()) }
