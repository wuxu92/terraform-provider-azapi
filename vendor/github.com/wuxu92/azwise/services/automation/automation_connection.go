package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationConnection provides resource knowledge for
// Microsoft.Automation/automationAccounts/connections.
//
// Contributing Terraform resources (all map to the same ARM type; each is a
// thin wrapper that hardcodes properties.connectionType.name and populates a
// different set of properties.fieldDefinitionValues map keys):
//   - azurerm_automation_connection                       (type is user-supplied, ForceNew)
//   - azurerm_automation_connection_service_principal     (connectionType "AzureServicePrincipal")
//   - azurerm_automation_connection_classic_certificate   (connectionType "AzureClassicCertificate")
//   - azurerm_automation_connection_certificate           (connectionType "Azure")
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_connection_resource.go
//     (schema L24-80, Create L83-135), automation_connection_service_principal_resource.go (L22-125),
//     automation_connection_classic_certificate_resource.go (L22-116),
//     automation_connection_certificate_resource.go (L22-111)
//   - go-azure-sdk resource-manager/automation/2024-10-23/connection:
//     model_connectioncreateorupdateproperties.go (Create body: connectionType, description,
//     fieldDefinitionValues) and model_connectionproperties.go (GET-only: creationTime,
//     lastModifiedTime).
//   - validate/connection_name.go ConnectionName regex.
//
// Notes:
//   - properties.fieldDefinitionValues is a free-form map[string]string whose keys vary by
//     connection type (ApplicationId/CertificateThumbprint/SubscriptionId/TenantId,
//     SubscriptionName/CertificateAssetName, AutomationCertificateName, ...). The AzureRM
//     validation.IsUUID checks on application_id/subscription_id/tenant_id are map-value
//     semantic validators with no fixed ARM schema path, so they are not expressed as
//     declarative StringRules (they belong in the azapin customizer if enforced).
type AutomationConnection struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationConnection)(nil)

// NewAutomationConnection returns knowledge for the automationAccounts/connections resource.
func NewAutomationConnection() *AutomationConnection {
	return &AutomationConnection{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/connections",
			ApiVersions:  []string{"2024-10-23"},
			// name & automation_account_name are envelope-owned (Required+ForceNew).
			// azurerm_automation_connection.type is ForceNew → properties.connectionType.name;
			// the specialized wrappers hardcode it, so it never changes for them either.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.connectionType.name"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// connectionType is Required (unmarshals non-omitempty in the Create model).
			RequiredFields: []string{
				"properties.connectionType.name",
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.ConnectionName
				{
					Regex:     `^[\w\-]{1,128}$`,
					MinLength: 1,
					MaxLength: 128,
					Message:   "must contain only letters, numbers, hyphens and underscores, 1-128 characters",
				},
			},
			// GET-only server timestamps, absent from the Create/Update model.
			ComputedFields: []string{
				"properties.creationTime",
				"properties.lastModifiedTime",
			},
		},
	}
}

func init() { azwise.Register(NewAutomationConnection()) }
