package datadog

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DatadogMonitor provides resource knowledge for Microsoft.Datadog/monitors.
//
// Mirrors azurerm_datadog_monitor.
//
// Sources:
//   - terraform-provider-azurerm internal/services/datadog/datadog_monitor_resource.go
//     schema (lines 45-166): name ForceNew + name regex; datadog_organization block
//     (api_key/application_key/enterprise_app_id/linking_auth_code/linking_client_id/
//     redirect_uri all ForceNew; api_key/application_key required+sensitive; linking_*
//     sensitive; name/id computed); sku_name required (DiffSuppress, no enum); user block
//     (name/email required+ForceNew, phone_number optional+ForceNew); monitoring_enabled
//     default true; marketplace_subscription_status computed; Create (170-215)
//     MonitoringStatus Enabled/Disabled from monitoring_enabled; timeouts 30m/5m/30m/30m.
//   - internal/services/datadog/validate/datadog_monitor_name.go (name regex),
//     datadog_users_name.go (user name 1-50), datadog_email_address.go (email regex),
//     datadog_phone_number.go (phone <=40).
//   - go-azure-sdk resource-manager/datadog/2021-03-01/monitorsresource:
//     MonitorProperties (datadogOrganizationProperties/userInfo/monitoringStatus settable;
//     marketplaceSubscriptionStatus/provisioningState/liftrResourceCategory read-only),
//     DatadogOrganizationProperties (name/id server-assigned), ResourceSku (name required),
//     constants MonitoringStatus{Disabled,Enabled}; id_monitor.go type Microsoft.Datadog/monitors.
type DatadogMonitor struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DatadogMonitor)(nil)

// NewDatadogMonitor returns knowledge for the monitors resource.
func NewDatadogMonitor() *DatadogMonitor {
	return &DatadogMonitor{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Datadog/monitors",
			ApiVersions:  []string{"2021-03-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.datadogOrganizationProperties.apiKey"},
				{PropertyPath: "properties.datadogOrganizationProperties.applicationKey"},
				{PropertyPath: "properties.datadogOrganizationProperties.enterpriseAppId"},
				{PropertyPath: "properties.datadogOrganizationProperties.linkingAuthCode"},
				{PropertyPath: "properties.datadogOrganizationProperties.linkingClientId"},
				{PropertyPath: "properties.datadogOrganizationProperties.redirectUri"},
				{PropertyPath: "properties.userInfo.name"},
				{PropertyPath: "properties.userInfo.emailAddress"},
				{PropertyPath: "properties.userInfo.phoneNumber"},
			},
			// Required for creation: sku.name (json:"name", no omitempty); the datadog org
			// api_key/application_key and user name/email are Required in the AzureRM schema.
			RequiredFields: []string{
				"sku.name",
				"properties.datadogOrganizationProperties.apiKey",
				"properties.datadogOrganizationProperties.applicationKey",
				"properties.userInfo.name",
				"properties.userInfo.emailAddress",
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). DatadogMonitorsName.
					Regex:     `^[a-zA-Z0-9_-]{2,32}$`,
					MinLength: 2,
					MaxLength: 32,
					Message:   "must be 2-32 characters, alphanumeric plus underscore and hyphen",
				},
				{
					// DatadogUsersName: non-empty, length <= 50.
					PropertyPath: "properties.userInfo.name",
					MinLength:    1,
					MaxLength:    50,
					Message:      "user name must be 1-50 characters",
				},
				{
					// DatadogMonitorsEmailAddress regex.
					PropertyPath: "properties.userInfo.emailAddress",
					Regex:        `^[A-Za-z0-9._%+-]+@(?:[A-Za-z0-9-]+\.)+[A-Za-z]{2,}$`,
					Message:      "user email must be a valid email address",
				},
				{
					// DatadogMonitorsPhoneNumber: length <= 40.
					PropertyPath: "properties.userInfo.phoneNumber",
					MaxLength:    40,
					Message:      "user phone_number must be 40 characters or fewer",
				},
			},
			SensitiveFields: []string{
				"properties.datadogOrganizationProperties.apiKey",
				"properties.datadogOrganizationProperties.applicationKey",
				"properties.datadogOrganizationProperties.linkingAuthCode",
				"properties.datadogOrganizationProperties.linkingClientId",
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.marketplaceSubscriptionStatus",
				"properties.provisioningState",
				"properties.liftrResourceCategory",
				"properties.datadogOrganizationProperties.name",
				"properties.datadogOrganizationProperties.id",
			},
			// monitoring_enabled defaults to true → MonitoringStatus "Enabled".
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.monitoringStatus", Value: "Enabled"},
			},
		},
	}
}

func init() { azwise.Register(NewDatadogMonitor()) }
