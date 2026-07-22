package recoveryservices

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SiteRecoveryReplicatedVM provides resource knowledge for
// Microsoft.RecoveryServices/vaults/replicationFabrics/replicationProtectionContainers/replicationProtectedItems.
//
// Two AzureRM TF resources map to this single ARM type and are merged here
// (union of universal knowledge only):
//   - azurerm_site_recovery_replicated_vm         (A2A, A2AEnableProtectionInput)
//   - azurerm_site_recovery_vmware_replicated_vm  (InMageRcm, InMageRcmEnableProtectionInput)
//
// Both create via replicationprotecteditems.NewReplicationProtectedItemID and
// EnableProtectionInputProperties{PolicyId, ProtectableItemId, ProviderSpecificDetails}.
// providerSpecificDetails is a large discriminated union keyed by instanceType,
// with per-kind shapes that share almost no ARM property paths.
//
// Not encoded (deliberate):
//   - target_disk_type / target_replica_disk_type map to DiskAccountType enum
//     fields nested inside the managed_disk / unmanaged_disk ARRAYS
//     (providerSpecificDetails.vmManagedDisks[*].* etc.). azwise cannot lower or
//     resolve a rule through an array element, so these enum rules are skipped.
//   - the bulk of the per-kind provider-specific body (network_interface,
//     target_* references, license/appliance settings) lives under the
//     discriminated providerSpecificDetails union and is not a stable shared path.
//
// Timeouts: the A2A resource uses create 180m / update 80m / delete 80m; the
// VMware resource uses create 120m / update 90m / delete 90m. The larger of each
// is used below as the conservative recommendation.
//
// Sources:
//   - AzureRM internal/services/recoveryservices/site_recovery_replicated_vm_resource.go
//     :56-61   (timeouts A2A create 180m, read 5m, update/delete 80m)
//     :63-298  (schema: name StringIsNotEmpty, policy/container refs, managed_disk/unmanaged_disk
//     target_disk_type enum PossibleValuesForDiskAccountType)
//     :492-495 (expand: EnableProtectionInputProperties.PolicyId + A2AEnableProtectionInput)
//   - AzureRM internal/services/recoveryservices/site_recovery_vmware_replicated_vm_resource.go
//     :82-88   (ResourceType + IDValidationFunc replicationprotecteditems.ValidateReplicationProtectedItemID)
//     :323,553,714,835 (timeouts create 120m, update/delete 90m, read 5m)
//   - go-azure-sdk resource-manager/recoveryservicessiterecovery/2024-04-01/replicationprotecteditems:
//     id_replicationprotecteditem.go
//     (Segments: .../replicationProtectionContainers/{container}/replicationProtectedItems/{name}),
//     model_enableprotectioninputproperties.go (policyId, providerSpecificDetails required),
//     constants.go (DiskAccountType enum)
type SiteRecoveryReplicatedVM struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SiteRecoveryReplicatedVM)(nil)

// NewSiteRecoveryReplicatedVM returns knowledge for the replicationProtectedItems resource.
func NewSiteRecoveryReplicatedVM() *SiteRecoveryReplicatedVM {
	return &SiteRecoveryReplicatedVM{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.RecoveryServices/vaults/replicationFabrics/replicationProtectionContainers/replicationProtectedItems",
			ApiVersions:  []string{"2024-04-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 180 * time.Minute,
				Read:   5 * time.Minute,
				Update: 90 * time.Minute,
				Delete: 90 * time.Minute,
			},
			// policyId and the provider-specific enable payload are required by
			// EnableProtectionInputProperties for both the A2A and VMware kinds.
			RequiredFields: []string{
				"properties.policyId",
				"properties.providerSpecificDetails",
			},
		},
	}
}

func init() { azwise.Register(NewSiteRecoveryReplicatedVM()) }
