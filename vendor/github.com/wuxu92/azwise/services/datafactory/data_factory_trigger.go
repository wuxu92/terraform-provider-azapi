package datafactory

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DataFactoryTrigger provides resource knowledge for
// Microsoft.DataFactory/factories/triggers.
//
// This ARM type is split by AzureRM into four typed Terraform resources,
// discriminated by the ARM body's properties.type field:
//   - azurerm_data_factory_trigger_schedule       -> properties.type = "ScheduleTrigger"
//   - azurerm_data_factory_trigger_blob_event     -> properties.type = "BlobEventsTrigger"
//   - azurerm_data_factory_trigger_custom_event   -> properties.type = "CustomEventsTrigger"
//   - azurerm_data_factory_trigger_tumbling_window -> properties.type = "TumblingWindowTrigger"
//
// Only knowledge that is universal across all four contributors is unioned here
// (see azwise "one ARM type, many resources" rules). Type-specific value
// constraints live inside the discriminated properties.* union, whose paths do
// not resolve through the SDK's interface-typed Properties field, so they are
// documented rather than emitted.
//
// Sources:
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_trigger_schedule_resource.go:26-246
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_trigger_blob_event_resource.go:24-149
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_trigger_custom_event_resource.go:44-83
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_trigger_tumbling_window_resource.go:44-90
//   - terraform-provider-azurerm internal/services/datafactory/validate/datafactory.go:14-23
//     (DataFactoryPipelineAndTriggerName regex — used by all four for `name`)
//   - kermit SDK datafactory/2018-06-01/datafactory Trigger discriminated union
//     (BlobEventsTrigger / CustomEventsTrigger / ScheduleTrigger / TumblingWindowTrigger)
//
// Intentionally skipped here:
//   - name: resource-name attribute (validated via StringRule with empty PropertyPath),
//     universal across all four triggers.
//   - data_factory_id: parent reference (AzAPI ID segment), not a body property.
//   - Type-specific fields, all inside the discriminated union under properties.* and
//     therefore non-resolving through the interface Properties field:
//       * ScheduleTrigger: frequency (RecurrenceFrequency enum, default Minute),
//         interval (IntAtLeast(1), default 1), recurrence schedule blocks, activated.
//       * BlobEventsTrigger: events StringInSlice{Microsoft.Storage.BlobCreated,
//         Microsoft.Storage.BlobDeleted}, storage_account_id (ForceNew),
//         blob_path_begins_with/ends_with AtLeastOneOf, ignore_empty_blobs.
//       * CustomEventsTrigger: events, subject filters, eventgrid_topic_id.
//       * TumblingWindowTrigger: frequency StringInSlice{Hour,Minute} (ForceNew),
//         interval, delay, retry, dependencies.
//     These do not corrupt other kinds and cannot be expressed as universal rules.
type DataFactoryTrigger struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DataFactoryTrigger)(nil)

// NewDataFactoryTrigger returns merged knowledge for the
// Microsoft.DataFactory/factories/triggers resource.
func NewDataFactoryTrigger() *DataFactoryTrigger {
	return &DataFactoryTrigger{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DataFactory/factories/triggers",
			ApiVersions:  []string{"2018-06-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). All four triggers use
					// DataFactoryPipelineAndTriggerName — universal, safe to union.
					Regex:   `^[A-Za-z0-9_][^<>*#.%&:\\+?/]*$`,
					Message: "invalid Data Factory trigger name, see https://docs.microsoft.com/en-us/azure/data-factory/naming-rules",
				},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDataFactoryTrigger()) }
