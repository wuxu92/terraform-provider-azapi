// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package servicebus

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ServiceBusSubscription provides resource knowledge for
// Microsoft.ServiceBus/namespaces/topics/subscriptions.
//
// Contributing Terraform resource: azurerm_servicebus_subscription.
//
// Scoping verified via SDK id parser: Subscriptions2Id.ID() formats
// .../Microsoft.ServiceBus/namespaces/%s/topics/%s/subscriptions/%s
// (subscriptions/id_subscriptions2.go L116-117).
//
// Sources:
//   - terraform-provider-azurerm internal/services/servicebus/servicebus_subscription_resource.go
//     (schema L53-171, Create body L219-256)
//   - terraform-provider-azurerm internal/services/servicebus/validate/subscription_name.go (L13-17)
//   - go-azure-sdk resource-manager/servicebus/2024-01-01/subscriptions:
//     model_sbsubscriptionproperties.go, model_sbclientaffineproperties.go, constants.go (EntityStatus L14-24)
//
// Notes:
//   - name/topic_id are envelope / parent-reference fields; not emitted as body rules.
//   - AzureRM status ValidateFunc accepts Active/Disabled/ReceiveDisabled, but ARM accepts the
//     full EntityStatus set; per azwise policy we emit the full ARM enum.
//   - max_delivery_count is Required in AzureRM (schema L104-107) → properties.maxDeliveryCount.
//   - client_scoped_subscription_enabled maps to properties.isClientAffine; the nested
//     client_scoped_subscription block maps to properties.clientAffineProperties
//     (clientId / isShared / isDurable). client_scoped_subscription.client_id and
//     is_client_scoped_subscription_shareable are ForceNew.
//   - forward_to / forward_dead_lettered_messages_to hold arbitrary entity names in the ARM
//     body (forwardTo / forwardDeadLetteredMessagesTo); no body StringRule emitted.
type ServiceBusSubscription struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ServiceBusSubscription)(nil)

// NewServiceBusSubscription returns knowledge for the
// namespaces/topics/subscriptions resource.
func NewServiceBusSubscription() *ServiceBusSubscription {
	return &ServiceBusSubscription{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ServiceBus/namespaces/topics/subscriptions",
			ApiVersions:  []string{"2024-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.requiresSession"},
				{PropertyPath: "properties.clientAffineProperties.clientId"},
				{PropertyPath: "properties.clientAffineProperties.isShared"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "", // resource name
					Regex:        "^[_a-zA-Z0-9][-._a-zA-Z0-9]{0,48}([_a-zA-Z0-9])?$",
					Message:      "subscription name can contain only letters, numbers, periods, hyphens and underscores, must start and end with a letter, number or underscore, and be at most 50 characters long",
				},
				{
					PropertyPath:  "properties.status",
					AllowedValues: []string{"Active", "Creating", "Deleting", "Disabled", "ReceiveDisabled", "Renaming", "Restoring", "SendDisabled", "Unknown"},
					Message:       "status must be a valid EntityStatus value",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.autoDeleteOnIdle", Value: "P10675199DT2H48M5.4775807S"},
				{PropertyPath: "properties.defaultMessageTimeToLive", Value: "P10675199DT2H48M5.4775807S"},
				{PropertyPath: "properties.lockDuration", Value: "PT1M"},
				{PropertyPath: "properties.deadLetteringOnFilterEvaluationExceptions", Value: true},
				{PropertyPath: "properties.isClientAffine", Value: false},
				{PropertyPath: "properties.status", Value: "Active"},
			},
			RequiredFields: []string{
				"properties.maxDeliveryCount",
			},
			ComputedFields: []string{
				"properties.accessedAt",
				"properties.createdAt",
				"properties.updatedAt",
				"properties.countDetails",
				"properties.messageCount",
				"properties.clientAffineProperties.isDurable",
			},
		},
	}
}

func init() { azwise.Register(NewServiceBusSubscription()) }
