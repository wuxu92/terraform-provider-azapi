package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementEmailTemplate provides resource knowledge for
// Microsoft.ApiManagement/service/templates.
//
// Mirrors azurerm_api_management_email_template.
//
// Sources (terraform-provider-azurerm internal/services/apimanagement):
//   - api_management_email_template_resource.go (schema L42-87; expand L119-124)
//   - SDK model EmailTemplateUpdateParameterProperties (subject/body/title/description/parameters)
//   - emailtemplates/constants.go TemplateName enum
//
// The resource name (`template_name`) is ForceNew and constrained to the
// TemplateName enum. AzAPI sends raw ARM values, so the full SDK enum set
// (camelCase) is used rather than AzureRM's TitleCase display variants.
//
// Note: `title` and `description` are marked Computed in the AzureRM schema but
// are present in the ARM update model (EmailTemplateUpdateParameterProperties),
// so they are settable via AzAPI and are NOT listed as ComputedFields.
type ApiManagementEmailTemplate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementEmailTemplate)(nil)

// NewApiManagementEmailTemplate returns knowledge for the templates resource.
func NewApiManagementEmailTemplate() *ApiManagementEmailTemplate {
	return &ApiManagementEmailTemplate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/templates",
			ApiVersions:  []string{"2022-08-01"},
			RequiredFields: []string{
				"properties.subject",
				"properties.body",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// Resource name is the email template name.
				{AllowedValues: []string{
					"accountClosedDeveloper",
					"applicationApprovedNotificationMessage",
					"confirmSignUpIdentityDefault",
					"emailChangeIdentityDefault",
					"inviteUserNotificationMessage",
					"newCommentNotificationMessage",
					"newDeveloperNotificationMessage",
					"newIssueNotificationMessage",
					"passwordResetByAdminNotificationMessage",
					"passwordResetIdentityDefault",
					"purchaseDeveloperNotificationMessage",
					"quotaLimitApproachingDeveloperNotificationMessage",
					"rejectDeveloperNotificationMessage",
					"requestDeveloperNotificationMessage",
				}, Message: "template_name must be a valid TemplateName"},
				{PropertyPath: "properties.body", MinLength: 1, Message: "body must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementEmailTemplate()) }
