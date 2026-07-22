package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// JobStep provides resource knowledge for Microsoft.Sql/servers/jobAgents/jobs/steps.
//
// Contributing Terraform resource: azurerm_mssql_job_step.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mssql/mssql_job_step_resource.go
//     (schema L52-146, Create body L216-239, CustomizeDiff L153-177)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/jobsteps:
//     model_jobstepproperties.go, model_jobstepexecutionoptions.go, model_jobstepaction.go
//
// Notes:
//   - name/job_id are envelope/parent references; not emitted as body rules.
//   - job_step_index maps to properties.stepId (>= 1). sql_script maps to
//     properties.action.value (StringIsNotEmpty, semantic; no declarative rule).
//   - retry_attempts validates 1..math.MaxInt32; MaxInt32 (2147483647) is the encoded
//     upper bound. maximum_retry_interval_seconds > initial_retry_interval_seconds is a
//     cross-field CustomizeDiff azwise cannot express and is not emitted.
//   - job_credential_id is conditionally ForceNew (only when cleared, per CustomizeDiff);
//     it is a resource reference (properties.credential) rather than a value-constrained
//     field, so no unconditional ForceNew rule is emitted.
type JobStep struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*JobStep)(nil)

// NewJobStep returns knowledge for the job steps resource.
func NewJobStep() *JobStep {
	return &JobStep{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/servers/jobAgents/jobs/steps",
			ApiVersions:  []string{"2023-08-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.stepId", MinValue: azwise.Ptr(int64(1))},
				{PropertyPath: "properties.executionOptions.initialRetryIntervalSeconds", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(2147483))},
				{PropertyPath: "properties.executionOptions.maximumRetryIntervalSeconds", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(2147483))},
				{PropertyPath: "properties.executionOptions.retryAttempts", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(2147483647))},
				{PropertyPath: "properties.executionOptions.timeoutSeconds", MinValue: azwise.Ptr(int64(1)), MaxValue: azwise.Ptr(int64(2147483))},
			},
			FloatRules: []azwise.FloatRule{
				{PropertyPath: "properties.executionOptions.retryIntervalBackoffMultiplier", MinValue: azwise.Ptr(float64(1))},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.executionOptions.initialRetryIntervalSeconds", Value: 1},
				{PropertyPath: "properties.executionOptions.maximumRetryIntervalSeconds", Value: 120},
				{PropertyPath: "properties.executionOptions.retryAttempts", Value: 10},
				{PropertyPath: "properties.executionOptions.retryIntervalBackoffMultiplier", Value: 2.0},
				{PropertyPath: "properties.executionOptions.timeoutSeconds", Value: 43200},
			},
			RequiredFields: []string{
				"properties.action.value",
				"properties.stepId",
				"properties.targetGroup",
			},
		},
	}
}

func init() { azwise.Register(NewJobStep()) }
