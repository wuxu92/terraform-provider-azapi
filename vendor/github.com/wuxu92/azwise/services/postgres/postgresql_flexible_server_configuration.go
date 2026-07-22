package postgres

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PostgresqlFlexibleServerConfiguration provides resource knowledge for
// Microsoft.DBforPostgreSQL/flexibleServers/configurations
// (azurerm_postgresql_flexible_server_configuration).
//
// Sources:
//   - AzureRM internal/services/postgres/postgresql_flexible_server_configuration_resource.go
//     :23-63   (schema: name StringIsNotEmpty ForceNew, server_id ForceNew, value Required)
//     :30-35   (timeouts: create/update/delete 30m, read 5m)
//     :81-86   (create/update mapping to configurations.ConfigurationProperties;
//     Source is hardcoded to "user-override")
//   - go-azure-sdk resource-manager/postgresql/2025-08-01/configurations:
//     model_configurationproperties.go (properties.value / properties.source),
//     id_configuration.go:125 (staticConfigurations)
type PostgresqlFlexibleServerConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PostgresqlFlexibleServerConfiguration)(nil)

// NewPostgresqlFlexibleServerConfiguration returns knowledge for the
// flexibleServers/configurations resource.
func NewPostgresqlFlexibleServerConfiguration() *PostgresqlFlexibleServerConfiguration {
	return &PostgresqlFlexibleServerConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforPostgreSQL/flexibleServers/configurations",
			ApiVersions:  []string{"2025-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.value",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.source", Value: "user-override"},
			},
		},
	}
}

func init() { azwise.Register(NewPostgresqlFlexibleServerConfiguration()) }
