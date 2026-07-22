// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventgrid

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EventGridNamespaceTopic provides resource knowledge for Microsoft.EventGrid/namespaces/topics.
//
// Sources:
//   - terraform-provider-azurerm internal/services/eventgrid/eventgrid_namespace_topic_resource.go:31-128
//     (azurerm_eventgrid_namespace_topic schema: ForceNew name, event_retention_in_days; create
//     mapping to namespacetopics.NamespaceTopic / NamespaceTopicProperties — inputSchema and
//     publisherType are hardcoded by AzureRM)
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/namespacetopics/model_namespacetopicproperties.go:6-11
//
// Intentionally skipped here:
//   - eventgrid_namespace_id: parent namespace ID, not a body property.
type EventGridNamespaceTopic struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EventGridNamespaceTopic)(nil)

func NewEventGridNamespaceTopic() *EventGridNamespaceTopic {
	return &EventGridNamespaceTopic{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.EventGrid/namespaces/topics",
			ApiVersions:  []string{"2025-02-15"},
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
					Regex:     `^[a-zA-Z0-9-]{3,50}$`,
					MinLength: 3,
					MaxLength: 50,
					Message:   "EventGrid namespace topic name must be 3-50 characters long, contain only letters, numbers and hyphens",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.eventRetentionInDays",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(7)),
					Message:      "event_retention_in_days must be between 1 and 7",
				},
			},
			ComputedFields: []string{
				"properties.provisioningState",
			},
			DefaultValues: []azwise.DefaultValue{
				// event_retention_in_days Default:7.
				{PropertyPath: "properties.eventRetentionInDays", Value: 7},
				// AzureRM hardcodes inputSchema and publisherType on create.
				{PropertyPath: "properties.inputSchema", Value: "CloudEventSchemaV1_0"},
				{PropertyPath: "properties.publisherType", Value: "Custom"},
			},
		},
	}
}

func init() { azwise.Register(NewEventGridNamespaceTopic()) }
