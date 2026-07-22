// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eventgrid

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EventGridPartnerConfiguration provides resource knowledge for Microsoft.EventGrid/partnerConfigurations.
//
// The ARM resource is resource-group-scoped (one configuration per resource group); AzureRM keys
// it by the resource group ID.
//
// Sources:
//   - terraform-provider-azurerm internal/services/eventgrid/eventgrid_partner_configuration_resource.go:26-133
//     (azurerm_eventgrid_partner_configuration schema: default_maximum_expiration_time_in_days,
//     partner_authorization; create mapping to partnerconfigurations.PartnerConfiguration /
//     PartnerConfigurationProperties / PartnerAuthorization)
//   - terraform-provider-azurerm vendor/.../eventgrid/2025-02-15/partnerconfigurations/model_partnerconfigurationproperties.go:6-9
//
// Intentionally skipped here:
//   - resource_group_name: envelope / ID scope, not a body property.
//   - location: hardcoded "global" envelope location, not a body property.
//   - partner_authorization[].partner_registration_id (IsUUID), partner_name (StringIsNotEmpty),
//     authorization_expiration_time_in_utc (IsRFC3339Time): array-element / semantic validators
//     under properties.partnerAuthorization.authorizedPartnersList[*] — not expressible declaratively.
type EventGridPartnerConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EventGridPartnerConfiguration)(nil)

func NewEventGridPartnerConfiguration() *EventGridPartnerConfiguration {
	return &EventGridPartnerConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.EventGrid/partnerConfigurations",
			ApiVersions:  []string{"2025-02-15"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.partnerAuthorization.defaultMaximumExpirationTimeInDays",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(365)),
					Message:      "default_maximum_expiration_time_in_days must be between 1 and 365",
				},
			},
			ComputedFields: []string{
				"properties.provisioningState",
			},
			DefaultValues: []azwise.DefaultValue{
				// default_maximum_expiration_time_in_days Default:7.
				{PropertyPath: "properties.partnerAuthorization.defaultMaximumExpirationTimeInDays", Value: 7},
			},
		},
	}
}

func init() { azwise.Register(NewEventGridPartnerConfiguration()) }
