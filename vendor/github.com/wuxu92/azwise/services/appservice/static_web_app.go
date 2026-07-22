package appservice

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StaticWebApp provides resource knowledge for Microsoft.Web/staticSites.
//
// Sources:
//   - terraform-provider-azurerm internal/services/appservice/static_web_app_resource.go:34-54
//     (StaticWebAppResourceModel fields)
//   - terraform-provider-azurerm internal/services/appservice/static_web_app_resource.go:56-162
//     (schema: name ForceNew, sku_tier/sku_size enums+defaults, repository RequiredWith, computed api_key/default_host_name)
//   - terraform-provider-azurerm internal/services/appservice/static_web_app_resource.go:176-281
//     (create: 30m timeout, envelope -> staticsites.StaticSiteARMResource + StaticSite props)
//   - terraform-provider-azurerm internal/services/appservice/static_web_app_resource.go:283-399
//     (read 5m, delete/update 30m timeouts)
//   - terraform-provider-azurerm internal/services/appservice/validate/static_web_app.go:12-23
//     (name regex)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-01-01/staticsites/model_staticsitearmresource.go:10-20
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-01-01/staticsites/model_staticsite.go:6-25
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-01-01/resourceproviders/constants.go:358-372 (SkuName)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-01-01/staticsites/constants.go:173-176 (StagingEnvironmentPolicy)
//
// Cross-field / relational constraints:
//   - repository_url / repository_token / repository_branch mutual RequiredWith
//     (static_web_app_resource.go:119-139): each requires the other two. These map to
//     the body properties properties.repositoryUrl / properties.repositoryToken /
//     properties.branch, so they are encoded as directional RequiredWith rules.
//
// Intentionally skipped here:
//   - app_settings: managed via a separate sub-resource API
//     (CreateOrUpdateStaticSiteAppSettings -> .../config/appsettings), not the
//     staticSites create body.
//   - basic_auth: managed via the basicAuth sub-resource (sdkhacks CreateOrUpdateBasicAuth).
//   - sku_tier / sku_size: these live on the envelope-level sku block
//     (sku.tier / sku.name) rather than under properties; captured as enum StringRules
//     and defaults on those paths.
//   - AzureRM computed attributes api_key and default_host_name are GET-only
//     (absent from the StaticSite create body) but are not listed as ComputedFields
//     because ARM returns them under keys not sent on create; no separate strip needed.
type StaticWebApp struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StaticWebApp)(nil)

// NewStaticWebApp returns knowledge for the staticSites resource.
func NewStaticWebApp() *StaticWebApp {
	return &StaticWebApp{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Web/staticSites",
			ApiVersions:  []string{"2023-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					Regex:     `^[0-9a-zA-Z-]{1,60}$`,
					MinLength: 1,
					MaxLength: 60,
					Message:   "must contain only alphanumeric characters and dashes, up to 60 characters",
				},
				{
					PropertyPath:  "sku.tier",
					AllowedValues: []string{"Free", "Standard"},
					Message:       "must be Free or Standard",
				},
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Free", "Standard"},
					Message:       "must be Free or Standard",
				},
				{
					PropertyPath:  "properties.stagingEnvironmentPolicy",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "must be Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be Enabled or Disabled",
				},
			},
			SensitiveFields: []string{
				"properties.repositoryToken",
			},
			RequiredWith: []azwise.RelationalRule{
				{Paths: []string{"properties.repositoryUrl", "properties.repositoryToken", "properties.branch"}},
				{Paths: []string{"properties.repositoryToken", "properties.repositoryUrl", "properties.branch"}},
				{Paths: []string{"properties.branch", "properties.repositoryUrl", "properties.repositoryToken"}},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.allowConfigFileUpdates", Value: true},
				{PropertyPath: "properties.stagingEnvironmentPolicy", Value: "Enabled"},
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				{PropertyPath: "sku.tier", Value: "Free"},
				{PropertyPath: "sku.name", Value: "Free"},
			},
		},
	}
}

func init() { azwise.Register(NewStaticWebApp()) }
