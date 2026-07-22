// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventgrid

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EventGridSystemTopic provides resource knowledge for Microsoft.EventGrid/systemTopics.
//
// Sources:
//   - terraform-provider-azurerm internal/services/eventgrid/eventgrid_system_topic_resource.go:31-189
//     (azurerm_eventgrid_system_topic schema: ForceNew, required source/topic_type, timeouts;
//     create mapping to systemtopics.SystemTopic / SystemTopicProperties)
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/systemtopics/model_systemtopicproperties.go:6-11
//
// Intentionally skipped here:
//   - identity: envelope `identity` block, not a properties.* body field.
//   - source_resource_id ValidateFunc (azure.ValidateResourceID / TenantScopedManagementGroupID):
//     a semantic resource-ID validator with no fixed enum/regex expressible declaratively —
//     belongs in an azapin customizer if enforced.
//   - topic_type ValidateFunc StringIsNotEmpty: non-emptiness only, no declarative rule.
type EventGridSystemTopic struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EventGridSystemTopic)(nil)

func NewEventGridSystemTopic() *EventGridSystemTopic {
	return &EventGridSystemTopic{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.EventGrid/systemTopics",
			ApiVersions:  []string{"2025-02-15"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "properties.source"},
				{PropertyPath: "properties.topicType"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					Regex:     `^[-a-zA-Z0-9]{3,128}$`,
					MinLength: 3,
					MaxLength: 128,
					Message:   "EventGrid system topic name must be 3-128 characters long, contain only letters, numbers and hyphens",
				},
			},
			ComputedFields: []string{
				"properties.metricResourceId",
				"properties.provisioningState",
			},
			RequiredFields: []string{
				"properties.source",
				"properties.topicType",
			},
		},
	}
}

func init() { azwise.Register(NewEventGridSystemTopic()) }
