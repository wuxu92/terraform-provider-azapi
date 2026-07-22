// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package servicebus

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ServiceBusTopic provides resource knowledge for Microsoft.ServiceBus/namespaces/topics.
//
// Contributing Terraform resource: azurerm_servicebus_topic.
//
// Scoping verified via SDK id parser: TopicId.ID() formats
// .../Microsoft.ServiceBus/namespaces/%s/topics/%s (topics/id_topic.go L110-111).
//
// Sources:
//   - terraform-provider-azurerm internal/services/servicebus/servicebus_topic_resource.go
//     (schema L46-136, Create body L172-184)
//   - terraform-provider-azurerm internal/services/servicebus/validate/topic_name.go (L13-17),
//     validate/servicebus.go (ServiceBusMaxMessageSizeInKilobytes IntBetween 1024-102400 L25-27)
//   - go-azure-sdk resource-manager/servicebus/2024-01-01/topics:
//     model_sbtopicproperties.go, constants.go (EntityStatus L14-24)
//
// Notes:
//   - name/namespace_id are envelope / parent-reference fields; not emitted as body rules.
//   - AzureRM topic status ValidateFunc only accepts Active/Disabled, but ARM accepts the full
//     EntityStatus set; per azwise policy we emit the full ARM enum (AzAPI sends raw values).
//   - max_size_in_megabytes uses IntInSlice{1024..81920} (discrete, non-contiguous) which
//     IntRule (min/max) cannot express; documented, not emitted.
//   - max_message_size_in_kilobytes is Optional+Computed with a sku-dependent server default;
//     represented as a nil-valued DefaultValue.
//   - partitioning_enabled and requires_duplicate_detection are ForceNew.
type ServiceBusTopic struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ServiceBusTopic)(nil)

// NewServiceBusTopic returns knowledge for the namespaces/topics resource.
func NewServiceBusTopic() *ServiceBusTopic {
	return &ServiceBusTopic{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ServiceBus/namespaces/topics",
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
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "", // resource name
					Regex:        "^[a-zA-Z0-9]([-._~/a-zA-Z0-9]{0,258}[a-zA-Z0-9])?$",
					Message:      "topic name can contain only letters, numbers, periods, hyphens, tildes, forward slashes and underscores, must start and end with a letter or number, and be less than 260 characters long",
				},
				{
					PropertyPath:  "properties.status",
					AllowedValues: []string{"Active", "Creating", "Deleting", "Disabled", "ReceiveDisabled", "Renaming", "Restoring", "SendDisabled", "Unknown"},
					Message:       "status must be a valid EntityStatus value",
				},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.maxMessageSizeInKilobytes", MinValue: azwise.Ptr(int64(1024)), MaxValue: azwise.Ptr(int64(102400)), Message: "max_message_size_in_kilobytes must be between 1024 and 102400"},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.autoDeleteOnIdle", Value: "P10675199DT2H48M5.4775807S"},
				{PropertyPath: "properties.defaultMessageTimeToLive", Value: "P10675199DT2H48M5.4775807S"},
				{PropertyPath: "properties.duplicateDetectionHistoryTimeWindow", Value: "PT10M"},
				{PropertyPath: "properties.status", Value: "Active"},
				// sku-dependent server default (no static AzureRM default).
				{PropertyPath: "properties.maxMessageSizeInKilobytes", Value: nil},
				{PropertyPath: "properties.maxSizeInMegabytes", Value: nil},
			},
			ComputedFields: []string{
				"properties.accessedAt",
				"properties.createdAt",
				"properties.updatedAt",
				"properties.countDetails",
				"properties.sizeInBytes",
				"properties.subscriptionCount",
			},
		},
	}
}

func init() { azwise.Register(NewServiceBusTopic()) }
