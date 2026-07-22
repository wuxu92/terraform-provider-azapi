package mysql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MySQLFlexibleServerFirewallRule provides resource knowledge for
// Microsoft.DBforMySQL/flexibleServers/firewallRules.
//
// Contributing Terraform resource: azurerm_mysql_flexible_server_firewall_rule.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mysql/mysql_flexible_server_firewall_rule_resource.go
//     (schema L47-74, Create/Update body L100-105, timeouts L36-41)
//   - go-azure-sdk resource-manager/mysql/2023-12-30/firewallrules:
//     model_firewallrule.go, model_firewallruleproperties.go, id_firewallrule.go
//
// Notes:
//   - name/server_name/resource_group_name are envelope / parent-reference fields; not emitted as body rules.
//   - start_ip_address (properties.startIpAddress) and end_ip_address (properties.endIpAddress) are
//     Required (non-pointer in the SDK model). AzureRM validates them with azValidate.IPv4Address; the
//     IPv4 semantic check is not expressed declaratively here.
type MySQLFlexibleServerFirewallRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MySQLFlexibleServerFirewallRule)(nil)

// NewMySQLFlexibleServerFirewallRule returns knowledge for the flexibleServers/firewallRules resource.
func NewMySQLFlexibleServerFirewallRule() *MySQLFlexibleServerFirewallRule {
	return &MySQLFlexibleServerFirewallRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforMySQL/flexibleServers/firewallRules",
			ApiVersions:  []string{"2023-12-30"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.startIpAddress",
				"properties.endIpAddress",
			},
		},
	}
}

func init() { azwise.Register(NewMySQLFlexibleServerFirewallRule()) }
