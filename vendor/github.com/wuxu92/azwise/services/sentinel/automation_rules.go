// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package sentinel

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationRules provides resource knowledge for
// Microsoft.SecurityInsights/automationRules (an extension resource on a
// Microsoft.OperationalInsights/workspaces scope).
//
// Contributing Terraform resource: azurerm_sentinel_automation_rule.
//
// Sources:
//   - terraform-provider-azurerm internal/services/sentinel/sentinel_automation_rule_resource.go
//     (schema L29-92 & timeouts L224-229, body L267-289)
//   - go-azure-sdk resource-manager/securityinsights/2024-09-01/automationrules:
//     id_automationrule.go (Microsoft.SecurityInsights/automationRules),
//     constants.go (TriggersOn, TriggersWhen), model_automationruleproperties.go
//     (displayName, order, triggeringLogic), model_automationruletriggeringlogic.go
//     (isEnabled, triggersOn, triggersWhen)
//
// Notes:
//   - name is a UUID (validation.IsUUID) and is the resource name (PropertyPath "").
//   - enabled/triggers_on/triggers_when live under properties.triggeringLogic.
type AutomationRules struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationRules)(nil)

// NewAutomationRules returns knowledge for the automationRules resource type.
func NewAutomationRules() *AutomationRules {
	return &AutomationRules{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.SecurityInsights/automationRules",
			ApiVersions:  []string{"2024-09-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 5 * time.Minute,
				Read:   5 * time.Minute,
				Update: 5 * time.Minute,
				Delete: 5 * time.Minute,
			},
			RequiredFields: []string{
				"properties.displayName",
				"properties.order",
				"properties.triggeringLogic",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
					Message:      "automation rule name must be a valid UUID",
				},
				{
					PropertyPath: "properties.displayName",
					MinLength:    1,
					Message:      "display_name must not be empty",
				},
				{
					PropertyPath:  "properties.triggeringLogic.triggersOn",
					AllowedValues: []string{"Alerts", "Incidents"},
					Message:       "triggers_on must be one of Alerts, Incidents",
				},
				{
					PropertyPath:  "properties.triggeringLogic.triggersWhen",
					AllowedValues: []string{"Created", "Updated"},
					Message:       "triggers_when must be one of Created, Updated",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.order",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(1000)),
					Message:      "order must be between 1 and 1000",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.triggeringLogic.triggersOn", Value: "Incidents"},
				{PropertyPath: "properties.triggeringLogic.triggersWhen", Value: "Created"},
				{PropertyPath: "properties.triggeringLogic.isEnabled", Value: true},
			},
		},
	}
}

func init() { azwise.Register(NewAutomationRules()) }
