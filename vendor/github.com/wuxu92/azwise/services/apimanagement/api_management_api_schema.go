package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementApiSchema provides resource knowledge for
// Microsoft.ApiManagement/service/apis/schemas.
//
// Mirrors azurerm_api_management_api_schema. `schema_id`, `api_name`,
// `api_management_name` and `resource_group_name` are envelope/parent references.
// The ARM body carries the required `content_type` (properties.contentType) and
// the schema document, supplied via exactly one of `value`, `definitions` or
// `components`:
//   - value       -> properties.document.value (string)
//   - components   -> properties.document.components (JSON object)
//   - definitions  -> properties.document.definitions (JSON object)
//
// AzureRM enforces ExactlyOneOf{value, definitions, components} at the schema
// layer. These three project onto distinct ARM paths, so the constraint is
// captured as an ExactlyOneOf relational rule.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_schema_resource.go
//     schema (lines 44-85); Create body (lines 112-139); Timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/apimanagement/2022-08-01/apischema
//     SchemaContractProperties: contentType (required), document{value, components,
//     definitions}.
type ApiManagementApiSchema struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementApiSchema)(nil)

// NewApiManagementApiSchema returns knowledge for the apis/schemas resource.
func NewApiManagementApiSchema() *ApiManagementApiSchema {
	return &ApiManagementApiSchema{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/apis/schemas",
			ApiVersions:  []string{"2022-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// content_type is Required (StringIsNotEmpty).
			RequiredFields: []string{
				"properties.contentType",
			},
			StringRules: []azwise.StringRule{
				{PropertyPath: "properties.contentType", MinLength: 1, Message: "content_type must not be empty"},
			},
			// AzureRM: ExactlyOneOf{value, definitions, components}.
			ExactlyOneOf: []azwise.RelationalRule{
				{
					Paths: []string{
						"properties.document.value",
						"properties.document.definitions",
						"properties.document.components",
					},
					Message: "exactly one of `value`, `definitions` or `components` must be set",
				},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementApiSchema()) }
