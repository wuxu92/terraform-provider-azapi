package postgres

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PostgresqlFirewallRule provides resource knowledge for
// Microsoft.DBforPostgreSQL/servers/firewallRules (azurerm_postgresql_firewall_rule).
//
// Deprecated Single Server sub-resource; DISTINCT from
// Microsoft.DBforPostgreSQL/flexibleServers/firewallRules.
//
// Sources:
//   - AzureRM internal/services/postgres/postgresql_firewall_rule_resource.go
//     :24-76   (schema: name StringMatch regex len 1..128 ForceNew, server_name ForceNew,
//     start_ip_address / end_ip_address IsIPv4Address Required ForceNew)
//     :34-38   (timeouts: create/delete 30m, read 5m)
//     :100-105 (create mapping to firewallrules.FirewallRuleProperties)
//   - go-azure-sdk resource-manager/postgresql/2017-12-01/firewallrules:
//     model_firewallruleproperties.go (properties.startIpAddress / properties.endIpAddress),
//     id_firewallrule.go:123-125 (staticServers/staticFirewallRules)
//
// Not encoded (deliberate):
//   - start_ip_address / end_ip_address use validation.IsIPv4Address, a semantic IPv4
//     check that belongs in an azapin customizer validator, not a declarative StringRule.
type PostgresqlFirewallRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PostgresqlFirewallRule)(nil)

// NewPostgresqlFirewallRule returns knowledge for the servers/firewallRules resource.
func NewPostgresqlFirewallRule() *PostgresqlFirewallRule {
	return &PostgresqlFirewallRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforPostgreSQL/servers/firewallRules",
			ApiVersions:  []string{"2017-12-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.startIpAddress"},
				{PropertyPath: "properties.endIpAddress"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					MinLength:    1,
					MaxLength:    128,
					Regex:        `^[-a-zA-Z0-9(_)]{1,128}$`,
					Message:      "firewall rule name must be 1-128 characters: letters, numbers, '_', '(', ')' and '-'",
				},
			},
			RequiredFields: []string{
				"properties.startIpAddress",
				"properties.endIpAddress",
			},
		},
	}
}

func init() { azwise.Register(NewPostgresqlFirewallRule()) }
