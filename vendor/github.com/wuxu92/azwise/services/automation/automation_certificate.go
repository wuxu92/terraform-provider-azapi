package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationCertificate provides resource knowledge for
// Microsoft.Automation/automationAccounts/certificates.
//
// Contributing Terraform resource: azurerm_automation_certificate.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_certificate_resource.go
//     (schema L21-84, Create L86-123)
//   - go-azure-sdk resource-manager/automation/2024-10-23/certificate:
//     model_certificatecreateorupdateproperties.go (Create body: base64Value, description,
//     isExportable, thumbprint).
//
// Notes:
//   - thumbprint is Computed in the Terraform schema but IS present in the ARM Create/Update
//     model (properties.thumbprint), so it is NOT listed in ComputedFields — stripping it would
//     discard a value an AzAPI user legitimately sends.
//   - base64 is Sensitive; when supplied via sensitive_body the StringRule on
//     properties.base64Value does not run (it only checks the plain body).
type AutomationCertificate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationCertificate)(nil)

// NewAutomationCertificate returns knowledge for the automationAccounts/certificates resource.
func NewAutomationCertificate() *AutomationCertificate {
	return &AutomationCertificate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/certificates",
			ApiVersions:  []string{"2024-10-23"},
			// name & automation_account_name are envelope-owned (Required+ForceNew).
			// base64 is ForceNew → properties.base64Value.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.base64Value"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// base64 is Required → properties.base64Value.
			RequiredFields: []string{
				"properties.base64Value",
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validation.StringIsNotEmpty
				{
					MinLength: 1,
					Message:   "must not be empty",
				},
				// ── base64 → properties.base64Value ── validation.StringIsBase64
				{
					PropertyPath: "properties.base64Value",
					Regex:        `^[A-Za-z0-9+/]*={0,2}$`,
					Message:      "must be a valid base64-encoded string",
				},
			},
			// exportable Default false → properties.isExportable.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.isExportable", Value: false},
			},
			SensitiveFields: []string{
				"properties.base64Value", // Sensitive: true in AzureRM schema
			},
		},
	}
}

func init() { azwise.Register(NewAutomationCertificate()) }
