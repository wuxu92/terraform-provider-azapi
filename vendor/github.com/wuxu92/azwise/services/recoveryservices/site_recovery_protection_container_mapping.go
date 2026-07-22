package recoveryservices

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SiteRecoveryProtectionContainerMapping provides resource knowledge for
// Microsoft.RecoveryServices/vaults/replicationFabrics/replicationProtectionContainers/replicationProtectionContainerMappings.
//
// Three AzureRM TF resources map to this single ARM type and are merged here
// (union of universal knowledge only):
//   - azurerm_site_recovery_protection_container_mapping   (A2A, with automatic_update block)
//   - azurerm_site_recovery_hyperv_replication_policy_association
//   - azurerm_site_recovery_vmware_replication_policy_association
//
// All three create via replicationprotectioncontainermappings.NewReplicationProtectionContainerMappingID
// and CreateProtectionContainerMappingInputProperties{PolicyId, TargetProtectionContainerId,
// ProviderSpecificInput}. providerSpecificInput is a discriminated union: the A2A
// container-mapping resource sets A2AContainerMappingInput (agentAutoUpdateStatus,
// automationAccountArmId, automationAccountAuthenticationType); the two association
// resources send the empty base impl. The enum rules below fire only when those
// A2A fields are present, so they are safe to union across all kinds.
//
// Sources:
//   - AzureRM internal/services/recoveryservices/site_recovery_protection_container_mapping_resource.go
//     :38-43   (timeouts create/update/delete 30m, read 5m)
//     :45-110  (schema: name, recovery_*_name/id, automatic_update.authentication_type
//     enum PossibleValuesForAutomationAccountAuthenticationType, default SystemAssignedIdentity)
//     :143-159 (expand: properties.policyId, properties.targetProtectionContainerId,
//     providerSpecificInput A2AContainerMappingInput{agentAutoUpdateStatus,
//     automationAccountArmId, automationAccountAuthenticationType})
//   - AzureRM internal/services/recoveryservices/site_recovery_hyperv_replication_policy_association_resource.go
//     :106-114 (NewReplicationProtectionContainerMappingID; PolicyId + TargetProtectionContainerId + base impl)
//   - AzureRM internal/services/recoveryservices/site_recovery_vmware_replication_policy_association_resource.go
//     :104-124 (NewReplicationProtectionContainerMappingID; PolicyId + TargetProtectionContainerId + base impl)
//   - go-azure-sdk resource-manager/recoveryservicessiterecovery/2024-04-01/replicationprotectioncontainermappings:
//     id_replicationprotectioncontainermapping.go
//     (Segments: .../replicationProtectionContainers/{container}/replicationProtectionContainerMappings/{name}),
//     model_createprotectioncontainermappinginputproperties.go, model_a2acontainermappinginput.go,
//     constants.go (AgentAutoUpdateStatus, AutomationAccountAuthenticationType)
type SiteRecoveryProtectionContainerMapping struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SiteRecoveryProtectionContainerMapping)(nil)

// NewSiteRecoveryProtectionContainerMapping returns knowledge for the
// replicationProtectionContainerMappings resource.
func NewSiteRecoveryProtectionContainerMapping() *SiteRecoveryProtectionContainerMapping {
	return &SiteRecoveryProtectionContainerMapping{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.RecoveryServices/vaults/replicationFabrics/replicationProtectionContainers/replicationProtectionContainerMappings",
			ApiVersions:  []string{"2024-04-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.providerSpecificInput.agentAutoUpdateStatus",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "agentAutoUpdateStatus must be Disabled or Enabled",
				},
				{
					PropertyPath:  "properties.providerSpecificInput.automationAccountAuthenticationType",
					AllowedValues: []string{"RunAsAccount", "SystemAssignedIdentity"},
					Message:       "automationAccountAuthenticationType must be RunAsAccount or SystemAssignedIdentity",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// AzureRM defaults automatic_update.authentication_type to
				// SystemAssignedIdentity (RunAsAccount is deprecated).
				{PropertyPath: "properties.providerSpecificInput.automationAccountAuthenticationType", Value: "SystemAssignedIdentity"},
			},
			// policyId and targetProtectionContainerId are set by every contributing
			// resource (A2A mapping + both policy associations).
			RequiredFields: []string{
				"properties.policyId",
				"properties.targetProtectionContainerId",
			},
		},
	}
}

func init() { azwise.Register(NewSiteRecoveryProtectionContainerMapping()) }
