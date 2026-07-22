// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package servicebus

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ServiceBusQueue provides resource knowledge for Microsoft.ServiceBus/namespaces/queues.
//
// Contributing Terraform resource: azurerm_servicebus_queue.
//
// Scoping verified via SDK id parser: QueueId.ID() formats
// .../Microsoft.ServiceBus/namespaces/%s/queues/%s (queues/id_queue.go L110-111).
//
// Sources:
//   - terraform-provider-azurerm internal/services/servicebus/servicebus_queue_resource.go
//     (schema L49-183, Create body L258-329)
//   - terraform-provider-azurerm internal/services/servicebus/validate/queue_name.go (L13-17),
//     validate/servicebus.go (ServiceBusMaxMessageSizeInKilobytes IntBetween 1024-102400 L25-27)
//   - go-azure-sdk resource-manager/servicebus/2024-01-01/queues:
//     model_sbqueueproperties.go, constants.go (EntityStatus L14-24)
//
// Notes:
//   - name/namespace_id are envelope / parent-reference fields; not emitted as body rules.
//   - forward_to / forward_dead_lettered_messages_to validate as queue names in AzureRM but
//     hold arbitrary entity names in the ARM body (forwardTo / forwardDeadLetteredMessagesTo);
//     no body StringRule emitted for them.
//   - max_size_in_megabytes uses IntInSlice{1024..81920} (discrete, non-contiguous) which
//     IntRule (min/max) cannot express; documented, not emitted.
//   - auto_delete_on_idle / default_message_ttl are Optional+Computed with sku-dependent
//     server defaults (no static ARM default); represented as nil-valued DefaultValues.
//   - partitioning_enabled, requires_duplicate_detection, requires_session are ForceNew.
type ServiceBusQueue struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ServiceBusQueue)(nil)

// NewServiceBusQueue returns knowledge for the namespaces/queues resource.
func NewServiceBusQueue() *ServiceBusQueue {
	return &ServiceBusQueue{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ServiceBus/namespaces/queues",
			ApiVersions:  []string{"2024-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.enablePartitioning"},
				{PropertyPath: "properties.requiresDuplicateDetection"},
				{PropertyPath: "properties.requiresSession"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "", // resource name
					Regex:        `^[a-zA-Z0-9][\w-./~]{0,258}([a-zA-Z0-9])?$`,
					Message:      "queue name can contain only letters, numbers, periods, hyphens, tildes, forward slashes and underscores, must start and end with a letter or number, and be less than 260 characters long",
				},
				{
					PropertyPath:  "properties.status",
					AllowedValues: []string{"Active", "Creating", "Deleting", "Disabled", "ReceiveDisabled", "Renaming", "Restoring", "SendDisabled", "Unknown"},
					Message:       "status must be a valid EntityStatus value",
				},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.maxDeliveryCount", MinValue: azwise.Ptr(int64(1)), Message: "max_delivery_count must be at least 1"},
				{PropertyPath: "properties.maxMessageSizeInKilobytes", MinValue: azwise.Ptr(int64(1024)), MaxValue: azwise.Ptr(int64(102400)), Message: "max_message_size_in_kilobytes must be between 1024 and 102400"},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.deadLetteringOnMessageExpiration", Value: false},
				{PropertyPath: "properties.duplicateDetectionHistoryTimeWindow", Value: "PT10M"},
				{PropertyPath: "properties.enableBatchedOperations", Value: true},
				{PropertyPath: "properties.enableExpress", Value: false},
				{PropertyPath: "properties.enablePartitioning", Value: false},
				{PropertyPath: "properties.lockDuration", Value: "PT1M"},
				{PropertyPath: "properties.maxDeliveryCount", Value: float64(10)},
				{PropertyPath: "properties.requiresDuplicateDetection", Value: false},
				{PropertyPath: "properties.requiresSession", Value: false},
				{PropertyPath: "properties.status", Value: "Active"},
				// sku-dependent server defaults (no static AzureRM default).
				{PropertyPath: "properties.autoDeleteOnIdle", Value: nil},
				{PropertyPath: "properties.defaultMessageTimeToLive", Value: nil},
				{PropertyPath: "properties.maxSizeInMegabytes", Value: nil},
			},
			ComputedFields: []string{
				"properties.accessedAt",
				"properties.createdAt",
				"properties.updatedAt",
				"properties.countDetails",
				"properties.messageCount",
				"properties.sizeInBytes",
			},
		},
	}
}

func init() { azwise.Register(NewServiceBusQueue()) }
