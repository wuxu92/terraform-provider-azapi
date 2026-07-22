package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Job provides resource knowledge for Microsoft.Sql/servers/jobAgents/jobs.
//
// Contributing Terraform resources:
//   - azurerm_mssql_job     (the job body: properties.description)
//   - azurerm_mssql_job_schedule (writes properties.schedule.* on the SAME jobs resource)
//
// Sources:
//   - terraform-provider-azurerm internal/services/mssql/mssql_job_resource.go
//     (schema L29-48, Create body L91-96)
//   - terraform-provider-azurerm internal/services/mssql/mssql_job_schedule_resource.go
//     (schema L38-79, Create body L160-171, CustomizeDiff L85-101)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/jobs:
//     model_jobproperties.go, model_jobschedule.go, constants.go (JobScheduleType)
//
// Notes:
//   - name/job_agent_id are envelope/parent references; not emitted as body rules.
//   - azurerm_mssql_job_schedule is not a distinct ARM resource: it mutates the
//     properties.schedule sub-object of this same jobs resource. Its universal value
//     constraint (schedule.type enum) is unioned here; its `interval`-required-when-
//     Recurring rule is a cross-field CustomizeDiff that azwise cannot express
//     declaratively and is intentionally not emitted.
//   - schedule.enabled/start_time/end_time are Optional+Computed (server-defaulted);
//     no explicit default value is emitted.
type Job struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Job)(nil)

// NewJob returns knowledge for the jobs resource.
func NewJob() *Job {
	return &Job{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/servers/jobAgents/jobs",
			ApiVersions:  []string{"2023-08-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.schedule.type",
					AllowedValues: []string{"Once", "Recurring"},
					Message:       "schedule type must be Once or Recurring",
				},
			},
		},
	}
}

func init() { azwise.Register(NewJob()) }
