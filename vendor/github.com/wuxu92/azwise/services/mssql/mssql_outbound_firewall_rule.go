package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MsSqlOutboundFirewallRule provides resource knowledge for
// Microsoft.Sql/servers/outboundFirewallRules (azurerm_mssql_outbound_firewall_rule).
//
// Sources:
//   - AzureRM internal/services/mssql/mssql_outbound_firewall_rule_resource.go
//     :33-37  (timeouts: create/delete 30m, read 5m; no update — create-only)
//     :40-51  (name ForceNew; server_id ForceNew)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/outboundfirewallrules:
//     model_outboundfirewallruleproperties.go (only provisioningState, read-only)
//
// NOTE: the rule's identity is its name (the allowed outbound FQDN); the resource has
// no writable body properties.
type MsSqlOutboundFirewallRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MsSqlOutboundFirewallRule)(nil)

// NewMsSqlOutboundFirewallRule returns knowledge for the servers/outboundFirewallRules resource.
func NewMsSqlOutboundFirewallRule() *MsSqlOutboundFirewallRule {
	return &MsSqlOutboundFirewallRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/servers/outboundFirewallRules",
			ApiVersions:  []string{"2023-08-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewMsSqlOutboundFirewallRule()) }
