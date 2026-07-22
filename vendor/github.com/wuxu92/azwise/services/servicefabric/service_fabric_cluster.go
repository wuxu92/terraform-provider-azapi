package servicefabric

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ServiceFabricCluster provides resource knowledge for Microsoft.ServiceFabric/clusters.
//
// Mirrors azurerm_service_fabric_cluster. The cluster body (cluster.Cluster ->
// cluster.ClusterProperties) is used for both create and update.
//
// Sources:
//   - terraform-provider-azurerm internal/services/servicefabric/service_fabric_cluster_resource.go:28-565
//     (schema), :567-679 (createUpdate -> cluster.Cluster), :35-40 (timeouts).
//   - Nested constraints inside the node_type ARRAY (durability_level enum,
//     reverse_proxy_endpoint_port PortNumber) are array-element paths
//     (properties.nodeTypes[*].*) and are skipped per azwise conventions.
//   - Semantic validators routed to the azapin customizer, not declarative rules:
//     azure_active_directory.{tenant_id,cluster_application_id,client_application_id}
//     (validation.IsUUID -> properties.azureActiveDirectory.{tenantId,clusterApplication,
//     clientApplication}, UUID validator); upgrade_policy.*_timeout (serviceFabricValidate.UpgradeTimeout
//     duration format -> properties.upgradeDescription.*).
//   - go-azure-sdk resource-manager/servicefabric/2021-06-01/cluster
//     model_cluster.go / model_clusterproperties.go / model_nodetypedescription.go /
//     model_clusterupgradepolicy.go / model_clusterhealthpolicy.go /
//     model_clusterupgradedeltahealthpolicy.go / id_cluster.go / constants.go.
type ServiceFabricCluster struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ServiceFabricCluster)(nil)

// NewServiceFabricCluster returns knowledge for the Service Fabric cluster resource.
func NewServiceFabricCluster() *ServiceFabricCluster {
	return &ServiceFabricCluster{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ServiceFabric/clusters",
			ApiVersions:  []string{"2021-06-01"},
			ForceNew: []azwise.ForceNewRule{
				// management_endpoint (ForceNew) -> properties.managementEndpoint.
				{PropertyPath: "properties.managementEndpoint"},
				// vm_image (ForceNew) -> properties.vmImage.
				{PropertyPath: "properties.vmImage"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.reliabilityLevel",
				"properties.upgradeMode",
				"properties.managementEndpoint",
				"properties.vmImage",
				"properties.nodeTypes",
			},
			DefaultValues: []azwise.DefaultValue{
				// cluster_code_version Optional+Computed, no explicit default (server decides).
				{PropertyPath: "properties.clusterCodeVersion", Value: nil},
				// upgrade_policy timeout defaults (apply only when upgrade_policy block present).
				{PropertyPath: "properties.upgradeDescription.healthCheckRetryTimeout", Value: "00:45:00"},
				{PropertyPath: "properties.upgradeDescription.healthCheckStableDuration", Value: "00:01:00"},
				{PropertyPath: "properties.upgradeDescription.healthCheckWaitDuration", Value: "00:00:30"},
				{PropertyPath: "properties.upgradeDescription.upgradeDomainTimeout", Value: "02:00:00"},
				{PropertyPath: "properties.upgradeDescription.upgradeReplicaSetCheckTimeout", Value: "10675199.02:48:05.4775807"},
				{PropertyPath: "properties.upgradeDescription.upgradeTimeout", Value: "12:00:00"},
			},
			StringRules: []azwise.StringRule{
				// reliability_level: full ARM ReliabilityLevel enum.
				{
					PropertyPath:  "properties.reliabilityLevel",
					AllowedValues: []string{"None", "Bronze", "Silver", "Gold", "Platinum"},
					Message:       "reliability_level must be None, Bronze, Silver, Gold or Platinum",
				},
				// upgrade_mode: UpgradeMode enum.
				{
					PropertyPath:  "properties.upgradeMode",
					AllowedValues: []string{"Automatic", "Manual"},
					Message:       "upgrade_mode must be Automatic or Manual",
				},
				// service_fabric_zonal_upgrade_mode: SfZonalUpgradeMode enum.
				{
					PropertyPath:  "properties.sfZonalUpgradeMode",
					AllowedValues: []string{"Hierarchical", "Parallel"},
					Message:       "service_fabric_zonal_upgrade_mode must be Hierarchical or Parallel",
				},
				// vmss_zonal_upgrade_mode: VMSSZonalUpgradeMode enum.
				{
					PropertyPath:  "properties.vmssZonalUpgradeMode",
					AllowedValues: []string{"Hierarchical", "Parallel"},
					Message:       "vmss_zonal_upgrade_mode must be Hierarchical or Parallel",
				},
			},
			IntRules: []azwise.IntRule{
				// upgrade_policy.health_policy.max_unhealthy_applications_percent: IntBetween(0,100).
				{
					PropertyPath: "properties.upgradeDescription.healthPolicy.maxPercentUnhealthyApplications",
					MinValue:     azwise.Ptr[int64](0),
					MaxValue:     azwise.Ptr[int64](100),
					Message:      "max_unhealthy_applications_percent must be between 0 and 100",
				},
				// upgrade_policy.health_policy.max_unhealthy_nodes_percent: IntBetween(0,100).
				{
					PropertyPath: "properties.upgradeDescription.healthPolicy.maxPercentUnhealthyNodes",
					MinValue:     azwise.Ptr[int64](0),
					MaxValue:     azwise.Ptr[int64](100),
					Message:      "max_unhealthy_nodes_percent must be between 0 and 100",
				},
				// upgrade_policy.delta_health_policy.max_delta_unhealthy_applications_percent: IntBetween(0,100).
				{
					PropertyPath: "properties.upgradeDescription.deltaHealthPolicy.maxPercentDeltaUnhealthyApplications",
					MinValue:     azwise.Ptr[int64](0),
					MaxValue:     azwise.Ptr[int64](100),
					Message:      "max_delta_unhealthy_applications_percent must be between 0 and 100",
				},
				// upgrade_policy.delta_health_policy.max_delta_unhealthy_nodes_percent: IntBetween(0,100).
				{
					PropertyPath: "properties.upgradeDescription.deltaHealthPolicy.maxPercentDeltaUnhealthyNodes",
					MinValue:     azwise.Ptr[int64](0),
					MaxValue:     azwise.Ptr[int64](100),
					Message:      "max_delta_unhealthy_nodes_percent must be between 0 and 100",
				},
				// upgrade_policy.delta_health_policy.max_upgrade_domain_delta_unhealthy_nodes_percent: IntBetween(0,100).
				{
					PropertyPath: "properties.upgradeDescription.deltaHealthPolicy.maxPercentUpgradeDomainDeltaUnhealthyNodes",
					MinValue:     azwise.Ptr[int64](0),
					MaxValue:     azwise.Ptr[int64](100),
					Message:      "max_upgrade_domain_delta_unhealthy_nodes_percent must be between 0 and 100",
				},
			},
			// AzureRM ConflictsWith on the certificate blocks.
			ConflictsWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.certificate", "properties.certificateCommonNames"},
					Message: "certificate conflicts with certificate_common_names",
				},
				{
					Paths:   []string{"properties.reverseProxyCertificate", "properties.reverseProxyCertificateCommonNames"},
					Message: "reverse_proxy_certificate conflicts with reverse_proxy_certificate_common_names",
				},
			},
			// Read-only body properties present in ClusterProperties (GET) but
			// server-computed, not user inputs.
			ComputedFields: []string{
				"properties.clusterEndpoint",
				"properties.clusterId",
				"properties.clusterState",
				"properties.provisioningState",
				"properties.availableClusterVersions",
			},
		},
	}
}

func init() { azwise.Register(NewServiceFabricCluster()) }
