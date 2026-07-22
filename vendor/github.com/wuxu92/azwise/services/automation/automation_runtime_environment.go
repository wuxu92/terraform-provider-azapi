package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationRuntimeEnvironment provides resource knowledge for
// Microsoft.Automation/automationAccounts/runtimeEnvironments.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_runtime_environment_resource.go
//     schema (81-139), CustomizeDiff (41-79), Create (153-215); timeouts 10m/5m/10m/10m.
//   - go-azure-sdk resource-manager/automation/2024-10-23/runtimeenvironment
//     RuntimeEnvironmentProperties: runtime(language/version)/defaultPackages/description.
//   - validators: StringIsNotEmpty (name, runtime_version, description),
//     StringInSlice (runtime_language: Python, PowerShell).
//
// name / automation_account_id are the envelope + parent id (ForceNew by
// construction).
type AutomationRuntimeEnvironment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationRuntimeEnvironment)(nil)

func NewAutomationRuntimeEnvironment() *AutomationRuntimeEnvironment {
	return &AutomationRuntimeEnvironment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/runtimeEnvironments",
			ApiVersions:  []string{"2024-10-23"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 10 * time.Minute,
				Read:   5 * time.Minute,
				Update: 10 * time.Minute,
				Delete: 10 * time.Minute,
			},
			// runtime_language, runtime_version and description are ForceNew in the body.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.runtime.language"},
				{PropertyPath: "properties.runtime.version"},
				{PropertyPath: "properties.description"},
			},
			// runtime_language and runtime_version are Required.
			RequiredFields: []string{
				"properties.runtime.language",
				"properties.runtime.version",
			},
			StringRules: []azwise.StringRule{
				// name — StringIsNotEmpty.
				{MinLength: 1, Message: "name must not be empty"},
				// runtime_language → properties.runtime.language — StringInSlice.
				{
					PropertyPath:  "properties.runtime.language",
					AllowedValues: []string{"Python", "PowerShell"},
					Message:       "runtime_language must be one of Python, PowerShell",
				},
				// runtime_version → properties.runtime.version — StringIsNotEmpty.
				{PropertyPath: "properties.runtime.version", MinLength: 1, Message: "runtime_version must not be empty"},
				// description → properties.description — StringIsNotEmpty.
				{PropertyPath: "properties.description", MinLength: 1, Message: "description must not be empty"},
			},
			// TODO: runtime_default_packages → properties.defaultPackages is conditionally
			// ForceNew (CustomizeDiff: only when a package key is removed; version-only
			// changes update in place). This map-key-conditional replacement cannot be
			// expressed with the declarative ForceNew rule set, so it is omitted here.
			// Also, defaultPackages cannot be set when runtime_language is "Python".
		},
	}
}

func init() { azwise.Register(NewAutomationRuntimeEnvironment()) }
