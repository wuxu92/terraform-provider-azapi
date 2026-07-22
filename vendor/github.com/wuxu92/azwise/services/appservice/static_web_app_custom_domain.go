package appservice

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StaticWebAppCustomDomain provides resource knowledge for
// Microsoft.Web/staticSites/customDomains.
//
// Mirrors azurerm_static_web_app_custom_domain. The ARM resource name is the
// domain name (azurerm domain_name); the parent envelope is the static site
// (azurerm static_web_app_id). The create body carries only the validation
// method; the validation token is returned read-only in the overview response.
//
// Sources:
//   - terraform-provider-azurerm internal/services/appservice/static_web_app_custom_domain_resource.go:26-176
//     (schema: domain_name/static_web_app_id/validation_type all ForceNew+Required,
//     validation_token Computed+Sensitive; create maps validation_type -> validationMethod;
//     timeouts Create 30m / Read 5m / Delete 30m)
//   - terraform-provider-azurerm internal/services/appservice/helpers/enums.go:14-17
//     (ValidationTypeTXT="dns-txt-token", ValidationTypeCName="cname-delegation")
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-01-01/staticsites/model_staticsitecustomdomainrequestpropertiesarmresourceproperties.go:6-8
//     (request model: validationMethod only)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-01-01/staticsites/model_staticsitecustomdomainoverviewarmresourceproperties.go:16-17
//     (overview/read model: status + validationToken read-only)
type StaticWebAppCustomDomain struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StaticWebAppCustomDomain)(nil)

// NewStaticWebAppCustomDomain returns knowledge for the customDomains resource.
func NewStaticWebAppCustomDomain() *StaticWebAppCustomDomain {
	return &StaticWebAppCustomDomain{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Web/staticSites/customDomains",
			ApiVersions:  []string{"2023-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.validationMethod"},
			},
			RequiredFields: []string{
				"properties.validationMethod",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{MinLength: 1, Message: "domain name (resource name) must not be empty"},
				{
					PropertyPath:  "properties.validationMethod",
					AllowedValues: []string{"dns-txt-token", "cname-delegation"},
					Message:       "validation method must be dns-txt-token or cname-delegation",
				},
			},
			SensitiveFields: []string{
				"properties.validationToken",
			},
			ComputedFields: []string{
				"properties.validationToken",
			},
		},
	}
}

func init() { azwise.Register(NewStaticWebAppCustomDomain()) }
