package storagemover

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StorageMoverJobDefinition provides resource knowledge for
// Microsoft.StorageMover/storageMovers/projects/jobDefinitions.
//
// Mirrors azurerm_storage_mover_job_definition.
//
// Sources:
//   - internal/services/storagemover/storage_mover_job_definition_resource.go
//     (Arguments 59-127: name Required ForceNew StringMatch ^[0-9a-zA-Z][-_0-9a-zA-Z]{0,63}$;
//     storage_mover_project_id ForceNew parent ref; source_name/target_name Required ForceNew;
//     copy_mode Required StringInSlice(Mirror,Additive); source_sub_path/target_sub_path
//     Optional ForceNew; agent_name Optional; description Optional; create 161-183 →
//     JobDefinition{Properties{CopyMode, SourceName, TargetName, AgentName, Description,
//     SourceSubpath, TargetSubpath}}; timeouts Create 30m).
//   - go-azure-sdk resource-manager/storagemover/2025-07-01/jobdefinitions:
//     model_jobdefinitionproperties.go (copyMode/sourceName/targetName required for create;
//     agentName/sourceSubpath/targetSubpath/description settable; agentResourceId/jobType/
//     latestJobRun*/provisioningState/sourceResourceId/targetResourceId/sourceTargetMap read-only),
//     constants.go (CopyMode Additive/Mirror),
//     id_jobdefinition.go (type segment casing "storageMovers/projects/jobDefinitions").
type StorageMoverJobDefinition struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StorageMoverJobDefinition)(nil)

// NewStorageMoverJobDefinition returns knowledge for the
// storageMovers/projects/jobDefinitions resource.
func NewStorageMoverJobDefinition() *StorageMoverJobDefinition {
	return &StorageMoverJobDefinition{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StorageMover/storageMovers/projects/jobDefinitions",
			ApiVersions:  []string{"2025-07-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.sourceName"},
				{PropertyPath: "properties.targetName"},
				{PropertyPath: "properties.sourceSubpath"},
				{PropertyPath: "properties.targetSubpath"},
			},
			StringRules: []azwise.StringRule{
				{
					// name: validation.StringMatch.
					Regex:     `^[0-9a-zA-Z][-_0-9a-zA-Z]{0,63}$`,
					MinLength: 1,
					MaxLength: 64,
					Message:   "must be 1-64 characters, begin with a letter or number, and contain only letters, numbers, dashes and underscores",
				},
				{
					// copy_mode: validation.StringInSlice(Mirror, Additive).
					PropertyPath:  "properties.copyMode",
					AllowedValues: []string{"Additive", "Mirror"},
				},
			},
			RequiredFields: []string{
				"properties.copyMode",
				"properties.sourceName",
				"properties.targetName",
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.agentResourceId",
				"properties.jobType",
				"properties.latestJobRunName",
				"properties.latestJobRunResourceId",
				"properties.latestJobRunStatus",
				"properties.sourceResourceId",
				"properties.targetResourceId",
				"properties.sourceTargetMap",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewStorageMoverJobDefinition()) }
