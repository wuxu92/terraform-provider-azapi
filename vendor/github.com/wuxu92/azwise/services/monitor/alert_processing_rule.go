// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package monitor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AlertProcessingRule provides resource knowledge for Microsoft.AlertsManagement/actionRules.
//
// This ARM type is surfaced by two AzureRM resources that share one API shape and differ only
// in the discriminated properties.actions[] entry they emit:
//   - azurerm_monitor_alert_processing_rule_action_group -> AddActionGroups action
//   - azurerm_monitor_alert_processing_rule_suppression   -> RemoveAllActionGroups action
//
// Only knowledge universal to BOTH resources is captured here (name/scopes/enabled/description,
// timeouts, name validation). The differing actions block and the condition/schedule blocks are
// left out — they are discriminated or nested/array-element paths.
//
// Sources:
//   - terraform-provider-azurerm internal/services/monitor/alert_processing_rule.go:68-98
//     (shared schemaAlertProcessingRule: name ForceNew + validator, scopes, description, enabled)
//   - terraform-provider-azurerm internal/services/monitor/monitor_alert_processing_rule_action_group_resource.go:47-113,186,252
//   - terraform-provider-azurerm internal/services/monitor/monitor_alert_processing_rule_suppression_resource.go:45-101
//   - terraform-provider-azurerm internal/services/monitor/validate/alert_processing_rule_name.go:17
//   - terraform-provider-azurerm vendor/.../alertsmanagement/2021-08-08/alertprocessingrules/model_alertprocessingruleproperties.go:11-18
//
// Intentionally skipped here:
//   - condition: array block (properties.conditions[*]) with per-field operator/value enums on
//     array-element paths.
//   - schedule: nested block (properties.schedule.*) with recurrence sub-blocks.
//   - actions / add_action_group_ids: discriminated properties.actions[*] union; differs per
//     contributing resource, so not universal and not expressible declaratively.
//   - location: hardcoded to "global" by both resources (envelope field).
type AlertProcessingRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AlertProcessingRule)(nil)

func NewAlertProcessingRule() *AlertProcessingRule {
	return &AlertProcessingRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AlertsManagement/actionRules",
			ApiVersions:  []string{"2021-08-08"},
			SoftDelete:   false,
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
					// validate.AlertProcessingRuleName: begins with a letter or number, then only
					// letters, numbers, underscores and hyphens.
					Regex:   `^([a-zA-Z\d])[a-zA-Z\d-_]*$`,
					Message: "name should begin with a letter or number, contain only letters, numbers, underscores and hyphens",
				},
			},
			// Scopes and Actions are non-pointer (required) in AlertProcessingRuleProperties.
			RequiredFields: []string{
				"properties.scopes",
				"properties.actions",
			},
			DefaultValues: []azwise.DefaultValue{
				// enabled Default:true -> properties.enabled true (universal to both resources).
				{PropertyPath: "properties.enabled", Value: true},
			},
		},
	}
}

func init() { azwise.Register(NewAlertProcessingRule()) }
