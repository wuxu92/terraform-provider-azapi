package cognitive

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CognitiveAccountProject provides resource knowledge for
// Microsoft.CognitiveServices/accounts/projects.
//
// Contributing TF resource:
//   - azurerm_cognitive_account_project — cognitive_account_project_resource.go
//
// Sources:
//   - AzureRM cognitive_account_project_resource.go schema + CRUD
//   - AzureRM validate/account_project_name.go (name regex)
//   - Azure SDK cognitive/2026-03-01/cognitiveservicesprojects models
type CognitiveAccountProject struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CognitiveAccountProject)(nil)

// NewCognitiveAccountProject returns a CognitiveAccountProject knowledge instance.
func NewCognitiveAccountProject() *CognitiveAccountProject {
	return &CognitiveAccountProject{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.CognitiveServices/accounts/projects",
			ApiVersions:  []string{"2026-03-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				// description and display_name are conditionally ForceNew (only when
				// changing from a non-empty value to an empty one, an Azure API
				// limitation). This is not representable declaratively, so it is
				// documented but not listed.
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ──
				// validate.AccountProjectName(): ^[a-zA-Z0-9][a-zA-Z0-9_.-]{1,63}$
				{
					Regex:     `^[a-zA-Z0-9][a-zA-Z0-9_.-]{1,63}$`,
					MinLength: 2,
					MaxLength: 64,
					Message:   "must be 2-64 characters, start with an alphanumeric character, and contain only alphanumeric characters, dashes, periods or underscores",
				},
			},
			// endpoints and default (isDefault) are server-populated read-only outputs.
			ComputedFields: []string{
				"properties.endpoints",
				"properties.isDefault",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewCognitiveAccountProject()) }
