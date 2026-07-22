package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementCertificate provides resource knowledge for
// Microsoft.ApiManagement/service/certificates.
//
// Mirrors azurerm_api_management_certificate. `name`, `resource_group_name` and
// `api_management_name` are envelope references (ForceNew). The ARM body accepts
// either an inline PFX (`data` + `password`) or a Key Vault reference
// (`key_vault_secret_id` + optional `key_vault_identity_client_id`):
//   - data                         -> properties.data (sensitive, base64 PFX)
//   - password                     -> properties.password (sensitive)
//   - key_vault_secret_id          -> properties.keyVault.secretIdentifier
//   - key_vault_identity_client_id -> properties.keyVault.identityClientId (UUID)
//
// `expiration`, `subject` and `thumbprint` are read-only: present in the GET
// CertificateContractProperties but absent from CertificateCreateOrUpdateProperties,
// so they are ComputedFields (stripped from the PUT body).
//
// The key_vault_identity_client_id IsUUID check is a semantic validator and is
// intentionally not expressed as a declarative StringRule (it belongs in an
// azapin UUID validator on properties.keyVault.identityClientId).
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_certificate_resource.go
//     schema (lines 43-95); Create body (lines 133-160); Timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/apimanagement/2022-08-01/certificate
//     CertificateCreateOrUpdateProperties: data, password, keyVault{secretIdentifier,
//     identityClientId}; CertificateContractProperties (GET): expirationDate,
//     subject, thumbprint (read-only).
type ApiManagementCertificate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementCertificate)(nil)

// NewApiManagementCertificate returns knowledge for the certificates resource.
func NewApiManagementCertificate() *ApiManagementCertificate {
	return &ApiManagementCertificate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/certificates",
			ApiVersions:  []string{"2022-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			SensitiveFields: []string{
				"properties.data",
				"properties.password",
			},
			ComputedFields: []string{
				"properties.expirationDate",
				"properties.subject",
				"properties.thumbprint",
			},
			// AzureRM: data AtLeastOneOf {data, key_vault_secret_id}.
			AtLeastOneOf: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.data", "properties.keyVault.secretIdentifier"},
					Message: "one of `data` or `key_vault_secret_id` must be set",
				},
			},
			// AzureRM: data ConflictsWith {key_vault_secret_id, key_vault_identity_client_id};
			// key_vault_secret_id ConflictsWith {data, password}.
			ConflictsWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.data", "properties.keyVault.secretIdentifier", "properties.keyVault.identityClientId"},
					Message: "`data` conflicts with `key_vault_secret_id` and `key_vault_identity_client_id`",
				},
				{
					Paths:   []string{"properties.keyVault.secretIdentifier", "properties.data", "properties.password"},
					Message: "`key_vault_secret_id` conflicts with `data` and `password`",
				},
			},
			// AzureRM: password RequiredWith data; key_vault_identity_client_id
			// RequiredWith key_vault_secret_id.
			RequiredWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.password", "properties.data"},
					Message: "`password` requires `data`",
				},
				{
					Paths:   []string{"properties.keyVault.identityClientId", "properties.keyVault.secretIdentifier"},
					Message: "`key_vault_identity_client_id` requires `key_vault_secret_id`",
				},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementCertificate()) }
