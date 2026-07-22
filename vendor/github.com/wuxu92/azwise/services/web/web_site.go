package web

import (
	"time"

	"github.com/wuxu92/azwise"
)

// WebSite provides resource knowledge for Microsoft.Web/sites.
//
// Sources:
//   - terraform-provider-azurerm internal/services/appservice/linux_web_app_resource.go:80-217
//     (name/location ForceNew, required service_plan_id, common site properties/defaults)
//   - terraform-provider-azurerm internal/services/appservice/linux_web_app_resource.go:281-415,524-738
//     (API imports, create/read/update/delete timeouts, ARM property mapping)
//   - terraform-provider-azurerm internal/services/appservice/windows_web_app_resource.go:79-216,396-430
//     (same common web-site schema/defaults on Windows apps)
//   - terraform-provider-azurerm internal/services/appservice/linux_function_app_resource.go:115-328,547-571
//     (function-app overlap: client certificate, daily memory quota, public network access)
//   - terraform-provider-azurerm internal/services/appservice/windows_function_app_resource.go:106-319,543-564
//     (function-app overlap and same API versions/timeouts)
//   - terraform-provider-azurerm internal/services/appservice/helpers/linux_web_app_schema.go:59-289
//   - terraform-provider-azurerm internal/services/appservice/helpers/windows_web_app_schema.go:63-302
//   - terraform-provider-azurerm internal/services/appservice/helpers/function_app_schema.go:68-356
//   - terraform-provider-azurerm internal/services/appservice/helpers/shared_schema.go:55-120,177-230,278-340,502-530
//   - terraform-provider-azurerm internal/services/appservice/validate/web_app_name.go:11-18
//   - terraform-provider-azurerm internal/services/appservice/helpers/enums.go:6-12
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-helpers/resourcemanager/commonschema/location.go:11-19
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/model_site.go:10-20
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/model_siteproperties.go:12-66
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/model_siteconfig.go:12-85
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/constants.go
//
// Cross-field / relational constraints:
//
//	The relational rule fields (ConflictsWith/RequiredWith/ExactlyOneOf/AtLeastOneOf on
//	BaseKnowledge) are now expressible, but only for object/scalar ARM paths that resolve
//	inside the Microsoft.Web/sites body. Every cross-field constraint AzureRM declares on
//	site/site_config/auth properties was evaluated; each resolves to an array-element, a
//	separate subresource API, a map-key (app setting), or an app-kind-specific stack field,
//	so none map to two object-level scalar body paths and none are encoded here. The
//	constraints considered and why each is skipped:
//	  - application_stack ExactlyOneOf/AtLeastOneOf/RequiredWith/ConflictsWith
//	    (helpers/app_stack.go:84-198,373-428; helpers/function_app_schema.go:1370-1750):
//	    app-kind-specific stack fields (linuxFxVersion/windowsFxVersion + app settings).
//	  - health_check_path <-> health_check_eviction_time_in_min RequiredWith
//	    (helpers/linux_web_app_schema.go:219-228, helpers/windows_web_app_schema.go:223-232,
//	    helpers/function_app_schema.go:280-289): the eviction side is only the app setting
//	    WEBSITE_HEALTHCHECK_MAXPINGFAILURES (map-key; linux_web_app_resource.go:423-424), so
//	    the relation cannot be expressed with two object-level paths.
//	  - function storage_account_name/storage_account_access_key/storage_uses_managed_identity/
//	    storage_key_vault_secret_id ExactlyOneOf/ConflictsWith (linux_function_app_resource.go:140-176,
//	    windows_function_app_resource.go:132-168): these expand to AzureWebJobsStorage* app
//	    settings (map-keys; linux_function_app_resource.go:485-531), not siteConfig scalars.
//	  - auth_settings / auth_settings_v2 ExactlyOneOf/ConflictsWith/AtLeastOneOf
//	    (helpers/shared_schema.go:701-1160, helpers/auth_v2_schema.go:354-2000): AzureRM
//	    manages these via the config/authsettings and config/authsettingsV2 subresources.
//	  - logs.http_logs file_system <-> azure_blob_storage ConflictsWith
//	    (helpers/common_web_app_schema.go:806,850): config/logs subresource.
//	  - sticky_settings app_setting_names/connection_string_names AtLeastOneOf
//	    (helpers/shared_schema.go:1680-1697): config/slotConfigNames subresource.
//	  - ip_restriction / scm_ip_restriction action one-of checks: array-element paths
//	    (ipSecurityRestrictions[*] / scmIpSecurityRestrictions[*]); see below.
//
// Intentionally skipped here:
//   - App-kind-specific stack/runtime fields (linuxFxVersion, windowsFxVersion,
//     application_stack, functions_extension_version, storage-account app settings) because
//     Microsoft.Web/sites is shared by web apps, function apps, container apps integration,
//     and slots; AzAPI users can set raw ARM values that AzureRM deliberately partitions.
//   - Auth, backup, diagnostic log, sticky-setting, publishing-credential policy, and
//     storage-account mount rules that AzureRM manages through separate Web Apps subresource
//     APIs rather than the Microsoft.Web/sites body.
//   - AzureRM semantic map validators such as validate.AppSettings, and cross-field
//     constraints whose paths are array-element or map-key (IP restriction one-of action
//     checks, the health_check app-setting RequiredWith, function-app storage one-of, and
//     the auth/logs/sticky-setting relations above). Relational rules now cover cross-field
//     constraints, but only for object/scalar ARM paths; array-element and map-key relations
//     remain unexpressible.
//   - IP restriction priority numeric ranges inside arrays; IntRule currently does not
//     evaluate [*] array-element paths, though StringRule does for action enums.
//   - AzureRM computed attributes (default_hostname, outbound IPs, hosting environment,
//     custom_domain_verification_id) are not listed as ComputedFields because the 2023-12-01
//     SDK Site/SiteProperties create model includes those JSON paths; stripping them from
//     AzAPI bodies would risk discarding user-provided raw ARM values.
type WebSite struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*WebSite)(nil)

// NewWebSite returns knowledge for the sites resource.
func NewWebSite() *WebSite {
	return &WebSite{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Web/sites",
			ApiVersions:  []string{"2023-01-01", "2023-12-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
			},
			RequiredFields: []string{
				"properties.serverFarmId",
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
					PropertyPath: "properties.siteConfig.minTlsCipherSuite",
					AllowedValues: []string{
						"TLS_AES_128_GCM_SHA256",
						"TLS_AES_256_GCM_SHA384",
						"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256",
						"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
						"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
						"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
						"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256",
						"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
						"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
						"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384",
						"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
						"TLS_RSA_WITH_AES_128_CBC_SHA",
						"TLS_RSA_WITH_AES_128_CBC_SHA256",
						"TLS_RSA_WITH_AES_128_GCM_SHA256",
						"TLS_RSA_WITH_AES_256_CBC_SHA",
						"TLS_RSA_WITH_AES_256_CBC_SHA256",
						"TLS_RSA_WITH_AES_256_GCM_SHA384",
					},
					Message: "must be a valid Web Apps TLS cipher suite",
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
				{PropertyPath: "properties.clientAffinityPartitioningEnabled", Value: false},
				{PropertyPath: "properties.clientCertEnabled", Value: false},
				{PropertyPath: "properties.enabled", Value: true},
				{PropertyPath: "properties.httpsOnly", Value: false},
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				{PropertyPath: "properties.siteConfig.publicNetworkAccess", Value: "Enabled"},
				{PropertyPath: "properties.vnetBackupRestoreEnabled", Value: false},
				{PropertyPath: "properties.vnetImagePullEnabled", Value: false},
				{PropertyPath: "properties.vnetRouteAllEnabled", Value: false},
				{PropertyPath: "properties.siteConfig.acrUseManagedIdentityCreds", Value: false},
				{PropertyPath: "properties.siteConfig.http20Enabled", Value: false},
				{PropertyPath: "properties.siteConfig.ipSecurityRestrictionsDefaultAction", Value: "Allow"},
				{PropertyPath: "properties.siteConfig.scmIpSecurityRestrictionsDefaultAction", Value: "Allow"},
				{PropertyPath: "properties.siteConfig.localMySqlEnabled", Value: false},
				{PropertyPath: "properties.siteConfig.loadBalancing", Value: "LeastRequests"},
				{PropertyPath: "properties.siteConfig.managedPipelineMode", Value: "Integrated"},
				{PropertyPath: "properties.siteConfig.remoteDebuggingEnabled", Value: false},
				{PropertyPath: "properties.siteConfig.webSocketsEnabled", Value: false},
				{PropertyPath: "properties.siteConfig.ftpsState", Value: "Disabled"},
				{PropertyPath: "properties.siteConfig.minTlsVersion", Value: "1.2"},
				{PropertyPath: "properties.siteConfig.scmMinTlsVersion", Value: "1.2"},
				{PropertyPath: "properties.siteConfig.cors.supportCredentials", Value: false},
			},
		},
	}
}

func init() { azwise.Register(NewWebSite()) }
