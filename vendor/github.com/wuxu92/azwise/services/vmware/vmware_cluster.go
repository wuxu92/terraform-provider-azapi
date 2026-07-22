package vmware

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VmwareCluster provides resource knowledge for Microsoft.AVS/privateClouds/clusters.
//
// Mirrors azurerm_vmware_cluster.
//
// Sources:
//   - internal/services/vmware/vmware_cluster_resource.go
//     (schema 44-96: name ForceNew StringIsNotEmpty; vmware_cloud_id ForceNew (parent
//     ref, not a body path); cluster_node_count IntBetween(3,16); sku_name ForceNew
//     StringInSlice; timeouts Create/Update/Delete 5h Read 5m;
//     create 125-132 → Cluster{Sku, Properties{ClusterSize}}).
//   - go-azure-sdk resource-manager/vmware/2022-05-01/clusters:
//     model_commonclusterproperties.go (clusterSize settable; clusterId/hosts/
//     provisioningState read-only), model_sku.go (Sku.Name free string),
//     id_cluster.go (type segment casing "Microsoft.AVS"/"privateClouds"/"clusters").
type VmwareCluster struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VmwareCluster)(nil)

// NewVmwareCluster returns knowledge for the privateClouds/clusters resource.
func NewVmwareCluster() *VmwareCluster {
	return &VmwareCluster{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AVS/privateClouds/clusters",
			ApiVersions:  []string{"2022-05-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 5 * time.Hour,
				Read:   5 * time.Minute,
				Update: 5 * time.Hour,
				Delete: 5 * time.Hour,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "sku.name"},
			},
			StringRules: []azwise.StringRule{
				{
					// name: StringIsNotEmpty.
					MinLength: 1,
					Message:   "must not be empty",
				},
				{
					// sku_name: AzureRM StringInSlice (SDK Sku.Name is a free string).
					PropertyPath:  "sku.name",
					AllowedValues: []string{"av20", "av36", "av36t", "av36p", "av36pt", "av48", "av48t", "av52", "av52t", "av64"},
				},
			},
			IntRules: []azwise.IntRule{
				{
					// cluster_node_count: IntBetween(3, 16).
					PropertyPath: "properties.clusterSize",
					MinValue:     azwise.Ptr(int64(3)),
					MaxValue:     azwise.Ptr(int64(16)),
				},
			},
			RequiredFields: []string{
				"sku.name",
				"properties.clusterSize",
			},
			ComputedFields: []string{
				"properties.clusterId",
				"properties.hosts",
				"properties.provisioningState",
			},
			// NOTE: vmware_cloud_id is the parent private-cloud reference (ForceNew in
			// AzureRM) and has no ARM body path — it is encoded into the resource ID.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewVmwareCluster()) }
