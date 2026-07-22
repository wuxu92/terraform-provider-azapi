// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventgrid

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EventGridEventSubscription provides resource knowledge for Microsoft.EventGrid/eventSubscriptions.
//
// This is the arbitrary-scope event subscription: AzureRM's azurerm_eventgrid_event_subscription
// is created with eventsubscriptions.NewScopedEventSubscriptionID(scope, name), whose ARM ID is
// "/{scope}/providers/Microsoft.EventGrid/eventSubscriptions/{name}". This is a DISTINCT ARM type
// segment from the systemTopic-scoped subscription
// (Microsoft.EventGrid/systemTopics/eventSubscriptions), so it gets its own knowledge file — see
// eventgrid_system_topic_event_subscription.go.
//
// Sources:
//   - terraform-provider-azurerm internal/services/eventgrid/eventgrid_event_subscription_resource.go:38-259
//     (azurerm_eventgrid_event_subscription schema: scope-parameterized ForceNew, timeouts,
//     event_delivery_schema; create mapping to eventsubscriptions.EventSubscription /
//     EventSubscriptionProperties)
//   - terraform-provider-azurerm internal/services/eventgrid/event_subscription_schema.go:35-57
//     (name regex, event_delivery_schema enum)
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/eventsubscriptions/model_eventsubscriptionproperties.go:14-26
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/eventsubscriptions/constants.go:327-329 (EventDeliverySchema)
//
// Intentionally skipped here:
//   - scope: envelope / ID scope segment, not a body property.
//   - destination / dead-letter endpoints (azure_function_endpoint, eventhub_id, hybrid_connection_id,
//     service_bus_queue_id, service_bus_topic_id, storage_queue_endpoint, webhook_endpoint): a
//     polymorphic, mutually-exclusive discriminated union (EventSubscriptionDestination) whose
//     endpointType is chosen at runtime — azwise cannot express the ConflictsWith across these
//     one-of blocks as declarative body-path rules.
//   - subject_filter / advanced_filter / included_event_types: nested properties.filter.* structures;
//     advanced_filter also carries a CustomizeDiff (<=25 total values) with no single body field.
//   - delivery_identity / dead_letter_identity type StringInSlice: array/nested identity blocks.
type EventGridEventSubscription struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EventGridEventSubscription)(nil)

func NewEventGridEventSubscription() *EventGridEventSubscription {
	return &EventGridEventSubscription{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.EventGrid/eventSubscriptions",
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

func init() { azwise.Register(NewEventGridEventSubscription()) }
