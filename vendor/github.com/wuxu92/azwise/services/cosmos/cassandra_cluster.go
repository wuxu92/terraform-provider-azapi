package cosmos

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CassandraCluster provides resource knowledge for
// Microsoft.DocumentDB/cassandraClusters (azurerm_cosmosdb_cassandra_cluster).
//
// Sources:
//   - AzureRM internal/services/cosmos/cosmosdb_cassandra_cluster_resource.go
//     :41-46   (timeouts: create/update/delete 30m, read 5m)
//     :48-140  (schema: name/delegated_management_subnet_id/default_admin_password
//     ForceNew+Required, authentication_method enum+default, version enum+default+ForceNew,
//     hours_between_backups/repair_enabled defaults)
//     :173-197 (create mapping to managedcassandras.ClusterResource /
//     ClusterResourceProperties)
//   - go-azure-sdk resource-manager/cosmosdb/2023-04-15/managedcassandras:
//     model_clusterresourceproperties.go (properties.* body paths),
//     constants.go:12-26 (AuthenticationMethod enum)
//
// Not encoded (deliberate):
//   - client_certificate_pems (properties.clientCertificates[*]),
//     external_gossip_certificate_pems (properties.externalGossipCertificates[*]) with
//     per-element validate.IsCert, and external_seed_node_ip_addresses
//     (properties.externalSeedNodes[*]) with per-element IsIPv4Address are array-element
//     constraints; azwise/azapin cannot lower or resolve a rule through an array element.
//   - properties.seedNodes / properties.gossipCertificates / properties.prometheusEndpoint
//     are server-populated, but the SDK reuses ClusterResourceProperties for the
//     create/update body (they are present in the Create model), so they are NOT listed
//     in ComputedFields (StripComputedFields must not discard user input).
type CassandraCluster struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CassandraCluster)(nil)

// NewCassandraCluster returns knowledge for the cassandraClusters resource.
func NewCassandraCluster() *CassandraCluster {
	return &CassandraCluster{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/cassandraClusters",
			ApiVersions:  []string{"2023-04-15"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.delegatedManagementSubnetId"},
				{PropertyPath: "properties.initialCassandraAdminPassword"},
				{PropertyPath: "properties.cassandraVersion"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.authenticationMethod",
					AllowedValues: []string{"Cassandra", "Ldap", "None"},
					Message:       "authentication_method must be one of Cassandra, Ldap, None",
				},
				{
					// AzureRM restricts cassandraVersion via StringInSlice; the SDK field
					// is a free-form *string (no enum type).
					PropertyPath:  "properties.cassandraVersion",
					AllowedValues: []string{"3.11", "4.0", "4.1", "5.0"},
					Message:       "version must be one of 3.11, 4.0, 4.1, 5.0",
				},
			},
			SensitiveFields: []string{"properties.initialCassandraAdminPassword"},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.authenticationMethod", Value: "Cassandra"},
				{PropertyPath: "properties.hoursBetweenBackups", Value: int64(24)},
				{PropertyPath: "properties.repairEnabled", Value: true},
			},
			RequiredFields: []string{
				"properties.delegatedManagementSubnetId",
				"properties.initialCassandraAdminPassword",
			},
		},
	}
}

func init() { azwise.Register(NewCassandraCluster()) }
