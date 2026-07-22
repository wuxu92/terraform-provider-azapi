package synapse

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SynapseSparkPool provides resource knowledge for
// Microsoft.Synapse/workspaces/bigDataPools.
//
// Mirrors azurerm_synapse_spark_pool.
//
// Sources:
//   - internal/services/synapse/synapse_spark_pool_resource.go
//     (schema 45-228: name ForceNew (SparkPoolName validator); node_size_family Required
//     enum; node_size Required enum; spark_version Required enum (3.4/3.5, +3.2/3.3 pre-5.0);
//     min_executors/max_executors IntBetween(0,200); node_count O+C IntBetween(3,200)
//     ExactlyOneOf auto_scale; auto_scale.min/max_node_count IntBetween(3,200);
//     auto_pause.delay_in_minutes IntBetween(5,10080); compute_isolation_enabled Default
//     false; dynamic_executor_allocation_enabled Default false; session_level_packages_enabled
//     Default false; spark_events_folder Default "/events"; spark_log_folder Default "/logs";
//     timeouts Create/Update/Delete 30m Read 5m; create 285-306 →
//     BigDataPoolResourceProperties).
//   - internal/services/synapse/validate/spark_pool_name.go
//     (regex ^[a-zA-Z][a-zA-Z\d]{0,14}$, 1-15 chars).
//   - go-azure-sdk resource-manager (track1) synapse models BigDataPoolResourceProperties /
//     AutoScaleProperties / AutoPauseProperties / DynamicExecutorAllocation (nodeSize/
//     nodeSizeFamily/nodeCount/cacheSize/autoScale/autoPause/dynamicExecutorAllocation/
//     sparkVersion/defaultSparkLogFolder/sparkEventsFolder json tags), enums.go NodeSize
//     (None/Small/Medium/Large/XLarge/XXLarge/XXXLarge) + NodeSizeFamily
//     (None/MemoryOptimized/HardwareAcceleratedFPGA/HardwareAcceleratedGPU);
//     resourceids.go SparkPool (segment casing "bigDataPools").
//
// Notes:
//   - spark_log_folder maps to properties.defaultSparkLogFolder (not sparkLogFolder).
//   - node_count vs auto_scale is an ExactlyOneOf, but node_count maps to
//     properties.nodeCount while auto_scale maps to properties.autoScale (with
//     properties.autoScale.enabled discriminating); emitted as an ExactlyOneOf relational
//     rule over those two paths.
//   - dynamic_executor_allocation min/max_executors map to
//     properties.dynamicExecutorAllocation.minExecutors/maxExecutors; auto_scale
//     min/max_node_count map to properties.autoScale.minNodeCount/maxNodeCount;
//     auto_pause delay maps to properties.autoPause.delayInMinutes.
//   - NodeSize/NodeSizeFamily enums use the full ARM SDK sets.
type SynapseSparkPool struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SynapseSparkPool)(nil)

func NewSynapseSparkPool() *SynapseSparkPool {
	return &SynapseSparkPool{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Synapse/workspaces/bigDataPools",
			ApiVersions:  []string{"2021-06-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			StringRules: []azwise.StringRule{
				{
					// validate.SparkPoolName
					Regex:     `^[a-zA-Z][a-zA-Z\d]{0,14}$`,
					MinLength: 1,
					MaxLength: 15,
					Message:   "must be 1-15 chars, start with a letter, and contain only letters or numbers",
				},
				{
					PropertyPath:  "properties.nodeSize",
					AllowedValues: []string{"None", "Small", "Medium", "Large", "XLarge", "XXLarge", "XXXLarge"},
					Message:       "must be one of None, Small, Medium, Large, XLarge, XXLarge or XXXLarge",
				},
				{
					PropertyPath:  "properties.nodeSizeFamily",
					AllowedValues: []string{"None", "MemoryOptimized", "HardwareAcceleratedFPGA", "HardwareAcceleratedGPU"},
					Message:       "must be one of None, MemoryOptimized, HardwareAcceleratedFPGA or HardwareAcceleratedGPU",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.nodeCount",
					MinValue:     azwise.Ptr(int64(3)),
					MaxValue:     azwise.Ptr(int64(200)),
				},
				{
					PropertyPath: "properties.dynamicExecutorAllocation.minExecutors",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(200)),
				},
				{
					PropertyPath: "properties.dynamicExecutorAllocation.maxExecutors",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(200)),
				},
				{
					PropertyPath: "properties.autoScale.minNodeCount",
					MinValue:     azwise.Ptr(int64(3)),
					MaxValue:     azwise.Ptr(int64(200)),
				},
				{
					PropertyPath: "properties.autoScale.maxNodeCount",
					MinValue:     azwise.Ptr(int64(3)),
					MaxValue:     azwise.Ptr(int64(200)),
				},
				{
					PropertyPath: "properties.autoPause.delayInMinutes",
					MinValue:     azwise.Ptr(int64(5)),
					MaxValue:     azwise.Ptr(int64(10080)),
				},
			},
			ExactlyOneOf: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.nodeCount", "properties.autoScale"},
					Message: "exactly one of nodeCount or autoScale must be set",
				},
			},
			RequiredFields: []string{
				"properties.nodeSize",
				"properties.nodeSizeFamily",
				"properties.sparkVersion",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.isComputeIsolationEnabled", Value: false},
				{PropertyPath: "properties.sessionLevelPackagesEnabled", Value: false},
				{PropertyPath: "properties.sparkEventsFolder", Value: "/events"},
				{PropertyPath: "properties.defaultSparkLogFolder", Value: "/logs"},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.creationDate",
				"properties.lastSucceededTimestamp",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewSynapseSparkPool()) }
