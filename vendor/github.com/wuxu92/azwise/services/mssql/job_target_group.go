package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// JobTargetGroup provides resource knowledge for
// Microsoft.Sql/servers/jobAgents/targetGroups.
//
// Contributing Terraform resource: azurerm_mssql_job_target_group.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mssql/mssql_job_target_group_resource.go
//     (schema L44-97, Create body L160-165, CustomizeDiff L103-121)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/jobtargetgroups:
//     model_jobtargetgroupproperties.go, model_jobtarget.go, constants.go
//     (JobTargetGroupMembershipType Include/Exclude, JobTargetType)
//
// Notes:
//   - name/job_agent_id are envelope/parent references; not emitted as body rules.
//   - job_target is an array (properties.members[*]); its per-element constraints
//     (membership_type Include/Exclude enum, type) are array-element paths that azwise
//     declarative rules do not support, so no rule is emitted for them.
//   - The database_name/elastic_pool_name mutual-exclusion is a per-array-element
//     CustomizeDiff and is likewise not expressible declaratively.
type JobTargetGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*JobTargetGroup)(nil)

// NewJobTargetGroup returns knowledge for the targetGroups resource.
func NewJobTargetGroup() *JobTargetGroup {
	return &JobTargetGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/servers/jobAgents/targetGroups",
			ApiVersions:  []string{"2023-08-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewJobTargetGroup()) }
