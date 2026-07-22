package synapse

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SynapseSqlPoolWorkloadClassifier provides resource knowledge for
// Microsoft.Synapse/workspaces/sqlPools/workloadGroups/workloadClassifiers.
//
// Mirrors azurerm_synapse_sql_pool_workload_classifier.
//
// Sources:
//   - internal/services/synapse/synapse_sql_pool_workload_classifier_resource.go
//     (schema 43-104: name ForceNew; member_name Required; context; end_time/start_time
//     StringMatch ^\d{2}:\d{2}$ (HH:MM UTC); importance enum
//     (low/below_normal/normal/above_normal/high); label; timeouts Create/Update/Delete 30m
//     Read 5m; create 134-142 → WorkloadClassifierProperties).
//   - go-azure-sdk resource-manager (track1) synapse model WorkloadClassifierProperties
//     (memberName/label/context/startTime/endTime/importance json tags), resourceids.go
//     SqlPoolWorkloadClassifier (segment casing "workloadClassifiers").
//
// Notes:
//   - start_time / end_time carry an HH:MM regex; both map to properties.startTime /
//     properties.endTime and are emitted as regex StringRules.
type SynapseSqlPoolWorkloadClassifier struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SynapseSqlPoolWorkloadClassifier)(nil)

func NewSynapseSqlPoolWorkloadClassifier() *SynapseSqlPoolWorkloadClassifier {
	return &SynapseSqlPoolWorkloadClassifier{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Synapse/workspaces/sqlPools/workloadGroups/workloadClassifiers",
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
					PropertyPath:  "properties.importance",
					AllowedValues: []string{"low", "below_normal", "normal", "above_normal", "high"},
					Message:       "must be one of low, below_normal, normal, above_normal or high",
				},
				{
					PropertyPath: "properties.startTime",
					Regex:        `^\d{2}:\d{2}$`,
					Message:      "must be in HH:MM (UTC) format",
				},
				{
					PropertyPath: "properties.endTime",
					Regex:        `^\d{2}:\d{2}$`,
					Message:      "must be in HH:MM (UTC) format",
				},
			},
			RequiredFields: []string{
				"properties.memberName",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewSynapseSqlPoolWorkloadClassifier()) }
