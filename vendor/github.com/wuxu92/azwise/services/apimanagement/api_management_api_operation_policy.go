package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementApiOperationPolicy provides resource knowledge for
// Microsoft.ApiManagement/service/apis/operations/policies.
//
// Contributing Terraform resource: azurerm_api_management_api_operation_policy.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_operation_policy_resource.go:26-72
//     (schema: rg/apim/api/operation ForceNew; xml_content Optional+Computed ConflictsWith xml_link)
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_operation_policy_resource.go:98-119
//     (create: PolicyContractProperties.Format + .Value; xml_content -> rawxml,
//     xml_link -> rawxml-link)
//   - terraform-provider-azurerm vendor/.../apimanagement/2022-08-01/apipolicy/constants.go:12-28
//     (PolicyContentFormat: rawxml, rawxml-link, xml, xml-link — shared model)
//
// Intentionally skipped here:
//   - resource_group_name / api_management_name / api_name / operation_id: AzAPI ID
//     segments, not body properties.
//   - xml_content ConflictsWith xml_link: both map to the single ARM body path
//     properties.value (differing only in properties.format); a ConflictsWith over one
//     shared path is not expressible and unnecessary (only one value is ever sent).
//   - xml_content Optional+Computed -> properties.value has no static ARM default.
type ApiManagementApiOperationPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementApiOperationPolicy)(nil)

// NewApiManagementApiOperationPolicy returns knowledge for the
// Microsoft.ApiManagement/service/apis/operations/policies resource.
func NewApiManagementApiOperationPolicy() *ApiManagementApiOperationPolicy {
	return &ApiManagementApiOperationPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/apis/operations/policies",
			ApiVersions:  []string{"2022-08-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.format",
					AllowedValues: []string{"rawxml", "rawxml-link", "xml", "xml-link"},
					Message:       "must be one of rawxml, rawxml-link, xml or xml-link",
				},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementApiOperationPolicy()) }
