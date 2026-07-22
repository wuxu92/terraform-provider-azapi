// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventgrid

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EventGridDomain provides resource knowledge for Microsoft.EventGrid/domains.
//
// Sources:
//   - terraform-provider-azurerm internal/services/eventgrid/eventgrid_domain_resource.go:27-267
//     (azurerm_eventgrid_domain schema: ForceNew, timeouts, validators, defaults; create mapping to domains.Domain / DomainProperties)
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/domains/model_domainproperties.go:11-26
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/domains/constants.go:184-188 (InputSchema),
//     :313-316 (PublicNetworkAccess), :146-148 (IPActionType)
//
// Intentionally skipped here:
//   - identity: commonschema.SystemOrUserAssignedIdentityOptional maps to the top-level
//     envelope `identity` block, not a properties.* body field.
//   - primary_access_key / secondary_access_key: computed, sensitive data-plane keys
//     retrieved via ListSharedAccessKeys — not part of the domain body.
//   - inbound_ip_rule[].action StringInSlice([Allow]) and its Default "Allow": an
//     array-element path (properties.inboundIpRules[*].action); azwise cannot express a
//     per-element enum/default, so it is omitted.
type EventGridDomain struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EventGridDomain)(nil)

func NewEventGridDomain() *EventGridDomain {
	return &EventGridDomain{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.EventGrid/domains",
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
					Message:   "EventGrid domain name must be 3-50 characters long, contain only letters, numbers and hyphens",
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
				// auto_create_topic_with_first_subscription Default:true.
				{PropertyPath: "properties.autoCreateTopicWithFirstSubscription", Value: true},
				// auto_delete_topic_with_last_subscription Default:true.
				{PropertyPath: "properties.autoDeleteTopicWithLastSubscription", Value: true},
				// input_schema Default:EventGridSchema.
				{PropertyPath: "properties.inputSchema", Value: "EventGridSchema"},
			},
		},
	}
}

func init() { azwise.Register(NewEventGridDomain()) }
