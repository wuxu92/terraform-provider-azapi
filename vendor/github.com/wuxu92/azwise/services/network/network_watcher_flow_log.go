package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkWatcherFlowLog provides resource knowledge for
// Microsoft.Network/networkWatchers/flowLogs.
//
// Mirrors azurerm_network_watcher_flow_log. name, network_watcher_name and
// resource_group are envelope-owned; location is a top-level ARM property that
// forces replacement.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_watcher_flow_log_resource.go
//     (schema, expandNetworkWatcherFlowLog*, CRUD timeouts 30/5/30/30)
//   - go-azure-sdk resource-manager/network/2025-01-01/flowlogs:
//     model_flowlogpropertiesformat.go, model_flowlogformatparameters.go,
//     id_flowlog.go (segment casing "networkWatchers/flowLogs")
//
// Not encoded (deliberate):
//   - traffic_analytics.interval_in_minutes uses IntInSlice{10,60} (a discrete
//     set, not a contiguous range) so it cannot be expressed as an IntRule
//     min/max; skipped.
//   - traffic_analytics.workspace_id uses validation.IsUUID — a generic semantic
//     validator. It maps to
//     properties.flowAnalyticsConfiguration.networkWatcherFlowAnalyticsConfiguration.workspaceId
//     and should be attached as a UUID validator in the resource customizer, not
//     declared here.
//   - retention_policy (properties.retentionPolicy) and traffic_analytics
//     (properties.flowAnalyticsConfiguration) are single nested objects with no
//     further declarative value constraints beyond the above.
type NetworkWatcherFlowLog struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkWatcherFlowLog)(nil)

// NewNetworkWatcherFlowLog returns knowledge for the networkWatchers/flowLogs resource.
func NewNetworkWatcherFlowLog() *NetworkWatcherFlowLog {
	return &NetworkWatcherFlowLog{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkWatchers/flowLogs",
			ApiVersions:  []string{"2025-01-01"},
			// location is ForceNew on the flow log; name/network_watcher are envelope.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.format.version",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(2)),
					Message:      "flow log format version must be 1 or 2",
				},
			},
			// target_resource_id, storage_account_id and enabled are Required.
			RequiredFields: []string{
				"properties.targetResourceId",
				"properties.storageId",
				"properties.enabled",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.format.version", Value: float64(1)},
			},
		},
	}
}

func init() { azwise.Register(NewNetworkWatcherFlowLog()) }
