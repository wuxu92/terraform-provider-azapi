package cognitive

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CognitiveAccountRaiBlocklist provides resource knowledge for
// Microsoft.CognitiveServices/accounts/raiBlocklists.
//
// Contributing TF resource:
//   - azurerm_cognitive_account_rai_blocklist — cognitive_account_rai_blocklist_resource.go
//
// Sources:
//   - AzureRM cognitive_account_rai_blocklist_resource.go schema + CRUD
//   - AzureRM validate/rai_blocklist_name.go (name regex)
//   - Azure SDK cognitive/2026-03-01/raiblocklists models
type CognitiveAccountRaiBlocklist struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CognitiveAccountRaiBlocklist)(nil)

// NewCognitiveAccountRaiBlocklist returns a CognitiveAccountRaiBlocklist knowledge instance.
func NewCognitiveAccountRaiBlocklist() *CognitiveAccountRaiBlocklist {
	return &CognitiveAccountRaiBlocklist{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.CognitiveServices/accounts/raiBlocklists",
			ApiVersions:  []string{"2026-03-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ──
				// validate.RaiBlocklistName(): ^[a-zA-Z0-9_-]{2,64}$
				{
					Regex:     `^[a-zA-Z0-9_-]{2,64}$`,
					MinLength: 2,
					MaxLength: 64,
					Message:   "must be 2-64 characters and contain only alphanumeric characters, hyphens or underscores",
				},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewCognitiveAccountRaiBlocklist()) }
