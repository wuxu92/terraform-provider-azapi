// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package servicebus

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ServiceBusSubscriptionRule provides resource knowledge for
// Microsoft.ServiceBus/namespaces/topics/subscriptions/rules.
//
// Contributing Terraform resource: azurerm_servicebus_subscription_rule.
//
// Scoping verified via SDK id parser: RuleId.ID() formats
// .../Microsoft.ServiceBus/namespaces/%s/topics/%s/subscriptions/%s/rules/%s
// (rules/id_rule.go L122-123).
//
// Sources:
//   - terraform-provider-azurerm internal/services/servicebus/servicebus_subscription_rule_resource.go
//     (schema L46-185, Create body L223-250)
//   - go-azure-sdk resource-manager/servicebus/2024-01-01/rules:
//     model_ruleproperties.go, model_correlationfilter.go, constants.go (FilterType L14-17)
//
// Notes:
//   - name/subscription_id are envelope / parent-reference fields; not emitted as body rules.
//   - filter_type maps to properties.filterType and is Required.
//   - action maps to properties.action.sqlExpression; sql_filter maps to
//     properties.sqlFilter.sqlExpression. correlation_filter maps to properties.correlationFilter
//     (correlationId / messageId / to / replyTo / label / sessionId / replyToSessionId /
//     contentType / properties). sql_filter ConflictsWith correlation_filter, and the
//     correlation_filter sub-fields are AtLeastOneOf — those are nested-block constraints
//     inside a single ARM sub-object, not top-level body paths; documented, not emitted.
//   - sql_filter_compatibility_level is Computed (reserved, hard-coded to 20 by the service).
type ServiceBusSubscriptionRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ServiceBusSubscriptionRule)(nil)

// NewServiceBusSubscriptionRule returns knowledge for the
// namespaces/topics/subscriptions/rules resource.
func NewServiceBusSubscriptionRule() *ServiceBusSubscriptionRule {
	return &ServiceBusSubscriptionRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ServiceBus/namespaces/topics/subscriptions/rules",
			ApiVersions:  []string{"2024-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "", // resource name — StringLenBetween(1, 50)
					MinLength:    1,
					MaxLength:    50,
					Message:      "subscription rule name must be between 1 and 50 characters long",
				},
				{
					PropertyPath:  "properties.filterType",
					AllowedValues: []string{"SqlFilter", "CorrelationFilter"},
					Message:       "filter_type must be one of SqlFilter, CorrelationFilter",
				},
			},
			RequiredFields: []string{
				"properties.filterType",
			},
			ComputedFields: []string{
				"properties.sqlFilter.compatibilityLevel",
			},
		},
	}
}

func init() { azwise.Register(NewServiceBusSubscriptionRule()) }
