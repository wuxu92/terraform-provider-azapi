package containerapps

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ContainerAppEnvironmentCertificate provides resource knowledge for
// Microsoft.App/managedEnvironments/certificates.
//
// Mirrors azurerm_container_app_environment_certificate.
//
// Sources:
//   - terraform-provider-azurerm internal/services/containerapps/container_app_environment_certificate_resource.go
//     schema (Arguments 65-128, Attributes 130-162) + Create (164-215): name/environment/blob/
//     password/key_vault ForceNew, blob-vs-key_vault ExactlyOneOf, blob RequiredWith password,
//     key vault identity default "System"; timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/containerapps/2025-07-01/certificates
//     CertificateProperties (value/password/certificateKeyVaultProperties + read-only subjectName/
//     issuer/issueDate/expirationDate/thumbprint/publicKeyHash/subjectAlternativeNames/valid/
//     provisioningState/deploymentErrors), CertificateKeyVaultProperties (identity/keyVaultUrl).
type ContainerAppEnvironmentCertificate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ContainerAppEnvironmentCertificate)(nil)

func NewContainerAppEnvironmentCertificate() *ContainerAppEnvironmentCertificate {
	return &ContainerAppEnvironmentCertificate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.App/managedEnvironments/certificates",
			ApiVersions:  []string{"2025-07-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.value"},
				{PropertyPath: "properties.password"},
				{PropertyPath: "properties.certificateKeyVaultProperties.identity"},
				{PropertyPath: "properties.certificateKeyVaultProperties.keyVaultUrl"},
			},
			// certificate_password is Sensitive.
			SensitiveFields: []string{
				"properties.password",
			},
			// Read-only certificate metadata populated by Azure.
			ComputedFields: []string{
				"properties.subjectName",
				"properties.issuer",
				"properties.issueDate",
				"properties.expirationDate",
				"properties.thumbprint",
				"properties.publicKeyHash",
				"properties.subjectAlternativeNames",
				"properties.valid",
				"properties.provisioningState",
				"properties.deploymentErrors",
			},
			// Exactly one of the PFX/PEM blob or the key vault source must be provided;
			// the blob form requires a password.
			ExactlyOneOf: []azwise.RelationalRule{
				{Paths: []string{"properties.value", "properties.certificateKeyVaultProperties"}, Message: "exactly one of certificate_blob_base64 or certificate_key_vault must be set"},
			},
			RequiredWith: []azwise.RelationalRule{
				{Paths: []string{"properties.value", "properties.password"}},
			},
			// certificate_key_vault.identity defaults to "System".
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.certificateKeyVaultProperties.identity", Value: "System"},
			},
		},
	}
}

func init() { azwise.Register(NewContainerAppEnvironmentCertificate()) }
