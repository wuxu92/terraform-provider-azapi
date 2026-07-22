package cognitive

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CognitiveAccountRaiPolicy provides resource knowledge for
// Microsoft.CognitiveServices/accounts/raiPolicies.
//
// Contributing TF resource:
//   - azurerm_cognitive_account_rai_policy — cognitive_account_rai_policy_resource.go
//
// Sources:
//   - AzureRM cognitive_account_rai_policy_resource.go schema + expand functions
//   - Azure SDK cognitive/2026-03-01/raipolicies models + constants.go
type CognitiveAccountRaiPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CognitiveAccountRaiPolicy)(nil)

// NewCognitiveAccountRaiPolicy returns a CognitiveAccountRaiPolicy knowledge instance.
func NewCognitiveAccountRaiPolicy() *CognitiveAccountRaiPolicy {
	return &CognitiveAccountRaiPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.CognitiveServices/accounts/raiPolicies",
			ApiVersions:  []string{"2026-03-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.basePolicyName"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// base_policy_name is Required; content_filter (properties.contentFilters)
			// is Required. The Required fields *inside* each content_filter element
			// (name, filter_enabled, block_enabled, severity_threshold, source) are
			// array-element paths and cannot be expressed as RequiredFields — noted here.
			RequiredFields: []string{
				"properties.basePolicyName",
				"properties.contentFilters",
			},
			StringRules: []azwise.StringRule{
				// ── properties.contentFilters[*].severityThreshold ──
				// PossibleValuesForContentLevel().
				{
					PropertyPath:  "properties.contentFilters[*].severityThreshold",
					AllowedValues: []string{"High", "Low", "Medium"},
					Message:       "must be High, Low or Medium",
				},
				// ── properties.contentFilters[*].source ──
				// PossibleValuesForRaiPolicyContentSource().
				{
					PropertyPath: "properties.contentFilters[*].source",
					AllowedValues: []string{
						"Completion", "PostRun", "PostToolCall", "PreRun", "PreToolCall", "Prompt",
					},
					Message: "must be a valid content filter source",
				},
				// ── properties.mode ──
				// PossibleValuesForRaiPolicyMode().
				{
					PropertyPath:  "properties.mode",
					AllowedValues: []string{"Asynchronous_filter", "Blocking", "Default", "Deferred"},
					Message:       "must be a valid RAI policy mode",
				},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewCognitiveAccountRaiPolicy()) }
