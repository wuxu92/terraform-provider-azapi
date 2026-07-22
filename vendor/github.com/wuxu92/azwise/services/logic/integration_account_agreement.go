package logic

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IntegrationAccountAgreement provides resource knowledge for
// Microsoft.Logic/integrationAccounts/agreements.
//
// Contributing Terraform resource: azurerm_logic_app_integration_account_agreement.
//
// Sources:
//   - terraform-provider-azurerm internal/services/logic/logic_app_integration_account_agreement_resource.go
//     (schema L42-138, Create body L164-183)
//   - terraform-provider-azurerm internal/services/logic/validate/
//     integration_account_agreement_name.go, integration_account_partner_name.go,
//     integration_account_partner_business_identity_qualifier.go,
//     integration_account_partner_business_identity_value.go
//   - go-azure-sdk resource-manager/logic/2019-05-01/integrationaccountagreements:
//     model_integrationaccountagreement.go, model_integrationaccountagreementproperties.go,
//     model_businessidentity.go, constants.go (PossibleValuesForAgreementType), id_agreement.go
//
// Notes:
//   - name (ForceNew) and integration_account_name (ForceNew parent-name segment) and
//     resource_group_name are envelope-owned; only the agreement-name regex is emitted
//     (PropertyPath "").
//   - agreement_type maps to properties.agreementType; the full AgreementType SDK enum
//     is emitted (AzureRM restricts to AS2/X12/Edifact).
//   - content is a required opaque JSON blob (AgreementContent) unmarshaled verbatim into
//     properties.content; its interior (AS2/X12/Edifact protocol settings) is freeform and
//     not modeled as declarative rules.
type IntegrationAccountAgreement struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IntegrationAccountAgreement)(nil)

// NewIntegrationAccountAgreement returns knowledge for the agreements child resource.
func NewIntegrationAccountAgreement() *IntegrationAccountAgreement {
	return &IntegrationAccountAgreement{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Logic/integrationAccounts/agreements",
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
					Message:      "agreement name contains only letters, numbers, dots, parentheses and hyphens, up to 80 characters",
				},
				{
					PropertyPath:  "properties.agreementType",
					AllowedValues: []string{"AS2", "Edifact", "NotSpecified", "X12"},
					Message:       "agreement_type must be a valid AgreementType",
				},
				{
					PropertyPath: "properties.guestPartner",
					Regex:        `^[A-Za-z0-9-().]+$`,
					MaxLength:    80,
					Message:      "guest_partner_name contains only letters, numbers, dots, parentheses and hyphens, up to 80 characters",
				},
				{
					PropertyPath: "properties.hostPartner",
					Regex:        `^[A-Za-z0-9-().]+$`,
					MaxLength:    80,
					Message:      "host_partner_name contains only letters, numbers, dots, parentheses and hyphens, up to 80 characters",
				},
				{
					PropertyPath: "properties.guestIdentity.qualifier",
					Regex:        `^[A-Za-z0-9]+$`,
					Message:      "guest_identity qualifier contains only letters and numbers",
				},
				{
					PropertyPath: "properties.guestIdentity.value",
					Regex:        `^[A-Za-z0-9-() ._]+$`,
					MaxLength:    128,
					Message:      "guest_identity value contains only letters, numbers, dots, parentheses, hyphens and underscores, up to 128 characters",
				},
				{
					PropertyPath: "properties.hostIdentity.qualifier",
					Regex:        `^[A-Za-z0-9]+$`,
					Message:      "host_identity qualifier contains only letters and numbers",
				},
				{
					PropertyPath: "properties.hostIdentity.value",
					Regex:        `^[A-Za-z0-9-() ._]+$`,
					MaxLength:    128,
					Message:      "host_identity value contains only letters, numbers, dots, parentheses, hyphens and underscores, up to 128 characters",
				},
			},
			RequiredFields: []string{
				"properties.agreementType",
				"properties.content",
				"properties.guestPartner",
				"properties.hostPartner",
				"properties.guestIdentity.qualifier",
				"properties.guestIdentity.value",
				"properties.hostIdentity.qualifier",
				"properties.hostIdentity.value",
			},
		},
	}
}

func init() { azwise.Register(NewIntegrationAccountAgreement()) }
