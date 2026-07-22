package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PostgreSQLFirewallRule provides resource knowledge for
// Microsoft.DBforPostgreSQL/serverGroupsv2/firewallRules
// (azurerm_cosmosdb_postgresql_firewall_rule).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_postgresql_firewall_rule_resource.go
//     :42-70   (schema: name FirewallRuleName ForceNew, cluster_id ForceNew,
//     start_ip_address/end_ip_address Required IsIPv4Address)
//     :104-109 (create mapping to firewallrules.FirewallRuleProperties:
//     startIpAddress / endIpAddress)
//     :78 (timeout: create 30m)
//   - AzureRM internal/services/cosmos/validate/firewall_rule_name.go:11-14
//     (name regex)
//   - go-azure-sdk resource-manager/postgresqlhsc/2022-11-08/firewallrules:
//     model_firewallruleproperties.go (properties.* body paths)
type PostgreSQLFirewallRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PostgreSQLFirewallRule)(nil)

// NewPostgreSQLFirewallRule returns knowledge for the serverGroupsv2/firewallRules resource.
func NewPostgreSQLFirewallRule() *PostgreSQLFirewallRule {
	return &PostgreSQLFirewallRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforPostgreSQL/serverGroupsv2/firewallRules",
			ApiVersions:  []string{"2022-11-08"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[a-zA-Z0-9][a-zA-Z0-9_.-]{0,}[a-zA-Z0-9_]$`,
					Message:      "name must consist of letters, digits, underscores, periods and hyphens; first character a letter or digit, last a letter, digit or underscore",
				},
				{
					PropertyPath: "properties.startIpAddress",
					Regex:        `^(\d{1,3}\.){3}\d{1,3}$`,
					Message:      "start_ip_address must be a valid IPv4 address",
				},
				{
					PropertyPath: "properties.endIpAddress",
					Regex:        `^(\d{1,3}\.){3}\d{1,3}$`,
					Message:      "end_ip_address must be a valid IPv4 address",
				},
			},
			RequiredFields: []string{
				"properties.startIpAddress",
				"properties.endIpAddress",
			},
		},
	}
}

func init() { azwise.Register(NewPostgreSQLFirewallRule()) }
