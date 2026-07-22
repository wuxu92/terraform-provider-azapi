package applicationinsights

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApplicationInsightsAnalyticsItem provides resource knowledge for
// Microsoft.Insights/components/analyticsItems (shared scope; the user scope uses
// the sibling myAnalyticsItems path).
//
// Mirrors azurerm_application_insights_analytics_item. The PUT body is flat with
// PascalCase json tags (Content/Name/Scope/Type) plus a nested
// Properties.functionAlias. name/application_insights_id are on the envelope;
// scope and type are ForceNew body properties.
//
// Deliberately not encoded here:
//   - name uses validation.StringIsNotEmpty on the envelope resource name
//     (empty-path StringRules are skipped by ApplyAzwise).
//   - Version/TimeCreated/TimeModified/Id are read-only in the GET model but share
//     the ApplicationInsightsComponentAnalyticsItem struct used for the PUT, so they
//     are NOT listed as strip-able ComputedFields.
//
// Sources:
//   - terraform-provider-azurerm internal/services/applicationinsights/application_insights_analytics_item_resource.go:29-197
//     (schema: name/scope/type ForceNew, scope StringInSlice[shared,user], type
//     StringInSlice[query,function,folder,recent], content Required, function_alias
//     optional, version/time_created/time_modified Computed, timeouts 30m/5m/30m/30m;
//     Create maps Content/Name/Scope/Type + Properties.functionAlias)
//   - go-azure-sdk resource-manager/applicationinsights/2015-05-01/analyticsitemsapis
//     model_applicationinsightscomponentanalyticsitem.go /
//     model_applicationinsightscomponentanalyticsitemproperties.go (json tags),
//     constants.go (ItemScope shared/user; ItemType function/none/query/recent +
//     ItemTypeParameter folder), id_providercomponent.go (ARM type + 2015-05-01)
type ApplicationInsightsAnalyticsItem struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApplicationInsightsAnalyticsItem)(nil)

// NewApplicationInsightsAnalyticsItem returns knowledge for the analyticsItems
// resource.
func NewApplicationInsightsAnalyticsItem() *ApplicationInsightsAnalyticsItem {
	return &ApplicationInsightsAnalyticsItem{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Insights/components/analyticsItems",
			ApiVersions:  []string{"2015-05-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "Scope"},
				{PropertyPath: "Type"},
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "Scope", AllowedValues: []string{"shared", "user"}},
				// AzureRM validates query/function/folder/recent; the ARM Type enum
				// best-effort-parses the value ("folder" comes from ItemTypeParameter).
				{PropertyPath: "Type", AllowedValues: []string{"query", "function", "folder", "recent"}},
			},
			RequiredFields: []string{
				"Content",
				"Scope",
				"Type",
			},
		},
	}
}

func init() { azwise.Register(NewApplicationInsightsAnalyticsItem()) }
