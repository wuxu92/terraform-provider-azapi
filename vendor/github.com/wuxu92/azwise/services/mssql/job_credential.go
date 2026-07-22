package mssql

import (
	"time"

	"github.com/wuxu92/azwise"
)

// JobCredential provides resource knowledge for Microsoft.Sql/servers/jobAgents/credentials.
//
// Contributing Terraform resource: azurerm_mssql_job_credential.
//
// Sources:
//   - terraform-provider-azurerm internal/services/mssql/mssql_job_credential_resource.go
//     (schema L41-80, Create body L116-122)
//   - go-azure-sdk resource-manager/sql/2023-08-01-preview/jobcredentials:
//     model_jobcredentialproperties.go
//
// Notes:
//   - name/job_agent_id are envelope/parent references; not emitted as body rules.
//   - password/password_wo are two representations of a single ARM property
//     (properties.password); the ExactlyOneOf/ConflictsWith between them is a
//     provider-side write-only-vs-plaintext concern with no distinct ARM field,
//     so it is not emitted as a relational rule.
//   - properties.username and properties.password are Required (non-pointer) in the SDK.
type JobCredential struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*JobCredential)(nil)

// NewJobCredential returns knowledge for the credentials resource.
func NewJobCredential() *JobCredential {
	return &JobCredential{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Sql/servers/jobAgents/credentials",
			ApiVersions:  []string{"2023-08-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			SensitiveFields: []string{
				"properties.password",
			},
			RequiredFields: []string{
				"properties.username",
				"properties.password",
			},
		},
	}
}

func init() { azwise.Register(NewJobCredential()) }
