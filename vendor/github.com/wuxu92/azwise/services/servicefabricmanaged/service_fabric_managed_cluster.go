package servicefabricmanaged

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ServiceFabricManagedCluster provides resource knowledge for
// Microsoft.ServiceFabric/managedClusters.
//
// Mirrors azurerm_service_fabric_managed_cluster. Only the managed-cluster body
// (managedcluster.ManagedCluster -> ManagedClusterProperties + Sku) is modelled
// here. The `node_type` blocks are a SEPARATE ARM child resource
// (Microsoft.ServiceFabric/managedClusters/nodeTypes) provisioned by AzureRM via
// a distinct nodeTypeClient, so their constraints (vm size, ports, vm_instance_count
// SKU minimums, etc.) belong in that sub-resource's knowledge file, NOT here.
//
// Sources:
//   - terraform-provider-azurerm internal/services/servicefabricmanaged/service_fabric_managed_cluster_resource.go:127-231
//     (schema), :245-342 (create -> managedcluster.ManagedCluster), :340/:383/:474/:493
//     timeouts (Create 90m / Read 5m / Update 90m / Delete 90m), :747-858 (expandClusterProperties).
//   - CustomizeDiff (:497-555) enforces SKU/node-count minimums, single-primary,
//     probe-path presence and node-type immutability — all over the node_type
//     sub-resource, so none map to this body.
//   - backup_service_enabled / dns_service_enabled map into the properties.addonFeatures
//     string ARRAY (BackupRestoreService / DnsService); an array of enum values, so no
//     scalar rule is emitted.
//   - Semantic validators routed to the azapin customizer, not declarative rules:
//     subnet_id (commonids.SubnetId -> properties.subnetId, resource-id validator).
//   - go-azure-sdk resource-manager/servicefabricmanagedcluster/2024-04-01/managedcluster
//     model_managedcluster.go / model_managedclusterproperties.go / model_sku.go /
//     id_managedcluster.go / constants.go.
type ServiceFabricManagedCluster struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ServiceFabricManagedCluster)(nil)

// NewServiceFabricManagedCluster returns knowledge for the Service Fabric managed cluster resource.
func NewServiceFabricManagedCluster() *ServiceFabricManagedCluster {
	return &ServiceFabricManagedCluster{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ServiceFabric/managedClusters",
			ApiVersions:  []string{"2024-04-01"},
			ForceNew: []azwise.ForceNewRule{
				// sku (ForceNew) -> sku.name (envelope-level Sku object).
				{PropertyPath: "sku.name"},
				// subnet_id (ForceNew) -> properties.subnetId.
				{PropertyPath: "properties.subnetId"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 90 * time.Minute,
				Delete: 90 * time.Minute,
			},
			// sku (has default), client/gateway ports (Required), and the ARM-required
			// non-pointer adminUserName / dnsName fields.
			RequiredFields: []string{
				"sku.name",
				"properties.clientConnectionPort",
				"properties.httpGatewayConnectionPort",
				"properties.adminUserName",
				"properties.dnsName",
			},
			DefaultValues: []azwise.DefaultValue{
				// sku Default "Basic".
				{PropertyPath: "sku.name", Value: "Basic"},
				// upgrade_wave Default Wave0 -> properties.clusterUpgradeCadence.
				{PropertyPath: "properties.clusterUpgradeCadence", Value: "Wave0"},
			},
			StringRules: []azwise.StringRule{
				// name: StringLenBetween(4,23) + lowercase/number/hyphen pattern.
				{
					MinLength: 4,
					MaxLength: 23,
					Regex:     `^[a-z0-9]+(-*[a-z0-9])*$`,
					Message:   "name must be 4-23 chars of lowercase letters, numbers and hyphens, starting with a letter and ending with a letter or number",
				},
				// sku: full ARM SkuName enum.
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Basic", "Standard"},
					Message:       "sku must be Basic or Standard",
				},
				// upgrade_wave -> properties.clusterUpgradeCadence: full ARM ClusterUpgradeCadence enum.
				{
					PropertyPath:  "properties.clusterUpgradeCadence",
					AllowedValues: []string{"Wave0", "Wave1", "Wave2"},
					Message:       "upgrade_wave must be Wave0, Wave1 or Wave2",
				},
				// username -> properties.adminUserName: StringLenBetween(1,15) + no special chars.
				{
					PropertyPath: "properties.adminUserName",
					MinLength:    1,
					MaxLength:    15,
					Regex:        `^[^\\/"\[\]:|<>+=;,?*$]{1,14}$`,
					Message:      "username must be 1-14 chars and cannot contain special characters",
				},
				// dns_name -> properties.dnsName: lowercase/number/hyphen pattern.
				{
					PropertyPath: "properties.dnsName",
					Regex:        `^[a-z0-9]+(-*[a-z0-9])*$`,
					Message:      "dns_name must be lowercase letters, numbers and hyphens, starting with a letter and ending with a letter or number",
				},
				// password -> properties.adminPassword: StringLenBetween(8,123).
				{
					PropertyPath: "properties.adminPassword",
					MinLength:    8,
					MaxLength:    123,
					Message:      "password must be between 8 and 123 characters",
				},
			},
			IntRules: []azwise.IntRule{
				// client_connection_port: IntBetween(1500,65535) -> properties.clientConnectionPort.
				{
					PropertyPath: "properties.clientConnectionPort",
					MinValue:     azwise.Ptr[int64](1500),
					MaxValue:     azwise.Ptr[int64](65535),
					Message:      "client_connection_port must be between 1500 and 65535",
				},
				// http_gateway_port: IntBetween(1500,65535) -> properties.httpGatewayConnectionPort.
				{
					PropertyPath: "properties.httpGatewayConnectionPort",
					MinValue:     azwise.Ptr[int64](1500),
					MaxValue:     azwise.Ptr[int64](65535),
					Message:      "http_gateway_port must be between 1500 and 65535",
				},
			},
			SensitiveFields: []string{
				"properties.adminPassword",
			},
			// Read-only body properties present in ManagedClusterProperties (GET) but
			// server-computed, not user inputs.
			ComputedFields: []string{
				"properties.clusterId",
				"properties.clusterState",
				"properties.fqdn",
				"properties.ipv4Address",
				"properties.provisioningState",
				"properties.clusterCertificateThumbprints",
			},
		},
	}
}

func init() { azwise.Register(NewServiceFabricManagedCluster()) }
