package recoveryservices

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SiteRecoveryNetworkMapping provides resource knowledge for
// Microsoft.RecoveryServices/vaults/replicationFabrics/replicationNetworks/replicationNetworkMappings.
//
// Two AzureRM TF resources map to this single ARM type and are merged here
// (union of universal knowledge only):
//   - azurerm_site_recovery_network_mapping         (A2A, AzureToAzureCreateNetworkMappingInput)
//   - azurerm_site_recovery_hyperv_network_mapping  (VmmToAzure, fabric resolved by friendly name)
//
// Both create via replicationnetworkmappings.NewReplicationNetworkMappingID and
// CreateNetworkMappingInputProperties{RecoveryNetworkId, RecoveryFabricName,
// FabricSpecificDetails}. recoveryNetworkId is required for both kinds;
// fabricSpecificDetails is a discriminated union whose inner shape differs per
// kind, so only its presence is asserted.
//
// Sources:
//   - AzureRM internal/services/recoveryservices/site_recovery_network_mapping_resource.go
//     :36-40   (timeouts create/delete 30m, read 5m)
//     :42-83   (schema: name StringIsNotEmpty, source/target fabric names, source/target network ids)
//     :122-130 (expand: properties.recoveryNetworkId, recoveryFabricName,
//     fabricSpecificDetails AzureToAzureCreateNetworkMappingInput{primaryNetworkId})
//   - AzureRM internal/services/recoveryservices/site_recovery_hyperv_network_mapping_resource.go
//     :89-95   (ResourceType + IDValidationFunc replicationnetworkmappings.ValidateReplicationNetworkMappingID)
//     :97-160  (create: same NewReplicationNetworkMappingID; recovery fabric "Microsoft Azure")
//   - go-azure-sdk resource-manager/recoveryservicessiterecovery/2024-04-01/replicationnetworkmappings:
//     id_replicationnetworkmapping.go
//     (Segments: .../replicationFabrics/{fabric}/replicationNetworks/{network}/replicationNetworkMappings/{name}),
//     model_createnetworkmappinginputproperties.go
//     (recoveryNetworkId required, fabricSpecificDetails discriminated union)
type SiteRecoveryNetworkMapping struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SiteRecoveryNetworkMapping)(nil)

// NewSiteRecoveryNetworkMapping returns knowledge for the
// replicationNetworkMappings resource.
func NewSiteRecoveryNetworkMapping() *SiteRecoveryNetworkMapping {
	return &SiteRecoveryNetworkMapping{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.RecoveryServices/vaults/replicationFabrics/replicationNetworks/replicationNetworkMappings",
			ApiVersions:  []string{"2024-04-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// recoveryNetworkId (target network) is required for both kinds;
			// fabricSpecificDetails carries the per-kind source-network payload.
			RequiredFields: []string{
				"properties.recoveryNetworkId",
				"properties.fabricSpecificDetails",
			},
		},
	}
}

func init() { azwise.Register(NewSiteRecoveryNetworkMapping()) }
