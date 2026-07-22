package privatednsresolver

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DnsForwardingRulesetForwardingRule provides resource knowledge for
// Microsoft.Network/dnsForwardingRulesets/forwardingRules.
//
// TF resource: azurerm_private_dns_resolver_forwarding_rule
//
// Sources:
//   - internal/services/privatednsresolver/private_dns_resolver_forwarding_rule_resource.go
//     (schema lines 49-105; enabled->forwardingRuleState lines 139-146;
//     Create/Update/Read/Delete timeouts lines 113/168/227/276)
//   - vendor/.../dnsresolver/2022-07-01/forwardingrules/model_forwardingruleproperties.go
//   - vendor/.../forwardingrules/constants.go (ForwardingRuleState: Enabled, Disabled)
//   - vendor/.../forwardingrules/id_forwardingrule.go
//     (child of dnsForwardingRulesets; segment "forwardingRules")
//
// Notes:
//   - dns_forwarding_ruleset_id is the parent-resource reference (envelope), not a body field.
//   - domain_name maps to properties.domainName (Required + ForceNew).
//   - enabled (bool, default true) maps to properties.forwardingRuleState:
//     true -> "Enabled", false -> "Disabled"; declared as a DefaultValue of "Enabled".
//   - target_dns_servers is Required -> properties.targetDnsServers; its element
//     fields (ip_address Required, port Optional) are array-element paths and are
//     not expressible as declarative rules here — skipped.
//   - metadata maps to properties.metadata (map, no key rules emitted).
type DnsForwardingRulesetForwardingRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DnsForwardingRulesetForwardingRule)(nil)

func NewDnsForwardingRulesetForwardingRule() *DnsForwardingRulesetForwardingRule {
	return &DnsForwardingRulesetForwardingRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/dnsForwardingRulesets/forwardingRules",
			ApiVersions:  []string{"2022-07-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.domainName"}, // domain_name (ForceNew)
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.forwardingRuleState",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "forwardingRuleState must be Enabled or Disabled",
				},
			},
			RequiredFields: []string{
				"properties.domainName",
				"properties.targetDnsServers",
			},
			ComputedFields: []string{
				"properties.provisioningState",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.forwardingRuleState", Value: "Enabled"}, // enabled default true
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDnsForwardingRulesetForwardingRule()) }
