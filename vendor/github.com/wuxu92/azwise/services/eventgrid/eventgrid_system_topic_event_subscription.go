// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventgrid

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EventGridSystemTopicEventSubscription provides resource knowledge for
// Microsoft.EventGrid/systemTopics/eventSubscriptions.
//
// This is the systemTopic-scoped event subscription: AzureRM's
// azurerm_eventgrid_system_topic_event_subscription is created with
// eventsubscriptions.NewSystemTopicEventSubscriptionID(sub, rg, systemTopic, name), whose ARM ID is
// ".../providers/Microsoft.EventGrid/systemTopics/{systemTopic}/eventSubscriptions/{name}". The ARM
// type segment (systemTopics/eventSubscriptions) DIFFERS from the arbitrary-scope subscription
// (Microsoft.EventGrid/eventSubscriptions), so this is a SEPARATE file from
// eventgrid_event_subscription.go — the two do NOT merge.
//
// Both resources share the same eventsubscriptions.EventSubscriptionProperties body model.
//
// Sources:
//   - terraform-provider-azurerm internal/services/eventgrid/eventgrid_system_topic_event_subscription_resource.go:38-242
//     (azurerm_eventgrid_system_topic_event_subscription schema: ForceNew name/system_topic,
//     timeouts, event_delivery_schema; create mapping to EventSubscriptionProperties)
//   - terraform-provider-azurerm internal/services/eventgrid/event_subscription_schema.go:35-57
//     (name regex, event_delivery_schema enum)
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/eventsubscriptions/model_eventsubscriptionproperties.go:14-26
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/eventsubscriptions/constants.go:327-329 (EventDeliverySchema)
//
// Intentionally skipped here:
//   - system_topic / resource_group_name: parent ID segment / envelope, not body properties.
//   - destination / dead-letter endpoints, subject_filter / advanced_filter / included_event_types,
//     identity blocks: same polymorphic-union / nested-filter reasons as
//     eventgrid_event_subscription.go — not expressible as declarative body-path rules.
type EventGridSystemTopicEventSubscription struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EventGridSystemTopicEventSubscription)(nil)

func NewEventGridSystemTopicEventSubscription() *EventGridSystemTopicEventSubscription {
	return &EventGridSystemTopicEventSubscription{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.EventGrid/systemTopics/eventSubscriptions",
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
					Regex:     `^[-a-zA-Z0-9]{3,64}$`,
					MinLength: 3,
					MaxLength: 64,
					Message:   "EventGrid event subscription name must be 3-64 characters long, contain only letters, numbers and hyphens",
				},
				{
					PropertyPath:  "properties.eventDeliverySchema",
					AllowedValues: []string{"CloudEventSchemaV1_0", "CustomInputSchema", "EventGridSchema"},
					Message:       "must be one of CloudEventSchemaV1_0, CustomInputSchema or EventGridSchema",
				},
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.topic",
			},
			DefaultValues: []azwise.DefaultValue{
				// event_delivery_schema Default:EventGridSchema.
				{PropertyPath: "properties.eventDeliverySchema", Value: "EventGridSchema"},
			},
		},
	}
}

func init() { azwise.Register(NewEventGridSystemTopicEventSubscription()) }
