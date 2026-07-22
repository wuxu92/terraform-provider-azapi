package logic

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IntegrationAccount provides resource knowledge for Microsoft.Logic/integrationAccounts.
//
// Contributing Terraform resource: azurerm_logic_app_integration_account.
//
// Sources:
//   - terraform-provider-azurerm internal/services/logic/logic_app_integration_account_resource.go
//     (schema L43-73, Create body L99-112)
//   - terraform-provider-azurerm internal/services/logic/validate/integration_account_name.go
//     (IntegrationAccountName L13-17)
//   - go-azure-sdk resource-manager/logic/2019-05-01/integrationaccounts:
//     model_integrationaccount.go, model_integrationaccountproperties.go,
//     model_integrationaccountsku.go, constants.go (PossibleValuesForIntegrationAccountSkuName)
//
// Notes:
//   - name/location/resource_group_name are envelope-owned; name ForceNew, emitted only as
//     the resource-name regex (PropertyPath "").
//   - sku_name maps to sku.name; the full IntegrationAccountSkuName SDK enum is emitted
//     (AzureRM restricts to Basic/Free/Standard, but AzAPI sends raw ARM values).
//   - integration_service_environment_id is a ForceNew Azure resource-ID reference
//     (properties.integrationServiceEnvironment.id); its ID-format validation is semantic and
//     belongs in an azapin customizer, not a declarative StringRule.
type IntegrationAccount struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IntegrationAccount)(nil)

// NewIntegrationAccount returns knowledge for the integrationAccounts resource.
func NewIntegrationAccount() *IntegrationAccount {
	return &IntegrationAccount{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Logic/integrationAccounts",
			ApiVersions:  []string{"2019-05-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.integrationServiceEnvironment.id"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[\w-().]{1,80}$`,
					Message:      "Integration name can contain only letters, numbers, '_','-', '(', ')' or '.', up to 80 characters",
				},
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Basic", "Free", "NotSpecified", "Standard"},
					Message:       "sku name must be a valid IntegrationAccountSkuName",
				},
			},
			RequiredFields: []string{
				"sku.name",
			},
		},
	}
}

func init() { azwise.Register(NewIntegrationAccount()) }
