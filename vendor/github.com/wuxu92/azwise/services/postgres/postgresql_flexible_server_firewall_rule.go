package postgres

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PostgresqlFlexibleServerFirewallRule provides resource knowledge for
// Microsoft.DBforPostgreSQL/flexibleServers/firewallRules
// (azurerm_postgresql_flexible_server_firewall_rule).
//
// Sources:
//   - AzureRM internal/services/postgres/postgresql_flexible_server_firewall_rule_resource.go
//     :26-73   (schema: name FlexibleServerFirewallRuleName len 1..128 regex ForceNew,
//     server_id ForceNew, start_ip_address / end_ip_address IsIPv4Address Required)
//     :33-38   (timeouts: create/update/delete 30m, read 5m)
//     :106-111 (create mapping to firewallrules.FirewallRuleProperties)
//   - go-azure-sdk resource-manager/postgresql/2025-08-01/firewallrules:
//     model_firewallruleproperties.go (properties.startIpAddress / properties.endIpAddress),
//     id_firewallrule.go:125 (staticFirewallRules)
//
// Not encoded (deliberate):
//   - start_ip_address / end_ip_address use validation.IsIPv4Address, a semantic IPv4
//     check that belongs in an azapin customizer validator, not a declarative StringRule.
type PostgresqlFlexibleServerFirewallRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PostgresqlFlexibleServerFirewallRule)(nil)

// NewPostgresqlFlexibleServerFirewallRule returns knowledge for the
// flexibleServers/firewallRules resource.
func NewPostgresqlFlexibleServerFirewallRule() *PostgresqlFlexibleServerFirewallRule {
	return &PostgresqlFlexibleServerFirewallRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforPostgreSQL/flexibleServers/firewallRules",
			ApiVersions:  []string{"2025-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					MinLength:    1,
					MaxLength:    128,
					Regex:        `^[a-zA-Z0-9-_]+$`,
					Message:      "firewall rule name must be 1-128 characters: letters, numbers, '-' and '_'",
				},
			},
			RequiredFields: []string{
				"properties.startIpAddress",
				"properties.endIpAddress",
			},
		},
	}
}

func init() { azwise.Register(NewPostgresqlFlexibleServerFirewallRule()) }
