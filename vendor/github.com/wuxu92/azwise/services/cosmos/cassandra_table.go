package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CassandraTable provides resource knowledge for
// Microsoft.DocumentDB/databaseAccounts/cassandraKeyspaces/tables
// (azurerm_cosmosdb_cassandra_table).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_cassandra_table_resource.go
//     :36-41   (timeouts: create/update/delete 30m, read 5m)
//     :43-84   (schema: name/cassandra_keyspace_id ForceNew, analytical_storage_ttl
//     ForceNew+IntBetween(-1,MaxInt32)\0, default_ttl IntAtLeast(-1), schema Required,
//     throughput CosmosThroughput, autoscale_settings ConflictsWith throughput)
//     :111-137 (create mapping to cosmosdb.CassandraTableCreateUpdateProperties:
//     resource.defaultTtl / resource.analyticalStorageTtl / resource.schema /
//     options.throughput / options.autoscaleSettings)
//   - AzureRM internal/services/cosmos/common/schema.go:14-77
//     (CassandraTableSchemaPropertySchema: cluster_key.order_by StringInSlice[Asc,Desc])
//   - go-azure-sdk resource-manager/cosmosdb/2024-08-15/cosmosdb:
//     model_cassandratablecreateupdateproperties.go, model_cassandratableresource.go,
//     model_cassandraschema.go, model_createupdateoptions.go, model_autoscalesettings.go
//
// Not encoded (deliberate):
//   - analytical_storage_ttl (properties.resource.analyticalStorageTtl) additionally
//     excludes the value 0 (IntNotInSlice); only the -1..MaxInt32 range is expressible.
//   - throughput / autoscale max_throughput multiple-of constraints are not expressible
//     (see cassandra_keyspace.go).
//   - schema.column[*] / partition_key[*] / cluster_key[*] (and cluster_key.order_by
//     StringInSlice[Asc,Desc]) are array-element constraints; azwise/azapin cannot lower
//     or resolve a rule through an array element.
type CassandraTable struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CassandraTable)(nil)

// NewCassandraTable returns knowledge for the cassandraKeyspaces/tables resource.
func NewCassandraTable() *CassandraTable {
	return &CassandraTable{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/databaseAccounts/cassandraKeyspaces/tables",
			ApiVersions:  []string{"2024-08-15"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.resource.analyticalStorageTtl"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.resource.defaultTtl",
					MinValue:     azwise.Ptr(int64(-1)),
					Message:      "default_ttl must be at least -1",
				},
				{
					PropertyPath: "properties.resource.analyticalStorageTtl",
					MinValue:     azwise.Ptr(int64(-1)),
					MaxValue:     azwise.Ptr(int64(2147483647)),
					Message:      "analytical_storage_ttl must be between -1 and 2147483647 and not 0",
				},
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
			RequiredFields: []string{"properties.resource.schema"},
		},
	}
}

func init() { azwise.Register(NewCassandraTable()) }
