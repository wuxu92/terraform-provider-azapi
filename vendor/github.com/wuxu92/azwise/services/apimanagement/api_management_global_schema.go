package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementGlobalSchema provides resource knowledge for
// Microsoft.ApiManagement/service/schemas.
//
// Mirrors azurerm_api_management_global_schema. `schema_id`, `api_management_name`
// and `resource_group_name` are envelope references. The ARM body carries the
// required `type` (properties.schemaType, enum json|xml), the required `value`,
// and optional `description`.
//
// AzureRM routes `value` to one of two ARM fields depending on `type`: for
// type=json it is JSON-unmarshalled into properties.document; for type=xml it is
// sent as-is in properties.value (schema resource_go lines 101-112). This
// one-Terraform-field-to-two-ARM-fields split cannot be expressed as a single
// declarative RequiredFields path, so `value` is not listed as a RequiredField
// here (either properties.document or properties.value satisfies it at the API).
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_global_schema_resource.go
//     schema (lines 43-68); Create body (lines 94-112); Timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/apimanagement/2022-08-01/schema
//     GlobalSchemaContractProperties: schemaType (enum), description, document, value.
//     SchemaType constants: json, xml (constants.go lines 15-16).
type ApiManagementGlobalSchema struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementGlobalSchema)(nil)

// NewApiManagementGlobalSchema returns knowledge for the schemas resource.
func NewApiManagementGlobalSchema() *ApiManagementGlobalSchema {
	return &ApiManagementGlobalSchema{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/schemas",
			ApiVersions:  []string{"2022-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// `type` is Required (StringInSlice PossibleValuesForSchemaType).
			RequiredFields: []string{
				"properties.schemaType",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.schemaType",
					AllowedValues: []string{"json", "xml"},
					Message:       "schema type must be one of json or xml",
				},
				{PropertyPath: "properties.description", MinLength: 1, Message: "description must not be empty"},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementGlobalSchema()) }
