package eventhub

import (
	"time"

	"github.com/wuxu92/azwise"
)

// EventHub provides resource knowledge for Microsoft.EventHub/namespaces/eventhubs.
//
// Contributing Terraform resource: azurerm_eventhub.
//
// Sources:
//   - terraform-provider-azurerm internal/services/eventhub/eventhub_resource.go
//     (schema L29-259, Create body L261-325, expand capture/retention L460-544)
//   - terraform-provider-azurerm internal/services/eventhub/validate/eventhub_names.go
//     (ValidateEventHubName L21-26), eventhub_partition.go (1-1024),
//     eventhub_message_retention.go (1-90), eventhub_archive.go (archive name format)
//   - go-azure-sdk resource-manager/eventhub/2024-01-01/eventhubs:
//     model_eventhubproperties.go, model_capturedescription.go, model_destination.go,
//     model_destinationproperties.go, model_captureidentity.go,
//     model_retentiondescription.go, constants.go (enums)
//
// Notes:
//   - name/namespace_id are envelope / parent-reference fields; not emitted as body rules.
//   - status uses the full EntityStatus SDK enum (AzAPI sends raw ARM values); AzureRM
//     restricts the settable subset to Active/Disabled/SendDisabled.
//   - storage_authentication_type maps to
//     properties.captureDescription.destination.identity.type only for
//     SystemAssigned/UserAssigned; the AzureRM default "StorageSAS" means "no identity"
//     and has no ARM identity.type value, so it is not emitted as a default.
//   - CustomizeDiff (eventhub_resource.go L218-227) requires storage_authentication_id
//     (destination.identity.userAssignedIdentity) when identity.type == "UserAssigned";
//     a value-conditional requirement not expressible as a declarative RequiredWith.
type EventHub struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*EventHub)(nil)

// NewEventHub returns knowledge for the namespaces/eventhubs resource.
func NewEventHub() *EventHub {
	return &EventHub{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.EventHub/namespaces/eventhubs",
			ApiVersions:  []string{"2024-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.retentionDescription.cleanupPolicy"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[a-zA-Z0-9]([-._a-zA-Z0-9]{0,254}[a-zA-Z0-9])?$`,
					Message:      "name may contain only letters, numbers, periods, hyphens and underscores, up to 256 characters, beginning and ending with a letter or number",
				},
				{
					PropertyPath:  "properties.status",
					AllowedValues: []string{"Active", "Creating", "Deleting", "Disabled", "ReceiveDisabled", "Renaming", "Restoring", "SendDisabled", "Unknown"},
					Message:       "status must be a valid EntityStatus value",
				},
				{
					PropertyPath:  "properties.captureDescription.encoding",
					AllowedValues: []string{"Avro", "AvroDeflate"},
					Message:       "encoding must be Avro or AvroDeflate",
				},
				{
					PropertyPath:  "properties.retentionDescription.cleanupPolicy",
					AllowedValues: []string{"Compact", "Delete"},
					Message:       "cleanup_policy must be Compact or Delete",
				},
				{
					PropertyPath:  "properties.captureDescription.destination.identity.type",
					AllowedValues: []string{"SystemAssigned", "UserAssigned"},
					Message:       "storage_authentication_type identity must be SystemAssigned or UserAssigned",
				},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.partitionCount", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(1024))},
				{PropertyPath: "properties.messageRetentionInDays", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(90))},
				{PropertyPath: "properties.captureDescription.intervalInSeconds", MinValue: azwise.Ptr(int64(60)), MaxValue: azwise.Ptr(int64(900))},
				{PropertyPath: "properties.captureDescription.sizeLimitInBytes", MinValue: azwise.Ptr(int64(10485760)), MaxValue: azwise.Ptr(int64(524288000))},
			},
			ExactlyOneOf: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.messageRetentionInDays", "properties.retentionDescription"},
					Message: "exactly one of message_retention or retention_description must be set",
				},
				{
					Paths:   []string{"properties.retentionDescription.retentionTimeInHours", "properties.retentionDescription.tombstoneRetentionTimeInHours"},
					Message: "exactly one of retention_time_in_hours or tombstone_retention_time_in_hours must be set",
				},
			},
			ComputedFields: []string{
				"properties.partitionIds",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.status", Value: "Active"},
				{PropertyPath: "properties.captureDescription.skipEmptyArchives", Value: false},
				{PropertyPath: "properties.captureDescription.intervalInSeconds", Value: 300},
				{PropertyPath: "properties.captureDescription.sizeLimitInBytes", Value: 314572800},
			},
			RequiredFields: []string{
				"properties.partitionCount",
			},
		},
	}
}

func init() { azwise.Register(NewEventHub()) }
