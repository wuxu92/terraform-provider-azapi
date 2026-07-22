package appservice

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AppServiceSourceControlSlot provides resource knowledge for
// Microsoft.Web/sites/slots/sourcecontrols (the fixed child instance named "web").
//
// Mirrors azurerm_app_service_source_control_slot. Identical body/schema to
// azurerm_app_service_source_control except the parent envelope is a deployment
// slot (azurerm slot_id -> Microsoft.Web/sites/slots) rather than the site
// itself. The resource has no Update path — every mapped body property is
// ForceNew. use_local_git is not represented: AzureRM implements it as a
// SiteConfig.scmType=LocalGit PATCH on the parent slot, not a sourceControls
// body field, so its ConflictsWith relation is not expressible here.
//
// Sources:
//   - terraform-provider-azurerm internal/services/appservice/app_service_source_control_slot_resource.go:24-225
//     (schema: repo_url/branch RequiredWith each other, all body fields ForceNew;
//     create maps repoUrl/branch/isManualIntegration/isMercurial/deploymentRollbackEnabled/
//     gitHubActionConfiguration; timeouts Create 30m / Read 5m / Delete 30m)
//   - terraform-provider-azurerm internal/services/appservice/source_control_schema.go:32-162
//     (shared github_action_configuration block schema/expand)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/model_sitesourcecontrolproperties.go:6-14
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/method_updatesourcecontrolslot.go:31 (PATCH {slot}/sourceControls/web)
type AppServiceSourceControlSlot struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AppServiceSourceControlSlot)(nil)

// NewAppServiceSourceControlSlot returns knowledge for the
// sites/slots/sourcecontrols resource.
func NewAppServiceSourceControlSlot() *AppServiceSourceControlSlot {
	return &AppServiceSourceControlSlot{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Web/sites/slots/sourcecontrols",
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

func init() { azwise.Register(NewAppServiceSourceControlSlot()) }
