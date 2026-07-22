package synapse

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SynapseSqlPoolWorkloadGroup provides resource knowledge for
// Microsoft.Synapse/workspaces/sqlPools/workloadGroups.
//
// Mirrors azurerm_synapse_sql_pool_workload_group.
//
// Sources:
//   - internal/services/synapse/synapse_sql_pool_workload_group_resource.go
//     (schema 42-93: name ForceNew; max_resource_percent Required IntBetween(1,100);
//     min_resource_percent Required IntBetween(0,100); importance Default "normal";
//     max_resource_percent_per_request FloatBetween(0,100) Default 3;
//     min_resource_percent_per_request FloatBetween(0,100);
//     query_execution_timeout_in_seconds IntAtLeast(0); timeouts Create/Update/Delete 30m
//     Read 5m; create 122-134 → WorkloadGroupProperties).
//   - go-azure-sdk resource-manager (track1) synapse model WorkloadGroupProperties
//     (minResourcePercent/maxResourcePercent *int32, minResourcePercentPerRequest/
//     maxResourcePercentPerRequest *float64, importance/queryExecutionTimeout json tags),
//     resourceids.go SqlPoolWorkloadGroup (segment casing "workloadGroups").
type SynapseSqlPoolWorkloadGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SynapseSqlPoolWorkloadGroup)(nil)

func NewSynapseSqlPoolWorkloadGroup() *SynapseSqlPoolWorkloadGroup {
	return &SynapseSqlPoolWorkloadGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Synapse/workspaces/sqlPools/workloadGroups",
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
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.maxResourcePercent",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(100)),
				},
				{
					PropertyPath: "properties.minResourcePercent",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(100)),
				},
				{
					PropertyPath: "properties.queryExecutionTimeout",
					MinValue:     azwise.Ptr(int64(0)),
				},
			},
			FloatRules: []azwise.FloatRule{
				{
					PropertyPath: "properties.maxResourcePercentPerRequest",
					MinValue:     azwise.Ptr(float64(0)),
					MaxValue:     azwise.Ptr(float64(100)),
				},
				{
					PropertyPath: "properties.minResourcePercentPerRequest",
					MinValue:     azwise.Ptr(float64(0)),
					MaxValue:     azwise.Ptr(float64(100)),
				},
			},
			RequiredFields: []string{
				"properties.maxResourcePercent",
				"properties.minResourcePercent",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.importance", Value: "normal"},
				{PropertyPath: "properties.maxResourcePercentPerRequest", Value: float64(3)},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewSynapseSqlPoolWorkloadGroup()) }
