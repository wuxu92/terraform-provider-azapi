package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CassandraKeyspace provides resource knowledge for
// Microsoft.DocumentDB/databaseAccounts/cassandraKeyspaces
// (azurerm_cosmosdb_cassandra_keyspace).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_cassandra_keyspace_resource.go
//     :41-46  (timeouts: create/update/delete 30m, read 5m)
//     :48-73  (schema: name/account_name ForceNew, throughput CosmosThroughput,
//     autoscale_settings ConflictsWith throughput)
//     :95-112 (create mapping to cosmosdb.CassandraKeyspaceCreateUpdateProperties:
//     options.throughput / options.autoscaleSettings)
//   - go-azure-sdk resource-manager/cosmosdb/2024-08-15/cosmosdb:
//     model_cassandrakeyspacecreateupdateproperties.go, model_createupdateoptions.go,
//     model_autoscalesettings.go (properties.options.* body paths)
//
// Not encoded (deliberate):
//   - throughput (properties.options.throughput) validate.CosmosThroughput additionally
//     requires the value be a multiple of 100; only the minimum (400) is expressible as an
//     IntRule.
//   - autoscale_settings.max_throughput (properties.options.autoscaleSettings.maxThroughput)
//     validate.CosmosMaxThroughput additionally requires a multiple of 1000; only the
//     minimum (1000) is expressible.
type CassandraKeyspace struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CassandraKeyspace)(nil)

// NewCassandraKeyspace returns knowledge for the cassandraKeyspaces resource.
func NewCassandraKeyspace() *CassandraKeyspace {
	return &CassandraKeyspace{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/databaseAccounts/cassandraKeyspaces",
			ApiVersions:  []string{"2024-08-15"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.options.throughput",
					MinValue:     azwise.Ptr(int64(400)),
					Message:      "throughput must be at least 400 (and a multiple of 100)",
				},
				{
					PropertyPath: "properties.options.autoscaleSettings.maxThroughput",
					MinValue:     azwise.Ptr(int64(1000)),
					Message:      "autoscale max_throughput must be at least 1000 (and a multiple of 1000)",
				},
			},
			ConflictsWith: []azwise.RelationalRule{
				{
					Paths: []string{
						"properties.options.autoscaleSettings.maxThroughput",
						"properties.options.throughput",
					},
					Message: "autoscale_settings conflicts with throughput",
				},
			},
		},
	}
}

func init() { azwise.Register(NewCassandraKeyspace()) }
