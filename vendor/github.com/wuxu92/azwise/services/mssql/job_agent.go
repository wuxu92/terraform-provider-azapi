package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// JobAgent provides resource knowledge for Microsoft.Sql/servers/jobAgents.
//
// Contributing Terraform resource: azurerm_mssql_job_agent.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mssql/mssql_job_agent_resource.go
//     (schema L50-79, Create body params L108-118)
//   - terraform-provider-azurerm internal/services/mssql/validate/mssql.go
//     (ValidateMsSqlJobAgentName L50-53)
//   - terraform-provider-azurerm internal/services/mssql/helper/constants.go
//     (PossibleValuesForJobAgentSku JA100/JA200/JA400/JA800)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/jobagents:
//     model_jobagentproperties.go, model_sku.go
//
// Notes:
//   - name/database_id are envelope/parent references; not emitted as body rules.
//   - database_id maps to properties.databaseId and is ForceNew (child of the database),
//     but is a required parent linkage rather than a mutable body property.
type JobAgent struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*JobAgent)(nil)

// NewJobAgent returns knowledge for the jobAgents resource.
func NewJobAgent() *JobAgent {
	return &JobAgent{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/servers/jobAgents",
			ApiVersions:  []string{"2023-08-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.databaseId"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[^?<>*%&:\/?]{0,127}[^?<>*%&:\/?. ]$`,
					Message:      "job agent name must not contain any of ?<>*%&:\\/?, must not end with a space or a period and can't have more than 128 characters",
				},
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"JA100", "JA200", "JA400", "JA800"},
					Message:       "sku must be one of JA100, JA200, JA400 or JA800",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "sku.name", Value: "JA100"},
			},
			RequiredFields: []string{
				"properties.databaseId",
			},
		},
	}
}

func init() { azwise.Register(NewJobAgent()) }
