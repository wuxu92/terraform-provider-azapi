package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// MsSqlServerDNSAlias provides resource knowledge for Microsoft.Sql/servers/dnsAliases
// (azurerm_mssql_server_dns_alias).
//
// Sources:
//   - AzureRM internal/services/mssql/mssql_server_dns_alias_resource.go
//     :30-45  (name ForceNew, mssql_server_id ForceNew)
//     :50-53  (dns_record Computed)
//     :65-149 (typed-resource timeouts: create 30m, read 5m, delete 10m)
//   - AzureRM internal/services/mssql/validate/mssql.go:63-69 (DNS alias name regex)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/serverdnsaliases:
//     model_serverdnsaliasproperties.go (azureDnsRecord is read-only)
type MsSqlServerDNSAlias struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MsSqlServerDNSAlias)(nil)

// NewMsSqlServerDNSAlias returns knowledge for the servers/dnsAliases resource.
func NewMsSqlServerDNSAlias() *MsSqlServerDNSAlias {
	return &MsSqlServerDNSAlias{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/servers/dnsAliases",
			ApiVersions:  []string{"2023-08-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 10 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[0-9a-z][-0-9a-z]{0,127}[0-9a-z]$`,
					Message:      "DNS alias name can only contain lowercase letters, numbers and the hyphen; the hyphen may not lead or trail",
				},
			},
			ComputedFields: []string{
				"properties.azureDnsRecord",
			},
		},
	}
}

func init() { azwise.Register(NewMsSqlServerDNSAlias()) }
