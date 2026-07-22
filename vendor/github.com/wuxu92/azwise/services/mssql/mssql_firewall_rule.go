package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MsSqlFirewallRule provides resource knowledge for Microsoft.Sql/servers/firewallRules
// (azurerm_mssql_firewall_rule).
//
// Sources:
//   - AzureRM internal/services/mssql/mssql_firewall_rule_resource.go
//     :35-40   (timeouts: create/update/delete 30m, read 5m)
//     :43-72   (name ForceNew; server_id ForceNew; start_ip_address/end_ip_address Required IsIPAddress)
//     :104-109 (create mapping: properties.startIpAddress / endIpAddress)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/firewallrules:
//     model_serverfirewallruleproperties.go
//
// NOTE: start_ip_address/end_ip_address use validation.IsIPAddress, which accepts both
// IPv4 and IPv6; a precise declarative regex would be lossy, so only the Required
// constraint is captured here.
type MsSqlFirewallRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MsSqlFirewallRule)(nil)

// NewMsSqlFirewallRule returns knowledge for the servers/firewallRules resource.
func NewMsSqlFirewallRule() *MsSqlFirewallRule {
	return &MsSqlFirewallRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/servers/firewallRules",
			ApiVersions:  []string{"2023-08-01-preview"},
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

func init() { azwise.Register(NewMsSqlFirewallRule()) }
