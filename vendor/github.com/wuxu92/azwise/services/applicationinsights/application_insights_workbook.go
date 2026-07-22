package applicationinsights

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApplicationInsightsWorkbook provides resource knowledge for
// Microsoft.Insights/workbooks.
//
// Mirrors azurerm_application_insights_workbook. AzureRM hardcodes the ARM kind to
// "shared" (the only creatable workbook kind). identity and storage_container_id
// (properties.storageUri) are ForceNew.
//
// Deliberately not encoded here:
//   - name uses validation.All(IsUUID, StringDoesNotContainUpperCaseLetter) on the
//     envelope resource name — a UUID + lowercase semantic check. It belongs in
//     the resource customizer (typegraph.Validator(validators.UUID) + a
//     no-uppercase rule), not a declarative name StringRule (empty-path StringRules
//     are skipped by ApplyAzwise).
//   - data_json (properties.serializedData) uses validation.StringIsJSON; there is
//     no declarative "must be JSON" rule type, so it is not expressed.
//   - tags uses the resource-specific validate.WorkbookTags map validator (rejects
//     reserved hidden-* keys); a map-key validator is not expressible declaratively.
//   - revision/timeModified/userId/version are read-only in the GET model but share
//     the WorkbookProperties struct used for create, so they are NOT listed as
//     strip-able ComputedFields.
//
// Sources:
//   - terraform-provider-azurerm internal/services/applicationinsights/application_insights_workbook_resource.go:26-201
//     (schema: name/identity/storage_container_id ForceNew, category/source_id
//     defaults, storage_container_id RequiredWith identity, timeouts 30m/5m/30m/30m;
//     Create maps kind=shared and properties category/displayName/serializedData/
//     sourceId/description/storageUri)
//   - terraform-provider-azurerm internal/services/applicationinsights/validate/workbook_name.go:11-24
//     (StringDoesNotContainUpperCaseLetter -> no-uppercase regex)
//   - go-azure-sdk resource-manager/applicationinsights/2022-04-01/workbooksapis
//     model_workbookproperties.go (json tags) and constants.go
//     (WorkbookSharedTypeKindShared = "shared")
type ApplicationInsightsWorkbook struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApplicationInsightsWorkbook)(nil)

// NewApplicationInsightsWorkbook returns knowledge for the workbooks resource.
func NewApplicationInsightsWorkbook() *ApplicationInsightsWorkbook {
	return &ApplicationInsightsWorkbook{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Insights/workbooks",
			ApiVersions:  []string{"2022-04-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "identity"},
				{PropertyPath: "properties.storageUri"},
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "properties.displayName", MinLength: 1, Message: "display_name must not be empty"},
				{PropertyPath: "properties.category", MinLength: 1, Message: "category must not be empty"},
				{PropertyPath: "properties.description", MinLength: 1, Message: "description must not be empty"},
				{PropertyPath: "properties.storageUri", MinLength: 1, Message: "storage_container_id must not be empty"},
				// source_id (validate.StringDoesNotContainUpperCaseLetter): no uppercase.
				{PropertyPath: "properties.sourceId", Regex: `^[^A-Z]*$`, Message: "source_id must not contain uppercase letters"},
			},
			DefaultValues: []azwise.DefaultValue{
				// AzureRM hardcodes the workbook kind to the only creatable value.
				{PropertyPath: "kind", Value: "shared"},
				{PropertyPath: "properties.category", Value: "workbook"},
				{PropertyPath: "properties.sourceId", Value: "azure monitor"},
			},
			RequiredFields: []string{
				"properties.displayName",
				"properties.serializedData",
			},
			// storage_container_id RequiredWith identity (workbook schema line 118).
			RequiredWith: []azwise.RelationalRule{
				{Paths: []string{"properties.storageUri", "identity"}},
			},
		},
	}
}

func init() { azwise.Register(NewApplicationInsightsWorkbook()) }
