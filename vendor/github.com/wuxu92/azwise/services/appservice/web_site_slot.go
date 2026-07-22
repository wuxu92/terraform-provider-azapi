package appservice

import (
	"time"

	"github.com/wuxu92/azwise"
)

// WebSiteSlot provides resource knowledge for Microsoft.Web/sites/slots.
//
// Merged from four AzureRM Terraform resources that all map to the same ARM type
// via the shared webapps.Site body model (they differ only in the parent-site
// envelope reference and app-kind-specific stack/runtime fields, which are handled
// as raw ARM values — see "Intentionally skipped" below):
//   - azurerm_windows_web_app_slot
//   - azurerm_windows_function_app_slot
//   - azurerm_linux_web_app_slot       (merged on behalf of AppserviceB)
//   - azurerm_linux_function_app_slot   (merged on behalf of AppserviceB)
//
// This is the slot counterpart of Microsoft.Web/sites (services/web/web_site.go);
// the site body properties, enums, and defaults are identical, but a slot's
// name is the only ForceNew field (location is inherited from the parent site and
// is not a slot argument) and service_plan_id is Optional (a slot inherits the
// parent plan), so there is no required serverFarmId.
//
// Sources:
//   - terraform-provider-azurerm internal/services/appservice/windows_web_app_slot_resource.go:32-238,293-766
//     (WindowsWebAppSlotModel, schema ForceNew/defaults/validators, create/read/update/delete timeouts)
//   - terraform-provider-azurerm internal/services/appservice/windows_function_app_slot_resource.go:36-78
//     (WindowsFunctionAppSlotModel: function-app overlap fields)
//   - terraform-provider-azurerm internal/services/appservice/linux_web_app_slot_resource.go
//     (linux web-app slot: same Site body, linuxFxVersion stack handled as raw ARM)
//   - terraform-provider-azurerm internal/services/appservice/linux_function_app_slot_resource.go
//     (linux function-app slot: same Site body)
//   - terraform-provider-azurerm internal/services/appservice/helpers/shared_schema.go:55-120,177-230,278-340,502-530
//   - terraform-provider-azurerm internal/services/appservice/validate/web_app_name.go:11-18 (name regex)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/model_site.go:10-20
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/model_siteproperties.go:12-66
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/model_siteconfig.go:12-85
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/constants.go:490-493 (ClientCertMode)
//
// Intentionally skipped here (identical rationale to services/web/web_site.go):
//   - app_service_id / function_app_id: the parent-site envelope/ID segment, not a body property.
//   - App-kind-specific stack/runtime fields (linuxFxVersion, windowsFxVersion,
//     application_stack, functions_extension_version, storage-account app settings):
//     Microsoft.Web/sites/slots is shared by web-app and function-app slots; AzAPI
//     users may set raw ARM values that AzureRM deliberately partitions per kind.
//   - Auth, backup, diagnostic log, storage-account mount, and publishing-credential
//     rules that AzureRM manages through separate Web Apps sub-resource APIs.
//   - Cross-field constraints whose paths are array-element or map-key (IP restriction
//     one-of action checks, health_check app-setting RequiredWith, function-app storage
//     one-of, auth/logs/sticky-setting relations): unexpressible with object-path RelationalRules.
//   - AzureRM computed attributes (default_hostname, outbound IPs, hosting environment,
//     custom_domain_verification_id) are not listed as ComputedFields because the
//     2023-12-01 Site/SiteProperties create model includes those JSON paths; stripping
//     them would risk discarding user-provided raw ARM values.
type WebSiteSlot struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*WebSiteSlot)(nil)

// NewWebSiteSlot returns knowledge for the sites/slots resource.
func NewWebSiteSlot() *WebSiteSlot {
	return &WebSiteSlot{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Web/sites/slots",
			ApiVersions:  []string{"2023-12-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
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
					PropertyPath:  "properties.clientCertMode",
					AllowedValues: []string{"Optional", "OptionalInteractiveUser", "Required"},
					Message:       "must be Optional, OptionalInteractiveUser, or Required",
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.siteConfig.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.siteConfig.ipSecurityRestrictionsDefaultAction",
					AllowedValues: []string{"Allow", "Deny"},
					Message:       "must be Allow or Deny",
				},
				{
					PropertyPath:  "properties.siteConfig.scmIpSecurityRestrictionsDefaultAction",
					AllowedValues: []string{"Allow", "Deny"},
					Message:       "must be Allow or Deny",
				},
				{
					PropertyPath:  "properties.siteConfig.ipSecurityRestrictions[*].action",
					AllowedValues: []string{"Allow", "Deny"},
					Message:       "must be Allow or Deny",
				},
				{
					PropertyPath:  "properties.siteConfig.scmIpSecurityRestrictions[*].action",
					AllowedValues: []string{"Allow", "Deny"},
					Message:       "must be Allow or Deny",
				},
				{
					PropertyPath: "properties.siteConfig.loadBalancing",
					AllowedValues: []string{
						"LeastRequests", "LeastResponseTime", "PerSiteRoundRobin",
						"RequestHash", "WeightedRoundRobin", "WeightedTotalTraffic",
					},
					Message: "must be a valid Site load-balancing mode",
				},
				{
					PropertyPath:  "properties.siteConfig.managedPipelineMode",
					AllowedValues: []string{"Classic", "Integrated"},
					Message:       "must be Classic or Integrated",
				},
				{
					PropertyPath:  "properties.siteConfig.ftpsState",
					AllowedValues: []string{"AllAllowed", "Disabled", "FtpsOnly"},
					Message:       "must be AllAllowed, Disabled, or FtpsOnly",
				},
				{
					PropertyPath:  "properties.siteConfig.minTlsVersion",
					AllowedValues: []string{"1.0", "1.1", "1.2", "1.3"},
					Message:       "must be 1.0, 1.1, 1.2, or 1.3",
				},
				{
					PropertyPath:  "properties.siteConfig.scmMinTlsVersion",
					AllowedValues: []string{"1.0", "1.1", "1.2", "1.3"},
					Message:       "must be 1.0, 1.1, 1.2, or 1.3",
				},
				{
					PropertyPath: "properties.siteConfig.connectionStrings[*].type",
					AllowedValues: []string{
						"ApiHub", "Custom", "DocDb", "EventHub", "MySql", "NotificationHub",
						"PostgreSQL", "RedisCache", "SQLAzure", "SQLServer", "ServiceBus",
					},
					Message: "must be a valid Web Apps connection string type",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.dailyMemoryTimeQuota",
					MinValue:     azwise.Ptr(int64(0)),
					Message:      "daily memory time quota must be non-negative",
				},
				{
					PropertyPath: "properties.siteConfig.numberOfWorkers",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(100)),
					Message:      "worker count must be between 1 and 100",
				},
			},
			SensitiveFields: []string{
				"properties.siteConfig.connectionStrings[*].connectionString",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.clientAffinityEnabled", Value: false},
				{PropertyPath: "properties.clientCertEnabled", Value: false},
				{PropertyPath: "properties.clientCertMode", Value: "Required"},
				{PropertyPath: "properties.enabled", Value: true},
				{PropertyPath: "properties.httpsOnly", Value: false},
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				{PropertyPath: "properties.siteConfig.publicNetworkAccess", Value: "Enabled"},
				{PropertyPath: "properties.vnetBackupRestoreEnabled", Value: false},
				{PropertyPath: "properties.vnetImagePullEnabled", Value: false},
			},
		},
	}
}

func init() { azwise.Register(NewWebSiteSlot()) }
