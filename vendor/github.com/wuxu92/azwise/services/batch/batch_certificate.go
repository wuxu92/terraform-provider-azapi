package batch

import (
	"time"

	"github.com/wuxu92/azwise"
)

// BatchCertificate provides resource knowledge for
// Microsoft.Batch/batchAccounts/certificates.
//
// Contributing Terraform resource: azurerm_batch_certificate.
//
// NOTE: AzureRM marks azurerm_batch_certificate deprecated because the Azure Batch
// Certificates feature was retired 2024-02-29; the ARM resource type still exists in
// the 2024-07-01 management API, so knowledge is emitted for AzAPI users who target it.
//
// Sources:
//   - terraform-provider-azurerm internal/services/batch/batch_certificate_resource.go
//     (schema L45-103, Create body L138-150)
//   - go-azure-sdk resource-manager/batch/2024-07-01/certificate:
//     model_certificatecreateorupdateproperties.go (Create body: data, format, password,
//     thumbprint, thumbprintAlgorithm), model_certificateproperties.go (GET-only fields),
//     constants.go (CertificateFormat).
//
// Notes:
//   - The Terraform "name" attribute is Computed (AzureRM derives it as
//     "<thumbprintAlgorithm>-<thumbprint>"); it is the resource name, not a body field.
//   - certificate → properties.data is Sensitive (Required). thumbprint / thumbprint_algorithm
//     are Required+ForceNew and map to body properties. account_name is envelope-owned
//     (Required+ForceNew) and not a body field.
//   - format cannot be "Cer" together with a password — a cross-field rule enforced in
//     AzureRM's validateBatchCertificateFormatAndPassword; it cannot be expressed as a
//     declarative StringRule/RelationalRule (value-conditional), so it is left as a note.
type BatchCertificate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*BatchCertificate)(nil)

// NewBatchCertificate returns knowledge for the batchAccounts/certificates resource.
func NewBatchCertificate() *BatchCertificate {
	return &BatchCertificate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Batch/batchAccounts/certificates",
			ApiVersions:  []string{"2024-07-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// thumbprint & thumbprint_algorithm are ForceNew and map to body properties.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.thumbprint"},
				{PropertyPath: "properties.thumbprintAlgorithm"},
			},
			// certificate (Required) → properties.data. thumbprint / thumbprint_algorithm
			// (Required) also map to body properties.
			RequiredFields: []string{
				"properties.data",
				"properties.thumbprint",
				"properties.thumbprintAlgorithm",
			},
			StringRules: []azwise.StringRule{
				// ── certificate → properties.data ── validation.StringLenBetween(1, 10000)
				{
					PropertyPath: "properties.data",
					MinLength:    1,
					MaxLength:    10000,
					Message:      "must be between 1 and 10000 characters",
				},
				// ── format → properties.format ── enum Cer/Pfx (full ARM SDK set)
				{
					PropertyPath:  "properties.format",
					AllowedValues: []string{"Cer", "Pfx"},
					Message:       "must be one of Cer or Pfx",
				},
				// ── thumbprint_algorithm → properties.thumbprintAlgorithm ── StringInSlice([SHA1])
				{
					PropertyPath:  "properties.thumbprintAlgorithm",
					AllowedValues: []string{"SHA1"},
					Message:       "must be SHA1",
				},
			},
			// certificate & password are Sensitive in AzureRM schema.
			SensitiveFields: []string{
				"properties.data",
				"properties.password",
			},
			// Read-only status properties returned by GET but absent from the create body.
			ComputedFields: []string{
				"properties.publicData",
				"properties.provisioningState",
				"properties.provisioningStateTransitionTime",
				"properties.previousProvisioningState",
				"properties.previousProvisioningStateTransitionTime",
				"properties.deleteCertificateError",
			},
		},
	}
}

func init() { azwise.Register(NewBatchCertificate()) }
