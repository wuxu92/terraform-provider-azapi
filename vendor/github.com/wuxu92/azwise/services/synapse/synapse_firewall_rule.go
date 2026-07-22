package synapse

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SynapseFirewallRule provides resource knowledge for
// Microsoft.Synapse/workspaces/firewallRules.
//
// Mirrors azurerm_synapse_firewall_rule.
//
// Sources:
//   - internal/services/synapse/synapse_firewall_rule_resource.go
//     (schema 42-68: name ForceNew (FirewallRuleName validator); start_ip_address +
//     end_ip_address Required (IsIPv4Address); timeouts Create/Update/Delete 30m Read 5m;
//     create 98-103 → IPFirewallRuleProperties startIpAddress/endIpAddress).
//   - internal/services/synapse/validate/firewall_rule_name.go
//     (regex ^[^<>*%&:\\/?]{0,127}[^.<>*%&:\\/?]$, 1-128 chars, no trailing '.').
//   - go-azure-sdk resource-manager (track1) synapse model IPFirewallRuleProperties
//     (startIpAddress/endIpAddress json tags), resourceids.go FirewallRule
//     (segment casing "firewallRules").
//
// Notes:
//   - IsIPv4Address on start_ip_address/end_ip_address is a semantic validator; attach an
//     IPv4 validator in an azapin customizer if desired (not declaratively expressible).
type SynapseFirewallRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SynapseFirewallRule)(nil)

func NewSynapseFirewallRule() *SynapseFirewallRule {
	return &SynapseFirewallRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Synapse/workspaces/firewallRules",
			ApiVersions:  []string{"2021-06-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			StringRules: []azwise.StringRule{
				{
					// validate.FirewallRuleName
					Regex:     `^[^<>*%&:\\/?]{0,127}[^.<>*%&:\\/?]$`,
					MinLength: 1,
					MaxLength: 128,
					Message:   "must be 1-128 chars, may not contain '<>*%&:\\/?' and may not end with '.'",
				},
			},
			RequiredFields: []string{
				"properties.startIpAddress",
				"properties.endIpAddress",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewSynapseFirewallRule()) }
