package web

import (
	"time"

	"github.com/wuxu92/azwise"
)

// WebCertificate provides resource knowledge for Microsoft.Web/certificates.
//
// This ARM type is shared by two legacy AzureRM Terraform resources, merged here
// by unioning only knowledge that is universally valid for every certificate body:
//   - azurerm_app_service_certificate (uploaded PFX / Key Vault backed certificate)
//   - azurerm_app_service_managed_certificate (App Service managed certificate)
//
// Sources:
//   - terraform-provider-azurerm internal/services/web/app_service_certificate_resource.go:27-143
//     (schema: name/location ForceNew, pfx_blob/password/key_vault_id/key_vault_secret_id/
//     app_service_plan_id ForceNew, sensitive fields, ExactlyOneOf(key_vault_secret_id,pfx_blob))
//   - terraform-provider-azurerm internal/services/web/app_service_certificate_resource.go:145-208
//     (create mapping: pfx_blob->properties.pfxBlob, password->properties.password,
//     app_service_plan_id->properties.serverFarmId, key_vault_id->properties.keyVaultId,
//     key_vault_secret_id->properties.keyVaultSecretName)
//   - terraform-provider-azurerm internal/services/web/app_service_managed_certificate_resource.go:29-120,122-201
//     (managed certificate: custom_hostname_binding_id ForceNew -> properties.canonicalName,
//     server_farm_id derived -> properties.serverFarmId; same 30/5/30/30 timeouts)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/certificates/id_certificate.go:103-120
//     (ID casing: Microsoft.Web/certificates)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/certificates/model_certificateproperties.go:12-34
//     (CertificateProperties JSON tags: pfxBlob, password, keyVaultId, keyVaultSecretName,
//     serverFarmId, canonicalName all *string)
//
// Intentionally skipped here:
//   - ExactlyOneOf(key_vault_secret_id, pfx_blob) and ConflictsWith(key_vault_secret_id,
//     [pfx_blob, password]) and RequiredWith(key_vault_id, key_vault_secret_id)
//     (app_service_certificate_resource.go:78,87-88): these are kind-specific to the
//     uploaded/Key Vault certificate. The managed certificate sets neither pfxBlob nor
//     keyVaultSecretName (it sets canonicalName + serverFarmId), so unioning a presence-based
//     relational rule would wrongly fire for managed-certificate bodies. Not encoded.
//   - key_vault_secret_id validation (keyvault.ValidateNestedItemID): the Terraform field is
//     the full data-plane secret URL, but ARM stores only the secret NAME in
//     properties.keyVaultSecretName, so the ID-format validator cannot map to the body path.
//   - pfx_blob validation.StringIsBase64: base64 format check on the sensitive
//     properties.pfxBlob; a declarative regex cannot faithfully express base64 padding, and
//     it applies only to the uploaded-certificate kind. Skipped rather than approximated.
//   - AzureRM computed attributes (friendly_name, subject_name, host_names, issuer,
//     issue_date, expiration_date, thumbprint, hosting_environment_profile_id) are NOT listed
//     as ComputedFields: the 2023-12-01 SDK uses a single CertificateProperties model for both
//     create and read, so those JSON paths are present in the create body and stripping them
//     would discard valid raw ARM input.
//
// ForceNew note: the per-kind body ForceNew paths below (properties.pfxBlob, .password,
// .keyVaultId, .keyVaultSecretName, .serverFarmId, .canonicalName) are each ForceNew for the
// kind that sets them and inert (never present, hence never "changed") for the other kind, so
// they are safe to union onto the shared type.
type WebCertificate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*WebCertificate)(nil)

// NewWebCertificate returns knowledge for the certificates resource.
func NewWebCertificate() *WebCertificate {
	return &WebCertificate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Web/certificates",
			ApiVersions:  []string{"2023-12-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "properties.pfxBlob"},
				{PropertyPath: "properties.password"},
				{PropertyPath: "properties.keyVaultId"},
				{PropertyPath: "properties.keyVaultSecretName"},
				{PropertyPath: "properties.serverFarmId"},
				{PropertyPath: "properties.canonicalName"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					MinLength: 1,
					Message:   "certificate name must not be empty",
				},
				{
					PropertyPath: "properties.keyVaultId",
					Regex:        `(?i)^/subscriptions/[^/]+/resourceGroups/[^/]+/providers/Microsoft\.KeyVault/vaults/[^/]+$`,
					Message:      "must be a Key Vault resource ID",
				},
			},
			SensitiveFields: []string{
				"properties.pfxBlob",
				"properties.password",
			},
		},
	}
}

func init() { azwise.Register(NewWebCertificate()) }
