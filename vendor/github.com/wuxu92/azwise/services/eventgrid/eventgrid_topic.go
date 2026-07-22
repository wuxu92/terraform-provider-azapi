// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventgrid

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EventGridTopic provides resource knowledge for Microsoft.EventGrid/topics.
//
// Sources:
//   - terraform-provider-azurerm internal/services/eventgrid/eventgrid_topic_resource.go:27-254
//     (azurerm_eventgrid_topic schema: ForceNew, timeouts, validators, defaults; create mapping
//     to topics.Topic / TopicProperties)
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/topics/model_topicproperties.go:11-24
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/topics/constants.go (InputSchema,
//     PublicNetworkAccess, IPActionType — same value sets as domains)
//
// Intentionally skipped here:
//   - identity: envelope `identity` block, not a properties.* body field.
//   - primary_access_key / secondary_access_key: computed, sensitive data-plane keys
//     retrieved via ListSharedAccessKeys — not part of the topic body.
//   - inbound_ip_rule[].action StringInSlice([Allow]) and its Default "Allow": an
//     array-element path (properties.inboundIpRules[*].action) — not expressible.
type EventGridTopic struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EventGridTopic)(nil)

func NewEventGridTopic() *EventGridTopic {
	return &EventGridTopic{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.EventGrid/topics",
			ApiVersions:  []string{"2025-02-15"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "properties.inputSchema"},
				// input_mapping_fields / input_mapping_default_values are ForceNew and
				// expand into the discriminated properties.inputSchemaMapping union.
				{PropertyPath: "properties.inputSchemaMapping"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					Regex:     `^[-a-zA-Z0-9]{3,50}$`,
					MinLength: 3,
					MaxLength: 50,
					Message:   "EventGrid topic name must be 3-50 characters long, contain only letters, numbers and hyphens",
				},
				{
					PropertyPath:  "properties.inputSchema",
					AllowedValues: []string{"CloudEventSchemaV1_0", "CustomEventSchema", "EventGridSchema"},
					Message:       "must be one of CloudEventSchemaV1_0, CustomEventSchema or EventGridSchema",
				},
			},
			ComputedFields: []string{
				"properties.endpoint",
				"properties.metricResourceId",
				"properties.provisioningState",
				"properties.privateEndpointConnections",
			},
			DefaultValues: []azwise.DefaultValue{
				// public_network_access_enabled Default:true -> publicNetworkAccess Enabled.
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				// local_auth_enabled Default:true -> disableLocalAuth false.
				{PropertyPath: "properties.disableLocalAuth", Value: false},
				// input_schema Default:EventGridSchema.
				{PropertyPath: "properties.inputSchema", Value: "EventGridSchema"},
			},
		},
	}
}

func init() { azwise.Register(NewEventGridTopic()) }
