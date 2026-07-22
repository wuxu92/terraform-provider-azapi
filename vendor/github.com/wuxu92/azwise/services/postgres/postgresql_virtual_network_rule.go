package postgres

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PostgresqlVirtualNetworkRule provides resource knowledge for
// Microsoft.DBforPostgreSQL/servers/virtualNetworkRules
// (azurerm_postgresql_virtual_network_rule).
//
// Deprecated Single Server sub-resource.
//
// Sources:
//   - AzureRM internal/services/postgres/postgresql_virtual_network_rule_resource.go
//     :26-75   (schema: name VirtualNetworkRuleName ForceNew, server_name ForceNew,
//     subnet_id ValidateSubnetID Required, ignore_missing_vnet_service_endpoint Optional)
//     :37-42   (timeouts: create/update/delete 30m, read 5m)
//     :101-106 (create mapping to virtualnetworkrules.VirtualNetworkRuleProperties)
//   - go-azure-sdk resource-manager/postgresql/2017-12-01/virtualnetworkrules:
//     model_virtualnetworkruleproperties.go (properties.virtualNetworkSubnetId /
//     ignoreMissingVnetServiceEndpoint), id_virtualnetworkrule.go:123-125
//     (staticServers/staticVirtualNetworkRules)
//
// Not encoded (deliberate):
//   - subnet_id uses commonids.ValidateSubnetID, a semantic Azure resource-id validator
//     (validators.AzureResourceID in an azapin customizer), not a declarative StringRule.
type PostgresqlVirtualNetworkRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PostgresqlVirtualNetworkRule)(nil)

// NewPostgresqlVirtualNetworkRule returns knowledge for the
// servers/virtualNetworkRules resource.
func NewPostgresqlVirtualNetworkRule() *PostgresqlVirtualNetworkRule {
	return &PostgresqlVirtualNetworkRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforPostgreSQL/servers/virtualNetworkRules",
			ApiVersions:  []string{"2017-12-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.virtualNetworkSubnetId",
			},
		},
	}
}

func init() { azwise.Register(NewPostgresqlVirtualNetworkRule()) }
