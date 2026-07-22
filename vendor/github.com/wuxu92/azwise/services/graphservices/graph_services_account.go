package graphservices

import (
	"time"

	"github.com/wuxu92/azwise"
)

// GraphServicesAccount provides resource knowledge for
// Microsoft.GraphServices/accounts (Microsoft Graph Services account).
//
// Contributing Terraform resource: azurerm_graph_services_account.
//
// Sources:
//   - terraform-provider-azurerm internal/services/graphservices/graph_services_account_resource.go
//     (schema L49-75, create L77-119 setting properties.appId, update L180-216)
//   - go-azure-sdk resource-manager/graphservices/2023-04-13/graphservicesprods:
//     model_accountresource.go, model_accountresourceproperties.go, constants.go.
//
// Notes:
//   - name & resource_group_name are envelope-owned; both name and application_id are
//     ForceNew. location is hardcoded to "global" by the provider.
//   - application_id (validation.IsUUID) maps to properties.appId; the UUID check is a
//     semantic validator ported at the azapin layer.
//   - billing_plan_id is read-only (absent from the create/update body).
type GraphServicesAccount struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*GraphServicesAccount)(nil)

func NewGraphServicesAccount() *GraphServicesAccount {
	return &GraphServicesAccount{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.GraphServices/accounts",
			ApiVersions:  []string{"2023-04-13"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.appId"},
			},
			ComputedFields: []string{
				"properties.billingPlanId",
				"properties.provisioningState",
			},
			// application_id is required for creation.
			RequiredFields: []string{
				"properties.appId",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewGraphServicesAccount()) }
