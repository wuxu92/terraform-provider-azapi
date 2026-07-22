package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationAccount provides resource knowledge for
// Microsoft.Automation/automationAccounts.
//
// Contributing Terraform resource: azurerm_automation_account.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_account_resource.go
//     (schema resourceAutomationAccount() L33-164, expandEncryption() L394-429, Create L166-228)
//   - go-azure-sdk resource-manager/automation/2024-10-23/automationaccount:
//     model_automationaccountcreateorupdateproperties.go (Create body: disableLocalAuth,
//     encryption, publicNetworkAccess, sku) and model_automationaccountproperties.go (GET-only
//     read fields) and constants.go (SkuNameEnum: Basic, Free).
//   - validate/account_name.go AutomationAccount() name regex.
//
// Notes:
//   - local_authentication_enabled (Default true) maps to the INVERTED ARM flag
//     properties.disableLocalAuth (= !local_authentication_enabled). Its ARM default is false.
//   - encryption.key_vault_key_id is a Key Vault nested-item URI that AzureRM splits into
//     properties.encryption.keyVaultProperties.{keyName,keyVersion,keyvaultUri}; that composite
//     validator (keyvault.ValidateNestedItemID) has no single ARM body field and is not
//     expressed here (see azapin customizer for semantic validators).
//   - encryption.user_assigned_identity_id (commonids.ValidateUserAssignedIdentityID) maps to
//     properties.encryption.identity.userAssignedIdentity — a generic AzureResourceID semantic
//     validator that belongs in the azapin customizer, not a declarative StringRule.
type AutomationAccount struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationAccount)(nil)

// NewAutomationAccount returns knowledge for the automationAccounts resource.
func NewAutomationAccount() *AutomationAccount {
	return &AutomationAccount{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts",
			// automationaccount SDK is 2024-10-23; agent-registration sub-calls use 2019-06-01.
			ApiVersions: []string{"2024-10-23", "2019-06-01"},
			// name & resource_group_name are envelope-owned (Required+RequiresReplace).
			// commonschema.Location() is ForceNew.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// sku_name is Required in AzureRM; ARM path properties.sku.name.
			RequiredFields: []string{
				"properties.sku.name",
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.AutomationAccount()
				{
					Regex:   `^[0-9a-zA-Z][-0-9a-zA-Z]{4,48}[0-9a-zA-Z]$`,
					Message: "must be 6-50 characters, start and end with a letter or number, and contain only letters, numbers and dashes",
				},
				// ── sku_name → properties.sku.name ── StringInSlice(PossibleValuesForSkuNameEnum)
				{
					PropertyPath:  "properties.sku.name",
					AllowedValues: []string{"Basic", "Free"},
					Message:       "must be one of Basic or Free",
				},
			},
			// Optional+Computed / server-defaulted flags.
			DefaultValues: []azwise.DefaultValue{
				// public_network_access_enabled Default true.
				{PropertyPath: "properties.publicNetworkAccess", Value: true},
				// local_authentication_enabled Default true → disableLocalAuth false (inverted).
				{PropertyPath: "properties.disableLocalAuth", Value: false},
			},
			// Present in the GET response model but absent from the Create/Update model,
			// so stripped from the PUT body to avoid perpetual diffs.
			ComputedFields: []string{
				"properties.automationHybridServiceUrl",
				"properties.creationTime",
				"properties.lastModifiedBy",
				"properties.lastModifiedTime",
				"properties.privateEndpointConnections",
				"properties.state",
				"properties.description",
			},
		},
	}
}

func init() { azwise.Register(NewAutomationAccount()) }
