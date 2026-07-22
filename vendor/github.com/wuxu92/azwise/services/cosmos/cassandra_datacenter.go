package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CassandraDatacenter provides resource knowledge for
// Microsoft.DocumentDB/cassandraClusters/dataCenters
// (azurerm_cosmosdb_cassandra_datacenter).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_cassandra_datacenter_resource.go
//     :39-44   (timeouts: create/update/delete 60m, read 5m)
//     :46-128  (schema: name/cassandra_cluster_id/delegated_management_subnet_id
//     ForceNew, location, node_count IntAtLeast(3)+default, disk_count IntBetween(1,10),
//     disk_sku/sku_name/availability_zones_enabled defaults)
//     :157-182 (create mapping to managedcassandras.DataCenterResource /
//     DataCenterResourceProperties)
//   - go-azure-sdk resource-manager/cosmosdb/2023-04-15/managedcassandras:
//     model_datacenterresourceproperties.go (properties.* body paths)
//
// Not encoded (deliberate):
//   - backup_storage_customer_key_uri (properties.backupStorageCustomerKeyUri) and
//     managed_disk_customer_key_uri (properties.managedDiskCustomerKeyUri) are validated
//     by keyvault.ValidateNestedItemID (a composite versioned Key Vault key URI); this is
//     a semantic validator, not a declarative enum/regex/range, so it is left to a
//     customizer.
//   - seed_node_ip_addresses (properties.seedNodes) is server-populated but present in the
//     shared DataCenterResourceProperties Create model, so it is NOT listed in
//     ComputedFields.
type CassandraDatacenter struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CassandraDatacenter)(nil)

// NewCassandraDatacenter returns knowledge for the dataCenters resource.
func NewCassandraDatacenter() *CassandraDatacenter {
	return &CassandraDatacenter{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/cassandraClusters/dataCenters",
			ApiVersions:  []string{"2023-04-15"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.delegatedSubnetId"},
				{PropertyPath: "properties.dataCenterLocation"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.nodeCount",
					MinValue:     azwise.Ptr(int64(3)),
					Message:      "node_count must be at least 3",
				},
				{
					PropertyPath: "properties.diskCapacity",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(10)),
					Message:      "disk_count must be between 1 and 10",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.diskSku", Value: "P30"},
				{PropertyPath: "properties.nodeCount", Value: int64(3)},
				{PropertyPath: "properties.availabilityZone", Value: true},
				{PropertyPath: "properties.sku", Value: "Standard_E16s_v5"},
			},
			RequiredFields: []string{
				"properties.delegatedSubnetId",
				"properties.dataCenterLocation",
			},
		},
	}
}

func init() { azwise.Register(NewCassandraDatacenter()) }
