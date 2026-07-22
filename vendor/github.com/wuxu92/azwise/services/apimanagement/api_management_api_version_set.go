package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementApiVersionSet provides resource knowledge for
// Microsoft.ApiManagement/service/apiVersionSets.
//
// Contributing Terraform resource: azurerm_api_management_api_version_set.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_version_set_resource.go:26-92
//     (schema: name/rg/apim ForceNew; display_name + versioning_scheme Required;
//     description, version_header_name, version_query_name Optional with ConflictsWith)
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_version_set_resource.go:118-161
//     (create: ApiVersionSetContractProperties displayName/versioningScheme/description/
//     versionHeaderName/versionQueryName; scheme-conditional header/query requirement)
//   - terraform-provider-azurerm internal/services/apimanagement/validate/api_management.go:12-21
//     (ApiManagementChildName regex via schemaz.SchemaApiManagementChildName)
//   - terraform-provider-azurerm vendor/.../apimanagement/2022-08-01/apiversionset/constants.go:14-26
//     (VersioningScheme: Header, Query, Segment)
//   - terraform-provider-azurerm vendor/.../apimanagement/2022-08-01/apiversionset/id_apiversionset.go:116-127
//     (resource ID segments: .../apiVersionSets/{versionSetId})
//
// Intentionally skipped here:
//   - resource_group_name / api_management_name: AzAPI ID segments, not body props.
//   - Scheme-conditional constraints (versioning_scheme==Header requires
//     version_header_name; ==Query requires version_query_name; the other must be
//     unset): a value-dependent cross-field rule that azwise's RelationalRule set
//     (ConflictsWith/RequiredWith/ExactlyOneOf/AtLeastOneOf) cannot express, since it
//     branches on the versioningScheme enum value. The plain ConflictsWith between the
//     two header/query names is captured below.
type ApiManagementApiVersionSet struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementApiVersionSet)(nil)

// NewApiManagementApiVersionSet returns knowledge for the
// Microsoft.ApiManagement/service/apiVersionSets resource.
func NewApiManagementApiVersionSet() *ApiManagementApiVersionSet {
	return &ApiManagementApiVersionSet{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/apiVersionSets",
			ApiVersions:  []string{"2022-08-01"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// name — validate.ApiManagementChildName.
					Regex:     `^[a-zA-Z0-9]([a-zA-Z0-9-_]{0,78}[a-zA-Z0-9])?$`,
					MaxLength: 80,
					Message:   "may only contain alphanumeric characters, underscores and dashes up to 80 characters in length, beginning and ending with an alphanumeric character",
				},
				{
					PropertyPath: "properties.displayName",
					MinLength:    1,
					Message:      "must not be empty",
				},
				{
					PropertyPath:  "properties.versioningScheme",
					AllowedValues: []string{"Header", "Query", "Segment"},
					Message:       "must be one of Header, Query or Segment",
				},
				{
					PropertyPath: "properties.description",
					MinLength:    1,
					Message:      "must not be empty",
				},
				{
					PropertyPath: "properties.versionHeaderName",
					MinLength:    1,
					Message:      "must not be empty",
				},
				{
					PropertyPath: "properties.versionQueryName",
					MinLength:    1,
					Message:      "must not be empty",
				},
			},
			RequiredFields: []string{
				"properties.displayName",
				"properties.versioningScheme",
			},
			// version_header_name ConflictsWith version_query_name (and reverse).
			ConflictsWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.versionHeaderName", "properties.versionQueryName"},
					Message: "`versionHeaderName` and `versionQueryName` cannot be set at the same time",
				},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementApiVersionSet()) }
