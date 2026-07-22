package vmware

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VmwarePrivateCloud provides resource knowledge for Microsoft.AVS/privateClouds.
//
// Mirrors azurerm_vmware_private_cloud.
//
// Sources:
//   - internal/services/vmware/vmware_private_cloud_resource.go
//     (schema 44-200: name ForceNew StringIsNotEmpty; sku_name ForceNew StringInSlice;
//     management_cluster.size IntBetween(3,16); network_subnet_cidr ForceNew IsCIDR;
//     internet_connection_enabled Optional Default false; nsxt_password/vcenter_password
//     ForceNew Sensitive; timeouts Create/Update/Delete 10h Read 5m;
//     create 229-244 → PrivateCloud{Sku, Properties{ManagementCluster.ClusterSize,
//     NetworkBlock, Internet, NsxtPassword, VcenterPassword}}).
//   - go-azure-sdk resource-manager/vmware/2022-05-01/privateclouds:
//     model_privatecloudproperties.go (managementCluster/networkBlock/internet/
//     nsxtPassword/vcenterPassword settable; circuit/endpoints/provisioningState/
//     *CertificateThumbprint/*Network read-only),
//     model_commonclusterproperties.go (clusterSize settable; clusterId/hosts read-only),
//     constants.go (InternetEnum Disabled/Enabled),
//     id_privatecloud.go (type segment casing "Microsoft.AVS"/"privateClouds").
type VmwarePrivateCloud struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VmwarePrivateCloud)(nil)

// NewVmwarePrivateCloud returns knowledge for the privateClouds resource.
func NewVmwarePrivateCloud() *VmwarePrivateCloud {
	return &VmwarePrivateCloud{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AVS/privateClouds",
			ApiVersions:  []string{"2022-05-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 10 * time.Hour,
				Read:   5 * time.Minute,
				Update: 10 * time.Hour,
				Delete: 10 * time.Hour,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "sku.name"},
				{PropertyPath: "properties.networkBlock"},
				{PropertyPath: "properties.nsxtPassword"},
				{PropertyPath: "properties.vcenterPassword"},
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
				{
					// internet_connection_enabled maps to properties.internet enum.
					PropertyPath:  "properties.internet",
					AllowedValues: []string{"Disabled", "Enabled"},
				},
			},
			IntRules: []azwise.IntRule{
				{
					// management_cluster.size: IntBetween(3, 16).
					PropertyPath: "properties.managementCluster.clusterSize",
					MinValue:     azwise.Ptr(int64(3)),
					MaxValue:     azwise.Ptr(int64(16)),
				},
			},
			SensitiveFields: []string{
				"properties.nsxtPassword",
				"properties.vcenterPassword",
			},
			RequiredFields: []string{
				"sku.name",
				"properties.networkBlock",
				"properties.managementCluster.clusterSize",
			},
			DefaultValues: []azwise.DefaultValue{
				// internet_connection_enabled Default false → properties.internet "Disabled".
				{PropertyPath: "properties.internet", Value: "Disabled"},
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.circuit",
				"properties.secondaryCircuit",
				"properties.endpoints",
				"properties.managementNetwork",
				"properties.provisioningNetwork",
				"properties.vmotionNetwork",
				"properties.nsxtCertificateThumbprint",
				"properties.vcenterCertificateThumbprint",
			},
			// NOTE: network_subnet_cidr uses validation.IsCIDR (semantic CIDR check),
			// which cannot be expressed as a declarative StringRule; map
			// properties.networkBlock to a CIDR validator in the resource customizer.
			// NOTE: management_cluster.hosts / management_cluster.id map to
			// properties.managementCluster.hosts / .clusterId (read-only array/int
			// element paths) and are not listed as scalar ComputedFields.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewVmwarePrivateCloud()) }
