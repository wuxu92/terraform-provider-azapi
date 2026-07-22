package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementApiOperation provides resource knowledge for
// Microsoft.ApiManagement/service/apis/operations.
//
// Contributing Terraform resource: azurerm_api_management_api_operation.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_operation_resource.go:25-153
//     (schema: operation_id/api/apim/rg ForceNew; display_name/method/url_template
//     Required; description Optional; request/response/template_parameter nested blocks;
//     CustomizeDiff cross-checking url_template vs template_parameter)
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_operation_resource.go:199-209
//     (create: OperationContractProperties Description/DisplayName/Method/Request/
//     Responses/TemplateParameters/UrlTemplate)
//   - terraform-provider-azurerm internal/services/apimanagement/validate/api_management.go:12-21
//     (ApiManagementChildName regex via schemaz.SchemaApiManagementChildName)
//   - terraform-provider-azurerm vendor/.../apimanagement/2022-08-01/apioperation/id_operation.go:122-135
//     (resource ID segments: .../apis/{apiId}/operations/{operationId})
//
// Intentionally skipped here:
//   - resource_group_name / api_management_name / api_name: AzAPI ID segments.
//   - CustomizeDiff (each {param} used in url_template must be declared in
//     template_parameter and vice-versa): a structural cross-field consistency check
//     between a string template and an array of parameter names. This has no single ARM
//     body field to attach to and cannot be expressed with azwise's declarative rules;
//     ARM enforces the pairing server-side.
//   - request/response/template_parameter nested blocks: modelled structurally by ARM
//     (properties.request, properties.responses[], properties.templateParameters[]);
//     they carry no enum/range/regex validators in AzureRM beyond required-ness of inner
//     fields, so only the top-level required contract is captured here.
type ApiManagementApiOperation struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementApiOperation)(nil)

// NewApiManagementApiOperation returns knowledge for the
// Microsoft.ApiManagement/service/apis/operations resource.
func NewApiManagementApiOperation() *ApiManagementApiOperation {
	return &ApiManagementApiOperation{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/apis/operations",
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
					// operation_id (name) — validate.ApiManagementChildName.
					Regex:     `^[a-zA-Z0-9]([a-zA-Z0-9-_]{0,78}[a-zA-Z0-9])?$`,
					MaxLength: 80,
					Message:   "may only contain alphanumeric characters, underscores and dashes up to 80 characters in length, beginning and ending with an alphanumeric character",
				},
			},
			RequiredFields: []string{
				"properties.displayName",
				"properties.method",
				"properties.urlTemplate",
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementApiOperation()) }
