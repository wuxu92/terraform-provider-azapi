// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventgrid

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EventGridPartnerNamespace provides resource knowledge for Microsoft.EventGrid/partnerNamespaces.
//
// Sources:
//   - terraform-provider-azurerm internal/services/eventgrid/eventgrid_partner_namespace_resource.go:27-170
//     (azurerm_eventgrid_partner_namespace schema: ForceNew, validators, defaults; create mapping
//     to partnernamespaces.PartnerNamespace / PartnerNamespaceProperties)
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/partnernamespaces/model_partnernamespaceproperties.go:6-16
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/partnernamespaces/constants.go:106-107
//     (PartnerTopicRoutingMode), :194-195 (PublicNetworkAccess), :100 (IPActionType)
//
// Intentionally skipped here:
//   - inbound_ip_rule[].action StringInSlice([Allow]) + Default "Allow": array-element path.
type EventGridPartnerNamespace struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EventGridPartnerNamespace)(nil)

func NewEventGridPartnerNamespace() *EventGridPartnerNamespace {
	return &EventGridPartnerNamespace{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.EventGrid/partnerNamespaces",
			ApiVersions:  []string{"2025-02-15"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "properties.partnerRegistrationFullyQualifiedId"},
				{PropertyPath: "properties.partnerTopicRoutingMode"},
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
					Message:   "EventGrid partner namespace name must be 3-50 characters long, contain only letters, numbers and hyphens",
				},
				{
					PropertyPath:  "properties.partnerTopicRoutingMode",
					AllowedValues: []string{"ChannelNameHeader", "SourceEventAttribute"},
					Message:       "must be one of ChannelNameHeader or SourceEventAttribute",
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
			},
			ArrayRules: []azwise.ArrayRule{
				{PropertyPath: "properties.inboundIpRules", MaxItems: 16},
			},
			ComputedFields: []string{
				"properties.endpoint",
				"properties.provisioningState",
				"properties.privateEndpointConnections",
			},
			RequiredFields: []string{
				"properties.partnerRegistrationFullyQualifiedId",
			},
			DefaultValues: []azwise.DefaultValue{
				// local_authentication_enabled Default:true -> disableLocalAuth false.
				{PropertyPath: "properties.disableLocalAuth", Value: false},
				// partner_topic_routing_mode Default:ChannelNameHeader.
				{PropertyPath: "properties.partnerTopicRoutingMode", Value: "ChannelNameHeader"},
				// public_network_access Default:Enabled.
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
			},
		},
	}
}

func init() { azwise.Register(NewEventGridPartnerNamespace()) }
