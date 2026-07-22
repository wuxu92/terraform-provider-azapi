// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventgrid

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EventGridNamespace provides resource knowledge for Microsoft.EventGrid/namespaces.
//
// Sources:
//   - terraform-provider-azurerm internal/services/eventgrid/eventgrid_namespace_resource.go:61-322,494-533
//     (azurerm_eventgrid_namespace schema: ForceNew, validators, defaults; create mapping to
//     namespaces.Namespace / NamespaceProperties / NamespaceSku / TopicSpacesConfiguration)
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/namespaces/model_namespaceproperties.go:6-15
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/namespaces/model_namespacesku.go:6-9
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/namespaces/model_topicspacesconfiguration.go:6-15
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/namespaces/constants.go:249-252 (PublicNetworkAccess),
//     :387-389 (SkuName), :99-101 (IPActionType)
//
// Intentionally skipped here:
//   - identity: envelope `identity` block, not a properties.* body field.
//   - inbound_ip_rule[].action StringInSlice([Allow]) + Default "Allow": array-element path.
//   - topic_spaces_configuration nested defaults (maximum_client_sessions_per_authentication_name
//     Default:1, maximum_session_expiry_in_hours Default:1): applying a DefaultValue under
//     properties.topicSpacesConfiguration would materialise that optional (ForceNew) block even
//     when the user omits it, so the block-scoped defaults are omitted.
//   - route_topic_id ValidateFunc topics.ValidateTopicID and alternative_authentication_name_source
//     / dynamic_routing_enrichment / static_routing_enrichment validators: nested / array-element
//     or semantic resource-ID validators, not expressible declaratively.
type EventGridNamespace struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EventGridNamespace)(nil)

func NewEventGridNamespace() *EventGridNamespace {
	return &EventGridNamespace{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.EventGrid/namespaces",
			ApiVersions:  []string{"2025-02-15"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "properties.topicSpacesConfiguration"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					MinLength: 3,
					MaxLength: 50,
					Message:   "EventGrid namespace name must be 3-50 characters long",
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Standard"},
					Message:       "must be Standard",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "sku.capacity",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(40)),
					Message:      "capacity must be between 1 and 40",
				},
				{
					PropertyPath: "properties.topicSpacesConfiguration.maximumClientSessionsPerAuthenticationName",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(100)),
					Message:      "maximum_client_sessions_per_authentication_name must be between 1 and 100",
				},
				{
					PropertyPath: "properties.topicSpacesConfiguration.maximumSessionExpiryInHours",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(8)),
					Message:      "maximum_session_expiry_in_hours must be between 1 and 8",
				},
			},
			ArrayRules: []azwise.ArrayRule{
				{PropertyPath: "properties.inboundIpRules", MaxItems: 128},
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.privateEndpointConnections",
			},
			DefaultValues: []azwise.DefaultValue{
				// capacity Default:1.
				{PropertyPath: "sku.capacity", Value: 1},
				// public_network_access Default:Enabled.
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				// sku Default:Standard.
				{PropertyPath: "sku.name", Value: "Standard"},
			},
		},
	}
}

func init() { azwise.Register(NewEventGridNamespace()) }
