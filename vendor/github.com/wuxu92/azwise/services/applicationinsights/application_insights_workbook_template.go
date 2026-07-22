package applicationinsights

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApplicationInsightsWorkbookTemplate provides resource knowledge for
// Microsoft.Insights/workbooktemplates.
//
// Mirrors azurerm_application_insights_workbook_template. name/location are on the
// envelope; the body carries author/galleries/localized/priority/templateData.
//
// Deliberately not encoded here:
//   - template_data (properties.templateData) uses validation.StringIsJSON and is
//     an untyped object in ARM (interface{}); there is no declarative JSON rule.
//   - localized uses validation.StringIsNotEmpty but is JSON-parsed into an ARM
//     map[string][]…, so a string length rule does not apply to the body shape.
//   - galleries is Required with MinItems:1; ArrayRule only carries MaxItems, so the
//     minimum-length and per-element Required (galleries[*].name / .category) are
//     not expressible as RequiredFields. The not-empty checks on those element
//     strings ARE expressed via [*] StringRules.
//   - per-element gallery defaults (order=0, resourceType="Azure Monitor",
//     type="workbook") are array-element paths; DefaultValue does not support [*],
//     so only the top-level priority default is encoded.
//
// Sources:
//   - terraform-provider-azurerm internal/services/applicationinsights/application_insights_workbook_template_resource.go:25-223
//     (schema: name ForceNew, galleries Required MinItems:1 with name/category
//     StringIsNotEmpty + order/resource_type/type defaults, priority default 0,
//     author/localized optional, template_data Required JSON, timeouts 30m/5m/30m/30m;
//     Create maps properties author/galleries/localized/priority/templateData)
//   - go-azure-sdk resource-manager/applicationinsights/2020-11-20/workbooktemplatesapis
//     model_workbooktemplateproperties.go and model_workbooktemplategallery.go (json tags)
type ApplicationInsightsWorkbookTemplate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApplicationInsightsWorkbookTemplate)(nil)

// NewApplicationInsightsWorkbookTemplate returns knowledge for the
// workbooktemplates resource.
func NewApplicationInsightsWorkbookTemplate() *ApplicationInsightsWorkbookTemplate {
	return &ApplicationInsightsWorkbookTemplate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Insights/workbooktemplates",
			ApiVersions:  []string{"2020-11-20"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "properties.author", MinLength: 1, Message: "author must not be empty"},
				// gallery element not-empty checks (StringRules support [*] paths).
				{PropertyPath: "properties.galleries[*].name", MinLength: 1, Message: "galleries.name must not be empty"},
				{PropertyPath: "properties.galleries[*].category", MinLength: 1, Message: "galleries.category must not be empty"},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.priority", Value: int64(0)},
			},
			RequiredFields: []string{
				"properties.galleries",
				"properties.templateData",
			},
		},
	}
}

func init() { azwise.Register(NewApplicationInsightsWorkbookTemplate()) }
