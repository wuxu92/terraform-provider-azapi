package appservice

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AppServiceSourceControl provides resource knowledge for
// Microsoft.Web/sites/sourcecontrols (the fixed child instance named "web").
//
// Mirrors azurerm_app_service_source_control. The parent envelope is the web app
// (azurerm app_id). The resource has no Update path — every mapped body property
// is ForceNew. use_local_git is not represented: AzureRM implements it as a
// SiteConfig.scmType=LocalGit PATCH on the parent site, not a sourceControls body
// field, so its ConflictsWith relation against the other fields is not expressible
// here.
//
// Sources:
//   - terraform-provider-azurerm internal/services/appservice/app_service_source_control_resource.go:23-222
//     (schema: repo_url/branch RequiredWith each other, all body fields ForceNew;
//     create maps repoUrl/branch/isManualIntegration/isMercurial/deploymentRollbackEnabled/
//     gitHubActionConfiguration; timeouts Create 30m / Read 5m / Delete 30m)
//   - terraform-provider-azurerm internal/services/appservice/source_control_schema.go:32-162
//     (github_action_configuration block: code/container config, runtime_stack enum,
//     generate_workflow_file default true, linux_action computed)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/model_sitesourcecontrolproperties.go:6-14
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/model_githubactionconfiguration.go:6-11
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/model_githubactioncodeconfiguration.go:6-9
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/model_githubactioncontainerconfiguration.go:6-11
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/method_updatesourcecontrol.go:32 (PATCH {site}/sourceControls/web)
type AppServiceSourceControl struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AppServiceSourceControl)(nil)

// NewAppServiceSourceControl returns knowledge for the sites/sourcecontrols resource.
func NewAppServiceSourceControl() *AppServiceSourceControl {
	return &AppServiceSourceControl{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Web/sites/sourcecontrols",
			ApiVersions:  []string{"2023-12-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.repoUrl"},
				{PropertyPath: "properties.branch"},
				{PropertyPath: "properties.isManualIntegration"},
				{PropertyPath: "properties.isMercurial"},
				{PropertyPath: "properties.deploymentRollbackEnabled"},
				{PropertyPath: "properties.gitHubActionConfiguration"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "properties.repoUrl", MinLength: 1, Message: "repo_url must not be empty"},
				{PropertyPath: "properties.branch", MinLength: 1, Message: "branch must not be empty"},
				{
					PropertyPath:  "properties.gitHubActionConfiguration.codeConfiguration.runtimeStack",
					AllowedValues: []string{"dotnetcore", "spring", "tomcat", "node", "python"},
					Message:       "runtime_stack must be one of dotnetcore, spring, tomcat, node, python",
				},
				{PropertyPath: "properties.gitHubActionConfiguration.codeConfiguration.runtimeVersion", MinLength: 1, Message: "runtime_version must not be empty"},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.isManualIntegration", Value: false},
				{PropertyPath: "properties.isMercurial", Value: false},
				{PropertyPath: "properties.deploymentRollbackEnabled", Value: false},
				{PropertyPath: "properties.gitHubActionConfiguration.generateWorkflowFile", Value: true},
			},
			RequiredWith: []azwise.RelationalRule{
				{Paths: []string{"properties.repoUrl", "properties.branch"}},
				{Paths: []string{"properties.branch", "properties.repoUrl"}},
			},
			SensitiveFields: []string{
				"properties.gitHubActionConfiguration.containerConfiguration.password",
			},
			ComputedFields: []string{
				"properties.isGitHubAction",
			},
		},
	}
}

func init() { azwise.Register(NewAppServiceSourceControl()) }
