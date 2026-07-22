package dns

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DnsMxRecord provides resource knowledge for Microsoft.Network/dnsZones/MX
// (MX record sets; ARM type segment is the record type "MX").
//
// Sources:
//   - internal/services/dns/dns_mx_record_resource.go (schema + CreateUpdate)
//   - vendor/.../dns/2018-05-01/recordsets/model_mxrecord.go
//   - vendor/.../dns/2018-05-01/recordsets/id_recordtype.go (.../dnsZones/{zone}/MX/{name})
//
// Notes:
//   - `name` is Optional with Default "@" in AzureRM, still ForceNew (URL segment).
type DnsMxRecord struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DnsMxRecord)(nil)

func NewDnsMxRecord() *DnsMxRecord {
	return &DnsMxRecord{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/dnsZones/MX",
			ApiVersions:  []string{"2018-05-01"},
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
			ComputedFields: []string{
				"properties.fqdn",
				"properties.provisioningState",
			},
			RequiredFields: []string{
				"properties.TTL",       // ttl (Required)
				"properties.MXRecords", // record (Required)
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDnsMxRecord()) }
