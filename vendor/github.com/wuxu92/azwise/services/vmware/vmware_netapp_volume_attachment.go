package vmware

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VmwareNetappVolumeAttachment provides resource knowledge for
// Microsoft.AVS/privateClouds/clusters/dataStores.
//
// Mirrors azurerm_vmware_netapp_volume_attachment, which creates a NetApp-backed
// datastore under an AVS cluster (Microsoft.AVS/privateClouds/clusters/dataStores).
// This is a real ARM resource with a body (not a pure-operation linker).
//
// Sources:
//   - internal/services/vmware/vmware_netapp_volume_attachment_resource.go
//     (schema 29-52: name ForceNew StringIsNotEmpty; netapp_volume_id ForceNew
//     StringIsNotEmpty; vmware_cluster_id ForceNew (parent ref, not a body path);
//     create 98-105 → Datastore{Properties{NetAppVolume.Id}}; Create timeout 30m,
//     Read 5m, Delete 30m).
//   - go-azure-sdk resource-manager/vmware/2022-05-01/datastores:
//     model_datastoreproperties.go (netAppVolume/diskPoolVolume settable;
//     provisioningState/status read-only), model_netappvolume.go (Id required),
//     id_datastore.go (segment casing "Microsoft.AVS"/"privateClouds"/"clusters"/"dataStores").
type VmwareNetappVolumeAttachment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VmwareNetappVolumeAttachment)(nil)

// NewVmwareNetappVolumeAttachment returns knowledge for the
// privateClouds/clusters/dataStores resource.
func NewVmwareNetappVolumeAttachment() *VmwareNetappVolumeAttachment {
	return &VmwareNetappVolumeAttachment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AVS/privateClouds/clusters/dataStores",
			ApiVersions:  []string{"2022-05-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.netAppVolume.id"},
			},
			StringRules: []azwise.StringRule{
				{
					// name: StringIsNotEmpty.
					MinLength: 1,
					Message:   "must not be empty",
				},
				{
					// netapp_volume_id: StringIsNotEmpty.
					PropertyPath: "properties.netAppVolume.id",
					MinLength:    1,
					Message:      "must not be empty",
				},
			},
			RequiredFields: []string{
				"properties.netAppVolume.id",
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.status",
			},
			// NOTE: vmware_cluster_id is the parent cluster reference (ForceNew in
			// AzureRM) and has no ARM body path — it is encoded into the resource ID.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewVmwareNetappVolumeAttachment()) }
