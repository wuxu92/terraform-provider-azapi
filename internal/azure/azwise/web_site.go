package azwise

import "time"

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
// Intentionally skipped here:
//   - App-kind-specific stack/runtime fields (linuxFxVersion, windowsFxVersion,
//     application_stack, functions_extension_version, storage-account app settings) because
//     Microsoft.Web/sites is shared by web apps, function apps, container apps integration,
//     and slots; AzAPI users can set raw ARM values that AzureRM deliberately partitions.
//   - Auth, backup, diagnostic log, sticky-setting, publishing-credential policy, and
//     storage-account mount rules that AzureRM manages through separate Web Apps subresource
//     APIs rather than the Microsoft.Web/sites body.
//   - AzureRM semantic map validators such as validate.AppSettings and IP restriction
//     one-of checks; azwise has no map-key/cross-field rule type and this assignment does
//     not permit native customizer changes.
//   - IP restriction priority numeric ranges inside arrays; IntRule currently does not
//     evaluate [*] array-element paths, though StringRule does for action enums.
//   - AzureRM computed attributes (default_hostname, outbound IPs, hosting environment,
//     custom_domain_verification_id) are not listed as ComputedFields because the 2023-12-01
//     SDK Site/SiteProperties create model includes those JSON paths; stripping them from
//     AzAPI bodies would risk discarding user-provided raw ARM values.
type WebSite struct {
	BaseKnowledge
}

var _ ResourceKnowledge = (*WebSite)(nil)

// NewWebSite returns knowledge for the sites resource.
func NewWebSite() *WebSite {
	return &WebSite{
		BaseKnowledge: BaseKnowledge{
			ResourceType: "Microsoft.Web/sites",
			ApiVersions:  []string{"2023-01-01", "2023-12-01"},
			ForceNew: []ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
			},
			RequiredFields: []string{
				"properties.serverFarmId",
			},
			TimeoutsConfig: &Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []StringRule{
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
			IntRules: []IntRule{
				{
					PropertyPath: "properties.dailyMemoryTimeQuota",
					MinValue:     ptr(int64(0)),
					Message:      "daily memory time quota must be non-negative",
				},
				{
					PropertyPath: "properties.siteConfig.numberOfWorkers",
					MinValue:     ptr(int64(1)),
					MaxValue:     ptr(int64(100)),
					Message:      "worker count must be between 1 and 100",
				},
			},
			SensitiveFields: []string{
				"properties.siteConfig.connectionStrings[*].connectionString",
			},
			DefaultValues: []DefaultValue{
				{PropertyPath: "properties.clientAffinityEnabled", Value: false},
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
