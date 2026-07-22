package logic

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IntegrationAccountCertificate provides resource knowledge for
// Microsoft.Logic/integrationAccounts/certificates.
//
// Contributing Terraform resource: azurerm_logic_app_integration_account_certificate.
//
// Sources:
//   - terraform-provider-azurerm internal/services/logic/logic_app_integration_account_certificate_resource.go
//     (schema L44-98, Create body L124-138, expand key_vault_key L205-224)
//   - terraform-provider-azurerm internal/services/logic/validate/integration_account_certificate_name.go
//   - go-azure-sdk resource-manager/logic/2019-05-01/integrationaccountcertificates:
//     model_integrationaccountcertificate.go, model_integrationaccountcertificateproperties.go,
//     model_keyvaultkeyreference.go, id_certificate.go
//
// Notes:
//   - name (ForceNew), integration_account_name (ForceNew parent segment) and
//     resource_group_name are envelope-owned; only the certificate-name regex is emitted.
//   - key_vault_key maps to properties.key: key_name → properties.key.keyName,
//     key_vault_id → properties.key.keyVault (a resource-ID reference, semantic — validated in
//     an azapin customizer, not declaratively), key_version → properties.key.keyVersion.
//   - public_certificate → properties.publicCertificate; metadata → opaque properties.metadata JSON.
//   - AzureRM requires at-least-one-of key_vault_key/public_certificate (config-level
//     AtLeastOneOf); key_name/key_vault_id are required only when the key_vault_key block is
//     present, so neither is emitted as an unconditional RequiredField.
type IntegrationAccountCertificate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IntegrationAccountCertificate)(nil)

// NewIntegrationAccountCertificate returns knowledge for the certificates child resource.
func NewIntegrationAccountCertificate() *IntegrationAccountCertificate {
	return &IntegrationAccountCertificate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Logic/integrationAccounts/certificates",
			ApiVersions:  []string{"2019-05-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[A-Za-z0-9-()._]+$`,
					MaxLength:    80,
					Message:      "certificate name contains only letters, numbers, dots, parentheses, hyphens and underscores, up to 80 characters",
				},
			},
		},
	}
}

func init() { azwise.Register(NewIntegrationAccountCertificate()) }
