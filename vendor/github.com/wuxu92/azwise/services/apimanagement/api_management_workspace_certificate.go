package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementWorkspaceCertificate provides resource knowledge for
// Microsoft.ApiManagement/service/workspaces/certificates.
//
// Mirrors azurerm_api_management_workspace_certificate. `name` and
// `api_management_workspace_id` are envelope/parent references (ForceNew).
// certificate_data_base64 (-> properties.data) and key_vault_secret_id
// (-> properties.keyVault.secretIdentifier) are ExactlyOneOf. password and
// data are sensitive. expiration/subject/thumbprint are read-only.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_workspace_certificate_resource.go
//     Arguments (lines 54-104); Attributes (lines 106-123, computed); Create
//     body (lines 154-181); Timeout 30m.
//   - Microsoft.ApiManagement/service/workspaces/certificates@2024-05-01
//     certificate.CertificateCreateOrUpdateProperties: data, password,
//     keyVault{secretIdentifier,identityClientId}. Response adds expirationDate,
//     subject, thumbprint (read-only).
type ApiManagementWorkspaceCertificate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementWorkspaceCertificate)(nil)

// NewApiManagementWorkspaceCertificate returns knowledge for the workspace certificate resource.
func NewApiManagementWorkspaceCertificate() *ApiManagementWorkspaceCertificate {
	return &ApiManagementWorkspaceCertificate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/workspaces/certificates",
			ApiVersions:  []string{"2024-05-01"},
			SensitiveFields: []string{
				"properties.data",
				"properties.password",
			},
			// ExactlyOneOf: certificate_data_base64 (data) vs key_vault_secret_id
			// (keyVault.secretIdentifier).
			ExactlyOneOf: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.data", "properties.keyVault.secretIdentifier"},
					Message: "exactly one of certificate_data_base64 or key_vault_secret_id must be set",
				},
			},
			// RequiredWith: password requires certificate_data_base64 (data);
			// user_assigned_identity_client_id requires key_vault_secret_id.
			RequiredWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.password", "properties.data"},
					Message: "password requires certificate_data_base64",
				},
				{
					Paths:   []string{"properties.keyVault.identityClientId", "properties.keyVault.secretIdentifier"},
					Message: "user_assigned_identity_client_id requires key_vault_secret_id",
				},
			},
			ComputedFields: []string{
				"properties.expirationDate",
				"properties.subject",
				"properties.thumbprint",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementWorkspaceCertificate()) }
