package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationSourceControl provides resource knowledge for
// Microsoft.Automation/automationAccounts/sourceControls.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_source_control_resource.go
//     schema (52-144), Create (158-214); timeouts 30m/5m/10m/10m.
//   - go-azure-sdk resource-manager/automation/2024-10-23/sourcecontrol
//     SourceControlCreateOrUpdateProperties: autoSync/branch/description/folderPath/
//     publishRunbook/repoUrl/sourceType/securityToken(accessToken/refreshToken/tokenType).
//   - sourcecontrol.SourceType / TokenType constants (constants.go:14-18, 58-61).
//   - validators: StringIsNotWhiteSpace (name), StringLenBetween (repository_url 0-2000,
//     branch/folder_path 0-255, token/refresh_token 0-1024), StringInSlice
//     (source_control_type, security.token_type).
//
// name / automation_account_id are the envelope + parent id (ForceNew by
// construction). security.token / security.refresh_token are credentials — see
// SensitiveFields.
type AutomationSourceControl struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationSourceControl)(nil)

func NewAutomationSourceControl() *AutomationSourceControl {
	return &AutomationSourceControl{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/sourceControls",
			ApiVersions:  []string{"2024-10-23"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 10 * time.Minute,
				Delete: 10 * time.Minute,
			},
			// repository_url, folder_path, source_control_type and security are Required.
			RequiredFields: []string{
				"properties.repoUrl",
				"properties.folderPath",
				"properties.sourceType",
				"properties.securityToken.accessToken",
				"properties.securityToken.tokenType",
			},
			DefaultValues: []azwise.DefaultValue{
				// automatic_sync → properties.autoSync, AzureRM schema Default false.
				{PropertyPath: "properties.autoSync", Value: false},
				// publish_runbook_enabled → properties.publishRunbook, AzureRM Default true.
				{PropertyPath: "properties.publishRunbook", Value: true},
			},
			SensitiveFields: []string{
				"properties.securityToken.accessToken",
				"properties.securityToken.refreshToken",
			},
			StringRules: []azwise.StringRule{
				// name — StringIsNotWhiteSpace (approximated by non-empty).
				{MinLength: 1, Message: "name must not be empty or whitespace"},
				// repository_url → properties.repoUrl — StringLenBetween(0, 2000).
				{PropertyPath: "properties.repoUrl", MaxLength: 2000, Message: "repository_url must be at most 2000 characters"},
				// branch → properties.branch — StringLenBetween(0, 255).
				{PropertyPath: "properties.branch", MaxLength: 255, Message: "branch must be at most 255 characters"},
				// folder_path → properties.folderPath — StringLenBetween(0, 255).
				{PropertyPath: "properties.folderPath", MaxLength: 255, Message: "folder_path must be at most 255 characters"},
				// source_control_type → properties.sourceType — StringInSlice (full SDK enum).
				{
					PropertyPath:  "properties.sourceType",
					AllowedValues: []string{"GitHub", "VsoGit", "VsoTfvc"},
					Message:       "source_control_type must be one of GitHub, VsoGit, VsoTfvc",
				},
				// security.token → properties.securityToken.accessToken — StringLenBetween(0, 1024).
				{PropertyPath: "properties.securityToken.accessToken", MaxLength: 1024, Message: "token must be at most 1024 characters"},
				// security.refresh_token → properties.securityToken.refreshToken — StringLenBetween(0, 1024).
				{PropertyPath: "properties.securityToken.refreshToken", MaxLength: 1024, Message: "refresh_token must be at most 1024 characters"},
				// security.token_type → properties.securityToken.tokenType — StringInSlice.
				{
					PropertyPath:  "properties.securityToken.tokenType",
					AllowedValues: []string{"Oauth", "PersonalAccessToken"},
					Message:       "token_type must be one of Oauth, PersonalAccessToken",
				},
			},
		},
	}
}

func init() { azwise.Register(NewAutomationSourceControl()) }
