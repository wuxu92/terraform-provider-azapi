package hdinsight

import (
	"time"

	"github.com/wuxu92/azwise"
)

// HDInsightCluster provides resource knowledge for Microsoft.HDInsight/clusters.
//
// AzureRM models one ARM type (Microsoft.HDInsight/clusters) as five typed
// Terraform resources, one per cluster kind. This file unions only the knowledge
// that is UNIVERSAL across every kind (envelope name validation, timeouts, the
// ForceNew scalar/network fields shared through the common schema, and the enum
// value constraints that hold for all kinds). Cluster-kind-specific component
// versions and role/node definitions differ per kind and are intentionally NOT
// unioned here (they would corrupt validation for the other kinds).
//
// Contributing Terraform resources:
//   - azurerm_hdinsight_hadoop_cluster
//   - azurerm_hdinsight_hbase_cluster
//   - azurerm_hdinsight_interactive_query_cluster
//   - azurerm_hdinsight_kafka_cluster
//   - azurerm_hdinsight_spark_cluster
//
// NOTE: this introduces the services/hdinsight package; the parent must add a
// blank import of it to services/all/all.go for the init() below to run.
//
// Sources:
//   - terraform-provider-azurerm internal/services/hdinsight/schema.go:27-34
//     (SchemaHDInsightName -> validate.HDInsightName, ForceNew),
//     :36-46 (SchemaHDInsightTier -> Standard/Premium enum, Required, ForceNew),
//     :48-59 (SchemaHDInsightTls -> 1.0/1.1/1.2 enum, Optional, ForceNew),
//     :61-69 (SchemaHDInsightClusterVersion -> validate.HDInsightClusterVersion, Required, ForceNew),
//     :245-272 (SchemaHDInsightsNetwork -> connection_direction/private_link_enabled, ForceNew + defaults),
//     :564-585 (ExpandHDInsightsNetwork mapping to networkProperties.*)
//   - terraform-provider-azurerm internal/services/hdinsight/validate/hdinsight.go:11-37
//     (HDInsightClusterVersion + HDInsightName regexes)
//   - terraform-provider-azurerm internal/services/hdinsight/hdinsight_hadoop_cluster_resource.go:72-207
//     (universal schema block: name/location/cluster_version/tier/tls_min_version/component_version/roles),
//     :275-299 (create body: ClusterCreateProperties.Tier/ClusterVersion/MinSupportedTlsVersion/
//     OsType(Linux)/ClusterDefinition{Kind,ComponentVersion}/ComputeProfile{Roles}/NetworkProperties)
//   - go-azure-sdk resource-manager/hdinsight/2021-06-01/clusters:
//     model_clustercreateparametersextended.go (envelope: Location, Properties, Tags, Identity),
//     model_clustercreateproperties.go:6-21 (properties.* body paths),
//     model_clusterdefinition.go:6-11 (clusterDefinition.kind/componentVersion),
//     model_networkproperties.go:6-9 (networkProperties.resourceProviderConnection/privateLink),
//     model_computeprofile.go (computeProfile.roles),
//     constants.go:12-30 (ClusterKind), 250-262 (OSType), 385-397 (PrivateLink),
//     523-535 (ResourceProviderConnection), 564-576 (Tier)
//
// Not encoded (deliberate):
//   - component_version sub-keys (clusterDefinition.componentVersion["hadoop"|"hbase"|...])
//     and the roles/node definitions are per-kind and vary between contributors, so they
//     are not unioned; only the presence of the componentVersion/computeProfile blocks is
//     universal and encoded in RequiredFields.
//   - https_endpoint / ssh_endpoint are Terraform Computed-only endpoints derived from the
//     GET response connectivityEndpoints array; no single scalar Create-body ARM path
//     exists, so they are not listed as ComputedFields.
type HDInsightCluster struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*HDInsightCluster)(nil)

// NewHDInsightCluster returns knowledge for the HDInsight clusters resource.
func NewHDInsightCluster() *HDInsightCluster {
	return &HDInsightCluster{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.HDInsight/clusters",
			ApiVersions:  []string{"2021-06-01"},
			// Identical across all five cluster-kind resources.
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				// Envelope fields (RequiresReplace in every kind's schema).
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				// Universal ForceNew body scalars.
				{PropertyPath: "properties.tier"},
				{PropertyPath: "properties.clusterVersion"},
				{PropertyPath: "properties.minSupportedTlsVersion"},
				// component_version block is ForceNew in every kind (sub-keys are kind-specific).
				{PropertyPath: "properties.clusterDefinition.componentVersion"},
				// network block (shared common schema) is ForceNew in every kind.
				{PropertyPath: "properties.networkProperties.resourceProviderConnection"},
				{PropertyPath: "properties.networkProperties.privateLink"},
			},
			RequiredFields: []string{
				// Required scalars/blocks in every cluster kind's schema.
				"properties.tier",
				"properties.clusterVersion",
				"properties.clusterDefinition.kind",
				"properties.clusterDefinition.componentVersion",
				"properties.computeProfile",
			},
			StringRules: []azwise.StringRule{
				{
					// validate.HDInsightName: 59 chars or fewer, letters/numbers/hyphens,
					// first and last character must be a letter or number.
					PropertyPath: "",
					Regex:        `^[a-zA-Z0-9][a-zA-Z0-9-]{1,57}[a-zA-Z0-9]$`,
					Message:      "cluster name must be 59 characters or fewer, contain only letters, numbers and hyphens, and begin and end with a letter or number",
				},
				{
					PropertyPath:  "properties.tier",
					AllowedValues: []string{"Standard", "Premium"},
					Message:       "tier must be Standard or Premium",
				},
				{
					PropertyPath:  "properties.minSupportedTlsVersion",
					AllowedValues: []string{"1.0", "1.1", "1.2"},
					Message:       "tls_min_version must be 1.0, 1.1 or 1.2",
				},
				{
					// validate.HDInsightClusterVersion: `x.y` or `a.b.c.d`.
					PropertyPath: "properties.clusterVersion",
					Regex:        `^(\d+\.\d+|\d+\.\d+\.\d+\.\d+)$`,
					Message:      "cluster_version must be in the format `x.y` or `a.b.c.d`",
				},
				{
					// Full ARM ClusterKind set; every contributing resource pins one of these.
					PropertyPath:  "properties.clusterDefinition.kind",
					AllowedValues: []string{"HADOOP", "HBASE", "INTERACTIVEHIVE", "KAFKA", "SPARK"},
					Message:       "cluster kind must be one of HADOOP, HBASE, INTERACTIVEHIVE, KAFKA or SPARK",
				},
				{
					// OsType is hardcoded to Linux by AzureRM but a real ARM enum users may set.
					PropertyPath:  "properties.osType",
					AllowedValues: []string{"Linux", "Windows"},
					Message:       "osType must be Linux or Windows",
				},
				{
					PropertyPath:  "properties.networkProperties.resourceProviderConnection",
					AllowedValues: []string{"Inbound", "Outbound"},
					Message:       "network connection_direction must be Inbound or Outbound",
				},
				{
					PropertyPath:  "properties.networkProperties.privateLink",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "network private_link must be Disabled or Enabled",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// network block defaults (fire only when the network block is present).
				{PropertyPath: "properties.networkProperties.resourceProviderConnection", Value: "Inbound"},
				{PropertyPath: "properties.networkProperties.privateLink", Value: "Disabled"},
			},
		},
	}
}

func init() { azwise.Register(NewHDInsightCluster()) }
