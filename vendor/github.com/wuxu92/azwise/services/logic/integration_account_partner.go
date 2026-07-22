package logic

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IntegrationAccountPartner provides resource knowledge for
// Microsoft.Logic/integrationAccounts/partners.
//
// Contributing Terraform resource: azurerm_logic_app_integration_account_partner.
//
// Sources:
//   - terraform-provider-azurerm internal/services/logic/logic_app_integration_account_partner_resource.go
//     (schema L41-84, Create body L110-119)
//   - terraform-provider-azurerm internal/services/logic/validate/integration_account_partner_name.go
//   - go-azure-sdk resource-manager/logic/2019-05-01/integrationaccountpartners:
//     model_integrationaccountpartner.go, model_integrationaccountpartnerproperties.go,
//     model_partnercontent.go, model_b2bpartnercontent.go, model_businessidentity.go,
//     constants.go (PossibleValuesForPartnerType), id_partner.go
//
// Notes:
//   - name (ForceNew), integration_account_name (ForceNew parent segment) and
//     resource_group_name are envelope-owned; only the partner-name regex is emitted.
//   - business_identity (a required set) maps to
//     properties.content.b2b.businessIdentities[*] (qualifier/value); per array-element policy
//     the per-item qualifier/value regexes are not emitted as declarative rules.
//   - partner_type is hardcoded by AzureRM to B2B → properties.partnerType; the full
//     PartnerType SDK enum is emitted for the value constraint.
type IntegrationAccountPartner struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IntegrationAccountPartner)(nil)

// NewIntegrationAccountPartner returns knowledge for the partners child resource.
func NewIntegrationAccountPartner() *IntegrationAccountPartner {
	return &IntegrationAccountPartner{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Logic/integrationAccounts/partners",
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
					Regex:        `^[A-Za-z0-9-().]+$`,
					MaxLength:    80,
					Message:      "partner name contains only letters, numbers, dots, parentheses and hyphens, up to 80 characters",
				},
				{
					PropertyPath:  "properties.partnerType",
					AllowedValues: []string{"B2B", "NotSpecified"},
					Message:       "partner_type must be a valid PartnerType",
				},
			},
			RequiredFields: []string{
				"properties.partnerType",
				"properties.content",
			},
		},
	}
}

func init() { azwise.Register(NewIntegrationAccountPartner()) }
