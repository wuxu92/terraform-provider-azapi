package recoveryservices

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SiteRecoveryReplicationPolicy provides resource knowledge for
// Microsoft.RecoveryServices/vaults/replicationPolicies.
//
// Three AzureRM TF resources map to this single ARM type and are merged here
// (union of universal knowledge only):
//   - azurerm_site_recovery_replication_policy         (A2A, A2APolicyCreationInput)
//   - azurerm_site_recovery_hyperv_replication_policy  (HyperVReplicaAzurePolicyInput)
//   - azurerm_site_recovery_vmware_replication_policy  (InMageRcmPolicyCreationInput)
//
// All create via replicationpolicies.NewReplicationPolicyID and
// CreatePolicyInputProperties{ProviderSpecificInput}. providerSpecificInput is a
// discriminated union keyed by instanceType, so each kind writes a disjoint set
// of numeric sub-fields. Only the kind-unique bounds are encoded below — each
// fires solely when its owning sub-object is present, so they are safe to union.
//
// Not encoded (deliberate):
//   - properties.providerSpecificInput.appConsistentFrequencyInMinutes is set by
//     BOTH the A2A kind (0..525600) and the VMware/InMageRcm kind (0..720) with
//     DIFFERENT bounds. A single declarative range would corrupt validation for
//     one of the kinds, so it is intentionally omitted (see rule: never union a
//     non-universal constraint onto a shared ARM type).
//   - properties.providerSpecificInput.replicationInterval (HyperV) is an
//     IntInSlice{30,300} discrete set, which IntRule (min/max only) cannot express.
//   - value-conditional CustomizeDiff (app-consistent frequency must be <= retention,
//     and cannot exceed zero when retention is zero) is a cross-field relation over
//     two numeric fields, not a single-property range — not expressible declaratively.
//
// Sources:
//   - AzureRM internal/services/recoveryservices/site_recovery_replication_policy_resource.go
//     :36-41   (timeouts create/update/delete 30m, read 5m)
//     :58-68   (schema A2A: recovery_point_retention_in_minutes/app_consistent_* IntBetween(0,525600))
//     :105-113 (expand: providerSpecificInput A2APolicyCreationInput{recoveryPointHistory, appConsistentFrequencyInMinutes})
//   - AzureRM internal/services/recoveryservices/site_recovery_hyperv_replication_policy_resource.go
//     :46-60   (schema HyperV: recovery_point_retention_in_hours 0..24, app_consistent_*_hours 0..12, replication_interval {30,300})
//     :108-116 (expand: HyperVReplicaAzurePolicyInput{recoveryPointHistoryDuration, applicationConsistentSnapshotFrequencyInHours, replicationInterval})
//   - AzureRM internal/services/recoveryservices/site_recovery_vmware_replication_policy_resource.go
//     :63-75   (schema VMware: recovery_point_retention_in_minutes 0..21600, app_consistent_* 0..720)
//     :121-130 (expand: InMageRcmPolicyCreationInput{recoveryPointHistoryInMinutes, appConsistentFrequencyInMinutes, crashConsistentFrequencyInMinutes, enableMultiVmSync})
//   - go-azure-sdk resource-manager/recoveryservicessiterecovery/2024-04-01/replicationpolicies:
//     id_replicationpolicy.go (Segments: .../vaults/{vaultName}/replicationPolicies/{name}),
//     model_a2apolicycreationinput.go, model_hypervreplicaazurepolicyinput.go,
//     model_inmagercmpolicycreationinput.go, model_createpolicyinputproperties.go
type SiteRecoveryReplicationPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SiteRecoveryReplicationPolicy)(nil)

// NewSiteRecoveryReplicationPolicy returns knowledge for the replicationPolicies resource.
func NewSiteRecoveryReplicationPolicy() *SiteRecoveryReplicationPolicy {
	return &SiteRecoveryReplicationPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.RecoveryServices/vaults/replicationPolicies",
			ApiVersions:  []string{"2024-04-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			IntRules: []azwise.IntRule{
				// A2A only (A2APolicyCreationInput.recoveryPointHistory).
				{
					PropertyPath: "properties.providerSpecificInput.recoveryPointHistory",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(525600)),
					Message:      "recoveryPointHistory (A2A recovery point retention) must be between 0 and 525600 minutes",
				},
				// HyperV only (HyperVReplicaAzurePolicyInput.recoveryPointHistoryDuration).
				{
					PropertyPath: "properties.providerSpecificInput.recoveryPointHistoryDuration",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(24)),
					Message:      "recoveryPointHistoryDuration (HyperV) must be between 0 and 24 hours",
				},
				// HyperV only (HyperVReplicaAzurePolicyInput.applicationConsistentSnapshotFrequencyInHours).
				{
					PropertyPath: "properties.providerSpecificInput.applicationConsistentSnapshotFrequencyInHours",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(12)),
					Message:      "applicationConsistentSnapshotFrequencyInHours (HyperV) must be between 0 and 12 hours",
				},
				// VMware/InMageRcm only (InMageRcmPolicyCreationInput.recoveryPointHistoryInMinutes).
				{
					PropertyPath: "properties.providerSpecificInput.recoveryPointHistoryInMinutes",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(21600)),
					Message:      "recoveryPointHistoryInMinutes (VMware) must be between 0 and 21600 minutes",
				},
			},
			// providerSpecificInput is required by CreatePolicyInputProperties for
			// every kind.
			RequiredFields: []string{
				"properties.providerSpecificInput",
			},
		},
	}
}

func init() { azwise.Register(NewSiteRecoveryReplicationPolicy()) }
