package appservice

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StaticWebAppFunctionAppRegistration provides resource knowledge for
// Microsoft.Web/staticSites/userProvidedFunctionApps.
//
// Mirrors azurerm_static_web_app_function_app_registration. The ARM resource
// name is the function app name (derived from function_app_id); the parent
// envelope is the static site (azurerm static_web_app_id). The registration is
// fully immutable — there is no update path. functionAppRegion is populated by
// the provider from the target function app's location.
//
// Sources:
//   - terraform-provider-azurerm internal/services/appservice/static_web_app_function_app_registration_resource.go:24-137
//     (schema: static_web_app_id/function_app_id both ForceNew+Required; create maps
//     function_app_id -> functionAppResourceId and derives functionAppRegion;
//     timeouts Create 30m / Read 5m / Delete 30m)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-01-01/staticsites/model_staticsiteuserprovidedfunctionapparmresourceproperties.go:12-16
//     (model: createdOn read-only, functionAppRegion, functionAppResourceId)
type StaticWebAppFunctionAppRegistration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StaticWebAppFunctionAppRegistration)(nil)

// NewStaticWebAppFunctionAppRegistration returns knowledge for the
// userProvidedFunctionApps resource.
func NewStaticWebAppFunctionAppRegistration() *StaticWebAppFunctionAppRegistration {
	return &StaticWebAppFunctionAppRegistration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Web/staticSites/userProvidedFunctionApps",
			ApiVersions:  []string{"2023-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.functionAppResourceId"},
			},
			RequiredFields: []string{
				"properties.functionAppResourceId",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ComputedFields: []string{
				"properties.createdOn",
			},
		},
	}
}

func init() { azwise.Register(NewStaticWebAppFunctionAppRegistration()) }
