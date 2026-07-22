package monitor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DataCollectionRule provides resource knowledge for
// Microsoft.Insights/dataCollectionRules.
//
// Mirrors azurerm_monitor_data_collection_rule.
//
// Sources:
//   - terraform-provider-azurerm internal/services/monitor/monitor_data_collection_rule_resource.go
//     Arguments (205-909): name Required+ForceNew (StringIsNotEmpty); resource_group_name +
//     location ForceNew (commonschema); data_collection_endpoint_id Optional (DCE id);
//     data_flow Required MinItems 1 → properties.dataFlows; destinations Required MaxItems/
//     MinItems 1 → properties.destinations (AtLeastOneOf across metrics/event_hub/
//     event_hub_direct/log_analytics/monitor_account/storage_blob/storage_blob_direct/
//     storage_table_direct); description Optional; identity SystemOrUserAssigned; kind
//     Optional enum Linux/Windows/AgentDirectToStore/WorkspaceTransforms (top-level ARM kind);
//     stream_declaration Optional. Attributes (912-919): immutable_id Computed. Create
//     (933-994) / Delete (1147-1164) / Update (1064-1145) 30m, Read (996-1062) 5m.
//   - go-azure-sdk resource-manager/insights/2023-03-11/datacollectionrules:
//     DataCollectionRuleResource.Kind is top-level (json:"kind"); DataCollectionRule
//     (model_datacollectionrule.go) fields immutableId/provisioningState/metadata/endpoints
//     read-only; id_datacollectionrule.go type Microsoft.Insights/dataCollectionRules.
//
// Note: deeply nested array-element enums (data_source.log_file.settings.text.format,
// syslog facility_names/log_levels, prometheus_forwarder streams, stream_declaration
// column.type) are array-element / map-key paths and are intentionally not emitted as
// declarative rules (azwise skips properties.foo[*].bar paths).
type DataCollectionRule struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DataCollectionRule)(nil)

// NewDataCollectionRule returns knowledge for the dataCollectionRules resource.
func NewDataCollectionRule() *DataCollectionRule {
	return &DataCollectionRule{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Insights/dataCollectionRules",
			ApiVersions:  []string{"2023-03-11"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name: StringIsNotEmpty.
					MinLength: 1,
					Message:   "name must not be empty",
				},
				{
					// kind is a top-level ARM property (not under properties).
					PropertyPath:  "kind",
					AllowedValues: []string{"Linux", "Windows", "AgentDirectToStore", "WorkspaceTransforms"},
					Message:       "kind must be one of Linux, Windows, AgentDirectToStore, WorkspaceTransforms",
				},
			},
			// data_flow and destinations are Required (MinItems 1).
			RequiredFields: []string{
				"properties.dataFlows",
				"properties.destinations",
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.immutableId",
				"properties.provisioningState",
				"properties.metadata",
				"properties.endpoints",
			},
		},
	}
}

func init() { azwise.Register(NewDataCollectionRule()) }
