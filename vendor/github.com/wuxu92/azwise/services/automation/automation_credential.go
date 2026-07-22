package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationCredential provides resource knowledge for
// Microsoft.Automation/automationAccounts/credentials.
//
// Contributing Terraform resource: azurerm_automation_credential.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_credential_resource.go
//     (schema L21-73, Create L75-114)
//   - go-azure-sdk resource-manager/automation/2024-10-23/credential:
//     model_credentialcreateorupdateproperties.go (Create body: description, password, userName)
//     model_credentialproperties.go (GET-only: creationTime, lastModifiedTime; note password is
//     never returned).
type AutomationCredential struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationCredential)(nil)

// NewAutomationCredential returns knowledge for the automationAccounts/credentials resource.
func NewAutomationCredential() *AutomationCredential {
	return &AutomationCredential{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/credentials",
			ApiVersions:  []string{"2024-10-23"},
			// name & automation_account_name are envelope-owned (Required+ForceNew).
			// No body-path ForceNew.
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// username & password are Required (non-omitempty in the Create model).
			RequiredFields: []string{
				"properties.userName",
				"properties.password",
			},
			SensitiveFields: []string{
				"properties.password", // Sensitive: true in AzureRM schema
			},
			// GET-only server timestamps, absent from the Create/Update model.
			ComputedFields: []string{
				"properties.creationTime",
				"properties.lastModifiedTime",
			},
		},
	}
}

func init() { azwise.Register(NewAutomationCredential()) }
