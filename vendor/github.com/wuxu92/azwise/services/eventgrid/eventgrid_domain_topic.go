// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventgrid

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EventGridDomainTopic provides resource knowledge for Microsoft.EventGrid/domains/topics.
//
// Sources:
//   - terraform-provider-azurerm internal/services/eventgrid/eventgrid_domain_topic_resource.go:23-127
//     (azurerm_eventgrid_domain_topic schema: ForceNew name + parent domain_name, timeouts;
//     create has no settable body — DomainTopic carries only a read-only provisioningState)
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/domaintopics/model_domaintopicproperties.go:6-8
//
// Intentionally skipped here:
//   - domain_name: parent ID segment (domains name), not a body property.
type EventGridDomainTopic struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EventGridDomainTopic)(nil)

func NewEventGridDomainTopic() *EventGridDomainTopic {
	return &EventGridDomainTopic{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.EventGrid/domains/topics",
			ApiVersions:  []string{"2025-02-15"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					Regex:     `^[-a-zA-Z0-9]{3,128}$`,
					MinLength: 3,
					MaxLength: 128,
					Message:   "EventGrid domain topic name must be 3-128 characters long, contain only letters, numbers and hyphens",
				},
			},
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewEventGridDomainTopic()) }
