package postgres

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PostgresqlConfiguration provides resource knowledge for
// Microsoft.DBforPostgreSQL/servers/configurations (azurerm_postgresql_configuration).
//
// Deprecated Single Server sub-resource; DISTINCT from
// Microsoft.DBforPostgreSQL/flexibleServers/configurations.
//
// Sources:
//   - AzureRM internal/services/postgres/postgresql_configuration_resource.go
//     :23-63   (schema: name ForceNew, server_name ForceNew, value Required ForceNew)
//     :33-37   (timeouts: create/delete 30m, read 5m)
//     :78-82   (create mapping to configurations.ConfigurationProperties)
//   - go-azure-sdk resource-manager/postgresql/2017-12-01/configurations:
//     model_configurationproperties.go (properties.value),
//     id_configuration.go:123-125 (staticServers/staticConfigurations)
type PostgresqlConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PostgresqlConfiguration)(nil)

// NewPostgresqlConfiguration returns knowledge for the servers/configurations resource.
func NewPostgresqlConfiguration() *PostgresqlConfiguration {
	return &PostgresqlConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforPostgreSQL/servers/configurations",
			ApiVersions:  []string{"2017-12-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.value"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.value",
			},
		},
	}
}

func init() { azwise.Register(NewPostgresqlConfiguration()) }
