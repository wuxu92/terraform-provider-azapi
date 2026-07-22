package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudCertificate provides resource knowledge for
// Microsoft.AppPlatform/Spring/certificates.
//
// Mirrors azurerm_spring_cloud_certificate. The resource has no Update — every
// configurable field is ForceNew — and the ARM properties are a discriminated
// union (ContentCertificate vs KeyVaultCertificate via properties.type).
//
// Sources:
//   - internal/services/springcloud/spring_cloud_certificate_resource.go:28-95
//     (Timeouts :46-50 Create 30m / Read 5m / Delete 30m — no Update;
//     certificate_content :69-75 ForceNew; exclude_private_key :77-81 ForceNew;
//     key_vault_certificate_id :83-89 ForceNew; AtLeastOneOf content/keyvault)
//   - go-azure-sdk .../appplatform model_contentcertificateproperties.go (Content json "content",
//     type "ContentCertificate"), model_keyvaultcertificateproperties.go (KeyVaultCertName
//     "keyVaultCertName", VaultUri "vaultUri", ExcludePrivateKey "excludePrivateKey",
//     type "KeyVaultCertificate")
type SpringCloudCertificate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudCertificate)(nil)

// NewSpringCloudCertificate returns knowledge for the Spring Cloud certificate resource.
func NewSpringCloudCertificate() *SpringCloudCertificate {
	return &SpringCloudCertificate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/certificates",
			ApiVersions:  []string{"2024-01-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// No Update func in AzureRM: the discriminator and every content/key-vault leaf
			// are replace-only. Each leaf fires only when present, so listing both kinds is safe.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.type"},
				{PropertyPath: "properties.content"},
				{PropertyPath: "properties.keyVaultCertName"},
				{PropertyPath: "properties.vaultUri"},
				{PropertyPath: "properties.excludePrivateKey"},
			},
			// AtLeastOneOf(certificate_content, key_vault_certificate_id) maps to the two
			// discriminated content sources; both live under properties with different type
			// tags, so it cannot be expressed as a single declarative ARM-path rule.
		},
	}
}

func init() { azwise.Register(NewSpringCloudCertificate()) }
