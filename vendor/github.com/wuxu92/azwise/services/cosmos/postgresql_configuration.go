package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PostgreSQLConfiguration provides resource knowledge for
// Microsoft.DBforPostgreSQL/serverGroupsv2/configurations.
//
// This ARM configurations type is the target of TWO AzureRM TF resources, which
// differ only in the server-role scope they update (coordinator vs node) via the
// UpdateOnCoordinator / UpdateOnNode operations — both send the identical
// ServerConfiguration body { properties.value }. They are MERGED here; only
// knowledge universal to both bodies is encoded.
//
// Contributing TF resources:
//   - azurerm_cosmosdb_postgresql_coordinator_configuration
//     (cosmosdb_postgresql_coordinator_configuration_resource.go)
//   - azurerm_cosmosdb_postgresql_node_configuration
//     (cosmosdb_postgresql_node_configuration_resource.go)
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_postgresql_coordinator_configuration_resource.go
//     :51-73  (schema: name/cluster_id ForceNew, value Required StringIsNotEmpty)
//     :79-112 (create mapping to configurations.ServerConfiguration:
//     properties.value, UpdateOnCoordinator scope)
//   - AzureRM internal/services/cosmos/cosmosdb_postgresql_node_configuration_resource.go
//     :41-63  (schema: identical name/cluster_id/value)
//     :69-105 (create mapping: properties.value, UpdateOnNode scope)
//   - go-azure-sdk resource-manager/postgresqlhsc/2022-11-08/configurations:
//     model_serverconfigurationproperties.go (properties.value body path)
//
// Not encoded (deliberate):
//   - Both resources share one configurations ARM type keyed by version, so no
//     scope-specific ForceNew/Required/Default is unioned — the two bodies are identical.
type PostgreSQLConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PostgreSQLConfiguration)(nil)

// NewPostgreSQLConfiguration returns knowledge for the serverGroupsv2/configurations resource.
func NewPostgreSQLConfiguration() *PostgreSQLConfiguration {
	return &PostgreSQLConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DBforPostgreSQL/serverGroupsv2/configurations",
			ApiVersions:  []string{"2022-11-08"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "properties.value",
					MinLength:    1,
					Message:      "value must not be empty",
				},
			},
			RequiredFields: []string{"properties.value"},
		},
	}
}

func init() { azwise.Register(NewPostgreSQLConfiguration()) }
