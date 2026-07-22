package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MsSqlVirtualNetworkRule provides resource knowledge for
// Microsoft.Sql/servers/virtualNetworkRules (azurerm_mssql_virtual_network_rule).
//
// Sources:
//   - AzureRM internal/services/mssql/mssql_virtual_network_rule_resource.go
//     :36-41   (timeouts: create/update/delete 30m, read 5m)
//     :43-68   (name ForceNew; server_id ForceNew; subnet_id Required;
//     ignore_missing_vnet_service_endpoint default false)
//     :105-110 (create mapping: properties.virtualNetworkSubnetId / ignoreMissingVnetServiceEndpoint)
//   - AzureRM internal/services/mssql/validate/virtual_network_rule_name.go:11-67
//     (name: 2-64 chars, alnum/underscore/period/hyphen, not start with ._- , not end with .- )
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/virtualnetworkrules:
//     model_virtualnetworkruleproperties.go
type MsSqlVirtualNetworkRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MsSqlVirtualNetworkRule)(nil)

// NewMsSqlVirtualNetworkRule returns knowledge for the servers/virtualNetworkRules resource.
func NewMsSqlVirtualNetworkRule() *MsSqlVirtualNetworkRule {
	return &MsSqlVirtualNetworkRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/servers/virtualNetworkRules",
			ApiVersions:  []string{"2023-08-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					MinLength:    2,
					MaxLength:    64,
					Regex:        `^[A-Za-z0-9][A-Za-z0-9._-]{0,62}[A-Za-z0-9_]$`,
					Message:      "name must be 2-64 characters of alphanumerics, underscores, periods or hyphens; cannot start with a period, underscore or hyphen, and cannot end with a period or hyphen",
				},
			},
			RequiredFields: []string{
				"properties.virtualNetworkSubnetId",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.ignoreMissingVnetServiceEndpoint", Value: false},
			},
		},
	}
}

func init() { azwise.Register(NewMsSqlVirtualNetworkRule()) }
