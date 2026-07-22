// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package monitor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ActionGroup provides resource knowledge for Microsoft.Insights/actionGroups.
//
// Sources:
//   - terraform-provider-azurerm internal/services/monitor/monitor_action_group_resource.go:31-523
//     (azurerm_monitor_action_group schema: ForceNew, timeouts, validators, defaults; create
//     mapping to actiongroupsapis.ActionGroupResource / ActionGroup)
//   - terraform-provider-azurerm vendor/.../insights/2023-01-01/actiongroupsapis/model_actiongroup.go:6-20
//   - terraform-provider-azurerm vendor/.../insights/2023-01-01/actiongroupsapis/model_actiongroupresource.go:6-13
//
// Intentionally skipped here:
//   - email_receiver / itsm_receiver / azure_app_push_receiver / sms_receiver / webhook_receiver /
//     automation_runbook_receiver / voice_receiver / logic_app_receiver / azure_function_receiver /
//     arm_role_receiver / event_hub_receiver: each is an array block under properties.*Receivers[*];
//     their per-element validators (StringIsNotEmpty, IsUUID, StringIsJSON) are array-element paths,
//     not expressible as declarative body rules.
//   - location Default "global": envelope field (ActionGroupResource.Location), not a properties.* path.
type ActionGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ActionGroup)(nil)

func NewActionGroup() *ActionGroup {
	return &ActionGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Insights/actionGroups",
			ApiVersions:  []string{"2023-01-01"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					MinLength: 1,
					Message:   "Monitor action group name must not be empty",
				},
				{
					PropertyPath: "properties.groupShortName",
					MinLength:    1,
					MaxLength:    12,
					Message:      "short_name must be between 1 and 12 characters",
				},
			},
			// groupShortName and enabled are non-pointer in the ActionGroup model; short_name is
			// Required in schema (enabled has a Default of true, so it is not user-required).
			RequiredFields: []string{
				"properties.groupShortName",
			},
			DefaultValues: []azwise.DefaultValue{
				// enabled Default:true -> properties.enabled true.
				{PropertyPath: "properties.enabled", Value: true},
			},
		},
	}
}

func init() { azwise.Register(NewActionGroup()) }
