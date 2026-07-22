// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventgrid

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EventGridPartnerRegistration provides resource knowledge for Microsoft.EventGrid/partnerRegistrations.
//
// Sources:
//   - terraform-provider-azurerm internal/services/eventgrid/eventgrid_partner_registration_resource.go:26-107
//     (azurerm_eventgrid_partner_registration schema: ForceNew name; create mapping to
//     partnerregistrations.PartnerRegistration — no user-settable body properties)
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/partnerregistrations/model_partnerregistrationproperties.go:6-9
//
// Intentionally skipped here:
//   - location: hardcoded "global" envelope location, not a body property.
type EventGridPartnerRegistration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EventGridPartnerRegistration)(nil)

func NewEventGridPartnerRegistration() *EventGridPartnerRegistration {
	return &EventGridPartnerRegistration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.EventGrid/partnerRegistrations",
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
					Regex:     `^[-a-zA-Z0-9]{3,50}$`,
					MinLength: 3,
					MaxLength: 50,
					Message:   "EventGrid partner registration name must be 3-50 characters long, contain only letters, numbers and hyphens",
				},
			},
			ComputedFields: []string{
				"properties.partnerRegistrationImmutableId",
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewEventGridPartnerRegistration()) }
