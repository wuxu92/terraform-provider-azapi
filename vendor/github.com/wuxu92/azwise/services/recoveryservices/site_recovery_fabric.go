package recoveryservices

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SiteRecoveryFabric provides resource knowledge for
// Microsoft.RecoveryServices/vaults/replicationFabrics.
//
// Two AzureRM TF resources map to this single ARM type and are merged here
// (union of universal knowledge only):
//   - azurerm_site_recovery_fabric              (Azure fabric, customDetails.instanceType = "Azure")
//   - azurerm_site_recovery_services_vault_hyperv_site (HyperV site fabric, customDetails.instanceType = "HyperVSite")
//
// Both create the resource via replicationfabrics.NewReplicationFabricID and
// FabricCreationInput{Properties.CustomDetails}. Neither exposes an Update, so
// the resource is entirely replace-on-change; name / recovery_vault are envelope
// or parent references. customDetails is a discriminated union whose shape
// differs per kind (AzureFabricCreationInput.location vs. HyperVSite instance),
// so per-kind nested fields are not unioned as declarative rules.
//
// Sources:
//   - AzureRM internal/services/recoveryservices/site_recovery_fabric_resource.go
//     :34-38  (timeouts create/delete 30m, read 5m)
//     :40-56  (schema: name StringIsNotEmpty, recovery_vault_name, location)
//     :87-93  (expand: properties.customDetails = AzureFabricCreationInput{location})
//   - AzureRM internal/services/recoveryservices/site_recovery_services_vault_hyperv_site_resource.go
//     :75-83  (expand: properties.customDetails = BaseFabricSpecificCreationInputImpl instanceType "HyperVSite")
//     :151-153 (IDValidationFunc replicationfabrics.ValidateReplicationFabricID)
//   - go-azure-sdk resource-manager/recoveryservicessiterecovery/2024-04-01/replicationfabrics:
//     id_replicationfabric.go (Segments: .../vaults/{vaultName}/replicationFabrics/{name}),
//     model_fabriccreationinputproperties.go (customDetails required, discriminated union)
type SiteRecoveryFabric struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SiteRecoveryFabric)(nil)

// NewSiteRecoveryFabric returns knowledge for the replicationFabrics resource.
func NewSiteRecoveryFabric() *SiteRecoveryFabric {
	return &SiteRecoveryFabric{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.RecoveryServices/vaults/replicationFabrics",
			ApiVersions:  []string{"2024-04-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// customDetails (the fabric-specific creation payload) is required for
			// both the Azure and HyperVSite kinds. Its inner shape is a discriminated
			// union, so only its presence is asserted here.
			RequiredFields: []string{
				"properties.customDetails",
			},
		},
	}
}

func init() { azwise.Register(NewSiteRecoveryFabric()) }
