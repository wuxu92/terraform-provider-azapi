package logic

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IntegrationAccountSession provides resource knowledge for
// Microsoft.Logic/integrationAccounts/sessions.
//
// Contributing Terraform resource: azurerm_logic_app_integration_account_session.
//
// Sources:
//   - terraform-provider-azurerm internal/services/logic/logic_app_integration_account_session_resource.go
//     (schema L41-63, Create body L89-95)
//   - go-azure-sdk resource-manager/logic/2019-05-01/integrationaccountsessions:
//     model_integrationaccountsession.go, model_integrationaccountsessionproperties.go,
//     id_session.go
//
// Notes:
//   - name (ForceNew, no format validator in AzureRM), integration_account_name
//     (ForceNew parent segment) and resource_group_name are envelope-owned; no name regex.
//   - content is a required opaque JSON blob mapped verbatim into properties.content; its
//     interior is freeform and not modeled as declarative rules.
type IntegrationAccountSession struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IntegrationAccountSession)(nil)

// NewIntegrationAccountSession returns knowledge for the sessions child resource.
func NewIntegrationAccountSession() *IntegrationAccountSession {
	return &IntegrationAccountSession{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Logic/integrationAccounts/sessions",
			ApiVersions:  []string{"2019-05-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.content",
			},
		},
	}
}

func init() { azwise.Register(NewIntegrationAccountSession()) }
