package dynatrace

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DynatraceMonitor provides resource knowledge for Microsoft.Dynatrace/monitors.
//
// Mirrors azurerm_dynatrace_monitor.
//
// Sources:
//   - terraform-provider-azurerm internal/services/dynatrace/dynatrace_monitor_resource.go
//     schema Arguments (lines 54-192: name/monitoring_enabled/marketplace_subscription/plan/user
//     ForceNew; marketplace_subscription StringInSlice enum; plan.billing_cycle/usage_type enums;
//     plan.effective_date Computed), Create (207-265), timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/dynatrace/2023-04-27/monitors
//     MonitorProperties (marketplaceSubscriptionStatus/monitoringStatus/planData/userInfo settable;
//     provisioningState/liftrResourceCategory/liftrResourcePreference server-populated),
//     PlanData (planDetails required, effectiveDate read-only), UserInfo (emailAddress/firstName/
//     lastName), constants.go PossibleValuesForMarketplaceSubscriptionStatus (Active, Suspended).
type DynatraceMonitor struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DynatraceMonitor)(nil)

// NewDynatraceMonitor returns knowledge for the Dynatrace monitors resource.
func NewDynatraceMonitor() *DynatraceMonitor {
	return &DynatraceMonitor{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Dynatrace/monitors",
			ApiVersions:  []string{"2023-04-27"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.monitoringStatus"},
				{PropertyPath: "properties.marketplaceSubscriptionStatus"},
				{PropertyPath: "properties.planData"},
				{PropertyPath: "properties.userInfo"},
			},
			RequiredFields: []string{
				"properties.marketplaceSubscriptionStatus",
				"properties.planData.planDetails",
				"properties.userInfo.emailAddress",
				"properties.userInfo.firstName",
				"properties.userInfo.lastName",
			},
			StringRules: []azwise.StringRule{
				{
					// marketplaceSubscriptionStatus enum (SDK constants.go).
					PropertyPath:  "properties.marketplaceSubscriptionStatus",
					AllowedValues: []string{"Active", "Suspended"},
				},
				{
					// planData.billingCycle is a free string in the ARM model; AzureRM restricts it.
					// See azure-rest-api-specs#31284 (should be an enum).
					PropertyPath:  "properties.planData.billingCycle",
					AllowedValues: []string{"MONTHLY", "WEEKLY", "YEARLY"},
				},
				{
					// planData.usageType is a free string in the ARM model; AzureRM restricts it.
					PropertyPath:  "properties.planData.usageType",
					AllowedValues: []string{"PAYG", "COMMITTED"},
				},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.liftrResourceCategory",
				"properties.liftrResourcePreference",
				"properties.planData.effectiveDate",
			},
		},
	}
}

func init() { azwise.Register(NewDynatraceMonitor()) }
