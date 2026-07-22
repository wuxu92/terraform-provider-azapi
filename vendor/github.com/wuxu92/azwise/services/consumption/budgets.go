package consumption

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Budget provides resource knowledge for Microsoft.Consumption/budgets.
//
// Contributing Terraform resources (all map to the same ARM type; they differ
// ONLY by parent scope, which is the envelope/ID — the request body is identical):
//   - azurerm_consumption_budget_subscription     (scope: subscription_id)
//   - azurerm_consumption_budget_resource_group   (scope: resource_group_id)
//   - azurerm_consumption_budget_management_group  (scope: management_group_id)
//
// Sources:
//   - terraform-provider-azurerm internal/services/consumption/consumption_budget_base.go
//     (shared schema L86-281, Create/Update body via subscription/rg/mgmt resources,
//     expand/flatten helpers L316-524, deleteFunc L287-304)
//   - internal/services/consumption/consumption_budget_subscription_resource.go
//     (schema L43-59, Create body L108-122, timeouts Create 30m/Read 5m/Update 30m)
//   - internal/services/consumption/consumption_budget_resource_group_resource.go (schema L41-57)
//   - internal/services/consumption/consumption_budget_management_group_resource.go
//     (schema L50-116 — management-group notification variant)
//   - internal/services/consumption/validate/name.go (subscription name regex)
//   - go-azure-sdk resource-manager/consumption/2019-10-01/budgets:
//     model_budgetproperties.go, model_notification.go, model_budgettimeperiod.go,
//     model_budgetfilter.go, constants.go (enums).
//
// Notes:
//   - Scope fields (subscription_id / resource_group_id / management_group_id) and the
//     resource name are envelope-owned, not ARM-body properties. name is ForceNew for all
//     three, kept as a ForceNewRule{PropertyPath: "name"}; the scope IDs are the parent and
//     are not body properties, so no body ForceNew rule is emitted for them.
//   - Resource-name validation is NOT unioned: the subscription resource enforces a regex
//     (^[-_a-zA-Z0-9]{1,63}$), while the resource-group and management-group resources only
//     require a non-whitespace string. Applying the stricter regex to all three would reject
//     names that are valid for the rg/mgmt scopes, so no name StringRule is emitted.
//   - properties.category is hard-coded to "Cost" by AzureRM for every scope; the ARM API
//     requires it, so it is listed in RequiredFields.
//   - notification.* validators (threshold IntBetween(0,1000); operator enum EqualTo/
//     GreaterThan/GreaterThanOrEqualTo; threshold_type enum Actual/Forecasted) live inside
//     properties.notifications, which is a map[string]Notification. Map-keyed paths are not
//     resolvable/lowerable, so those rules are documented here but not emitted.
//   - filter.dimension.operator / filter.tag.operator are fixed to "In" (only allowed value).
//   - The management-group notification schema omits contact_groups/contact_roles and requires
//     contact_emails; this is a schema-shape difference on the same ARM body and does not
//     change any universal rule below.
type Budget struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Budget)(nil)

// NewBudget returns knowledge for the Microsoft.Consumption/budgets resource.
func NewBudget() *Budget {
	return &Budget{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Consumption/budgets",
			ApiVersions:  []string{"2019-10-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// name is ForceNew for every scope (envelope). time_grain and
			// time_period.start_date are ForceNew body properties across all three.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.timeGrain"},
				{PropertyPath: "properties.timePeriod.startDate"},
			},
			StringRules: []azwise.StringRule{
				// time_grain → properties.timeGrain (full ARM SDK enum set)
				{
					PropertyPath:  "properties.timeGrain",
					AllowedValues: []string{"Annually", "BillingAnnual", "BillingMonth", "BillingQuarter", "Monthly", "Quarterly"},
					Message:       "must be one of Annually, BillingAnnual, BillingMonth, BillingQuarter, Monthly, or Quarterly",
				},
			},
			// amount → properties.amount (validation.FloatAtLeast(1.0), no upper bound)
			FloatRules: []azwise.FloatRule{
				{PropertyPath: "properties.amount", MinValue: azwise.Ptr(float64(1)), Message: "amount must be at least 1"},
			},
			// AzureRM base schema: filter { dimension AtLeastOneOf tag }.
			AtLeastOneOf: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.filter.dimensions", "properties.filter.tags"},
					Message: "at least one of `filter.dimension` or `filter.tag` must be set",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// time_grain Default "Monthly".
				{PropertyPath: "properties.timeGrain", Value: "Monthly"},
				// time_period.end_date is Optional+Computed — server assigns when omitted.
				{PropertyPath: "properties.timePeriod.endDate"},
			},
			RequiredFields: []string{
				"properties.amount",
				"properties.category", // AzureRM hard-codes "Cost"; ARM requires it
				"properties.timePeriod",
				"properties.timePeriod.startDate",
				"properties.notifications",
			},
			// GET-only spend telemetry returned in BudgetProperties but never set on create.
			ComputedFields: []string{
				"properties.currentSpend",
				"properties.forecastSpend",
			},
		},
	}
}

// Self-registers into the azwise registry.
func init() { azwise.Register(NewBudget()) }
