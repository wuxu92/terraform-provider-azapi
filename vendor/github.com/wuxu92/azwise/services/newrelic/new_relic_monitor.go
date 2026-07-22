package newrelic

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NewRelicMonitor provides resource knowledge for NewRelic.Observability/monitors.
//
// Mirrors azurerm_new_relic_monitor.
//
// ARM type casing verified against go-azure-sdk monitors/id_monitor.go Segments:
// resource provider "NewRelic.Observability" (no "Microsoft." prefix), static segment
// "monitors".
//
// Sources:
//   - terraform-provider-azurerm internal/services/newrelic/new_relic_monitor_resource.go
//     Arguments (73-234: name StringMatch ^[\w\-]{1,32}$ ForceNew; plan block ForceNew with
//     effective_date/billing_cycle/plan_id/usage_type; user block ForceNew; account_creation_source/
//     org_creation_source enums with defaults; account_id/organization_id Optional+Computed;
//     identity SystemAssignedOptionalForceNew; ingestion_key Sensitive ForceNew), Create (240-292:
//     create-only sdk.Resource, no Update), expand funcs (391-491), timeouts Create 30m / Read 5m /
//     Delete 30m.
//   - go-azure-sdk resource-manager/newrelic/2024-03-01/monitors
//     MonitorProperties (accountCreationSource/orgCreationSource/newRelicAccountProperties/planData/
//     userInfo settable; liftr*/marketplace*/monitoringStatus/provisioningState/saaS*/subscriptionState
//     read-only), NewRelicAccountProperties (accountInfo/organizationInfo/userId), AccountInfo
//     (accountId/ingestionKey), PlanData (billingCycle/effectiveDate/planDetails/usageType), UserInfo
//     (emailAddress/firstName/lastName/phoneNumber), constants.go AccountCreationSource [LIFTR,
//     NEWRELIC] / OrgCreationSource [LIFTR, NEWRELIC] / UsageType [COMMITTED, PAYG].
//
// Note: plan_id maps to properties.planData.planDetails but AzureRM appends a marketplace plan
// suffix (PlanSuffix) to the user value, so its enum ("newrelic-pay-as-you-go-free-live") is not
// emitted as a StringRule — the ARM value differs from the user input.
type NewRelicMonitor struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NewRelicMonitor)(nil)

// NewNewRelicMonitor returns knowledge for the NewRelic monitors resource.
func NewNewRelicMonitor() *NewRelicMonitor {
	return &NewRelicMonitor{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "NewRelic.Observability/monitors",
			ApiVersions:  []string{"2024-03-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// The resource is create-only (sdk.Resource, no Update); every field is ForceNew.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "identity"},
				{PropertyPath: "properties.accountCreationSource"},
				{PropertyPath: "properties.orgCreationSource"},
				{PropertyPath: "properties.planData"},
				{PropertyPath: "properties.userInfo"},
				{PropertyPath: "properties.newRelicAccountProperties.accountInfo.accountId"},
				{PropertyPath: "properties.newRelicAccountProperties.accountInfo.ingestionKey"},
				{PropertyPath: "properties.newRelicAccountProperties.organizationInfo.organizationId"},
				{PropertyPath: "properties.newRelicAccountProperties.userId"},
			},
			RequiredFields: []string{
				"properties.planData.effectiveDate",
				"properties.userInfo.emailAddress",
				"properties.userInfo.firstName",
				"properties.userInfo.lastName",
				"properties.userInfo.phoneNumber",
			},
			StringRules: []azwise.StringRule{
				{
					// name StringMatch (resource.go:79-82).
					PropertyPath: "name",
					MinLength:    1,
					MaxLength:    32,
					Regex:        `^[\w\-]{1,32}$`,
					Message:      "name must be 1-32 characters and may contain only letters, numbers, hyphens and underscores",
				},
				{
					PropertyPath:  "properties.accountCreationSource",
					AllowedValues: []string{"LIFTR", "NEWRELIC"},
				},
				{
					PropertyPath:  "properties.orgCreationSource",
					AllowedValues: []string{"LIFTR", "NEWRELIC"},
				},
				{
					// billingCycle is a free string in the ARM model; AzureRM restricts it
					// (azure-rest-api-specs#31093 removed the enum).
					PropertyPath:  "properties.planData.billingCycle",
					AllowedValues: []string{"MONTHLY", "WEEKLY", "YEARLY"},
				},
				{
					PropertyPath:  "properties.planData.usageType",
					AllowedValues: []string{"COMMITTED", "PAYG"},
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.accountCreationSource", Value: "LIFTR"},
				{PropertyPath: "properties.orgCreationSource", Value: "LIFTR"},
				{PropertyPath: "properties.planData.billingCycle", Value: "MONTHLY"},
				{PropertyPath: "properties.planData.usageType", Value: "PAYG"},
			},
			SensitiveFields: []string{
				"properties.newRelicAccountProperties.accountInfo.ingestionKey",
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.monitoringStatus",
				"properties.marketplaceSubscriptionId",
				"properties.marketplaceSubscriptionStatus",
				"properties.liftrResourceCategory",
				"properties.liftrResourcePreference",
				"properties.saaSAzureSubscriptionStatus",
				"properties.subscriptionState",
			},
		},
	}
}

func init() { azwise.Register(NewNewRelicMonitor()) }
