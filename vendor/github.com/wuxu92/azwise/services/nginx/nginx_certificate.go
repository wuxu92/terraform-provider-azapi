package nginx

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NginxCertificate provides resource knowledge for Nginx.NginxPlus/nginxDeployments/certificates.
//
// Mirrors azurerm_nginx_certificate.
//
// ARM type casing verified against go-azure-sdk nginxcertificate/id_certificate.go Segments:
// "Nginx.NginxPlus" / "nginxDeployments" / "certificates".
//
// Sources:
//   - terraform-provider-azurerm internal/services/nginx/nginx_certificate_resource.go
//     Arguments (34-74: name StringIsNotEmpty ForceNew; nginx_deployment_id parent ref ForceNew;
//     key_virtual_path/certificate_virtual_path StringIsNotEmpty; key_vault_secret_id KeyVault
//     nested secret id), Create (88-130), timeouts Create 30m / Read 5m / Update 30m / Delete 10m.
//   - go-azure-sdk resource-manager/nginx/2024-11-01-preview/nginxcertificate
//     NginxCertificateProperties (certificateVirtualPath/keyVaultSecretId/keyVirtualPath settable;
//     certificateError/keyVaultSecretCreated/keyVaultSecretVersion/provisioningState/sha1Thumbprint
//     read-only).
type NginxCertificate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NginxCertificate)(nil)

// NewNginxCertificate returns knowledge for the Nginx certificates resource.
func NewNginxCertificate() *NginxCertificate {
	return &NginxCertificate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Nginx.NginxPlus/nginxDeployments/certificates",
			ApiVersions:  []string{"2024-11-01-preview"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 10 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			RequiredFields: []string{
				"properties.keyVirtualPath",
				"properties.certificateVirtualPath",
				"properties.keyVaultSecretId",
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "name", MinLength: 1, Message: "name must not be empty"},
				{PropertyPath: "properties.keyVirtualPath", MinLength: 1, Message: "key_virtual_path must not be empty"},
				{PropertyPath: "properties.certificateVirtualPath", MinLength: 1, Message: "certificate_virtual_path must not be empty"},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.certificateError",
				"properties.keyVaultSecretCreated",
				"properties.keyVaultSecretVersion",
				"properties.sha1Thumbprint",
			},
		},
	}
}

func init() { azwise.Register(NewNginxCertificate()) }
