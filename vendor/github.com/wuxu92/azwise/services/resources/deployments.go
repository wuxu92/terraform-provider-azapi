package resources

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Deployments provides resource knowledge for Microsoft.Resources/deployments.
//
// AzureRM models one ARM type (Microsoft.Resources/deployments) as four scope-specific
// Terraform resources that share a single base schema; they differ only by the scope the
// deployment targets (resource group / subscription / management group / tenant). Their
// universal operational knowledge is merged here for the single ARM type. Scope-specific
// fields (deployment_mode is Required only at RG scope, location is Required+ForceNew only
// at subscription/mgmt-group/tenant scope, management_group_id) are intentionally NOT
// unioned — they would corrupt validation for the sibling scopes that never set them.
//
// Sources:
//   - terraform-provider-azurerm internal/services/resource/resource_group_template_deployment_resource.go:48-136,161-189
//     (RG scope: name=TemplateDeploymentName ForceNew, deployment_mode Required enum,
//     template_content/template_spec_version_id ExactlyOneOf, debug_level enum, 180m timeouts,
//     output_content Computed, Mode from field)
//   - terraform-provider-azurerm internal/services/resource/subscription_template_deployment_resource.go:48-102,127-134
//     (subscription scope: adds location, hardcodes Mode=Incremental)
//   - terraform-provider-azurerm internal/services/resource/management_group_template_deployment_resource.go:50-111,142-145
//     (mgmt-group scope: adds management_group_id + location, hardcodes Mode=Incremental)
//   - terraform-provider-azurerm internal/services/resource/tenant_template_deployment_resource.go:48-102,128-129
//     (tenant scope: adds location, hardcodes Mode=Incremental)
//   - terraform-provider-azurerm internal/services/resource/validate/template_deployment_name.go:11-22
//     (name regex ^([a-zA-Z0-9-._\(\)]){1,}?$)
//   - terraform-provider-azurerm internal/services/resource/template_deployment_common.go:24-50
//     (debug_level -> properties.debugSetting.detailLevel: none/requestContent/responseContent/
//     "requestContent, responseContent"; default "none")
//   - terraform-provider-azurerm vendor/github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2020-06-01/resources/models.go:317-346,774-789,3336-3343
//     (Deployment{location,properties,tags}; DeploymentProperties{template,templateLink,parameters,
//     mode,debugSetting}; DebugSetting.detailLevel; TemplateLink.id — ARM body json tags)
//
// Intentionally skipped (documented, no rule emitted):
//   - location: Required+ForceNew via commonschema.Location() only at subscription/mgmt-group/
//     tenant scope; absent at RG scope. Kind-specific — not unioned.
//   - management_group_id: parent scope, mgmt-group only; lives on the ID, not the body.
//   - properties.parameters / properties.template: free-form JObject payloads, no expressible rule.
type Deployments struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Deployments)(nil)

// NewDeployments returns knowledge for the deployments resource.
func NewDeployments() *Deployments {
	return &Deployments{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Resources/deployments",
			ApiVersions:  []string{"2020-06-01"},
			SoftDelete:   false,
			// Only the deployment name is replace-only across every scope. deployment_mode is
			// updatable (RG scope patches it in place); location is ForceNew but scope-specific.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			// properties.mode is present in every body this provider emits (RG maps it from
			// deployment_mode; the other scopes hardcode Incremental) and is required by ARM.
			RequiredFields: []string{
				"properties.mode",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 180 * time.Minute,
				Read:   5 * time.Minute,
				Update: 180 * time.Minute,
				Delete: 180 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). AzureRM validate.TemplateDeploymentName:
					// alphanumeric, dashes, dots, underscores and parentheses.
					Regex:   `^([a-zA-Z0-9-._()]){1,}?$`,
					Message: "name may only contain alphanumeric characters, dashes, full-stops, underscores and parentheses",
				},
				{
					// deployment_mode -> properties.mode. Enum fires only when mode is present
					// (always, here). Full ARM set is Incremental/Complete.
					PropertyPath:  "properties.mode",
					AllowedValues: []string{"Incremental", "Complete"},
					Message:       "deployment mode must be one of Incremental, Complete",
				},
				{
					// debug_level -> properties.debugSetting.detailLevel.
					PropertyPath:  "properties.debugSetting.detailLevel",
					AllowedValues: []string{"none", "requestContent", "responseContent", "requestContent, responseContent"},
					Message:       "debug_level must be one of none, requestContent, responseContent, or \"requestContent, responseContent\"",
				},
			},
			ComputedFields: []string{
				// output_content is read-only (DeploymentPropertiesExtended.outputs), absent
				// from the create body.
				"properties.outputs",
			},
			DefaultValues: []azwise.DefaultValue{
				// expandTemplateDeploymentDebugSetting defaults detailLevel to "none".
				{PropertyPath: "properties.debugSetting.detailLevel", Value: "none"},
			},
			// template_content (properties.template) and template_spec_version_id
			// (properties.templateLink) are ExactlyOneOf across every scope.
			ExactlyOneOf: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.template", "properties.templateLink"},
					Message: "exactly one of template_content (properties.template) or template_spec_version_id (properties.templateLink) must be set",
				},
			},
		},
	}
}

func init() { azwise.Register(NewDeployments()) }
