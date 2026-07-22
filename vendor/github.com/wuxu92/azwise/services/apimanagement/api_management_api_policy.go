package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementApiPolicy provides resource knowledge for
// Microsoft.ApiManagement/service/apis/policies.
//
// Contributing Terraform resource: azurerm_api_management_api_policy.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_policy_resource.go:26-70
//     (schema: rg/apim/api ForceNew; xml_content Optional+Computed ConflictsWith xml_link)
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_policy_resource.go:96-122
//     (create: PolicyContractProperties.Format + .Value; xml_link -> rawxml-link,
//     xml_content -> rawxml)
//   - terraform-provider-azurerm vendor/.../apimanagement/2022-08-01/apipolicy/constants.go:12-28
//     (PolicyContentFormat: rawxml, rawxml-link, xml, xml-link)
//   - terraform-provider-azurerm vendor/.../apimanagement/2022-08-01/apipolicy/id_api.go:116-127
//     (resource ID segments: .../apis/{apiId}; policy name segment is fixed "policy")
//
// Intentionally skipped here:
//   - resource_group_name / api_management_name / api_name: AzAPI ID segments, not
//     body properties.
//   - xml_content ConflictsWith xml_link: both Terraform fields map to the single ARM
//     body path properties.value (differing only in properties.format). azwise cannot
//     express a ConflictsWith over one shared path, so the mutual-exclusion is omitted;
//     the underlying ARM body only ever carries one value regardless.
//   - xml_content Optional+Computed: maps to properties.value which has no static ARM
//     default (server echoes the applied policy), so no DefaultValue is emitted.
type ApiManagementApiPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementApiPolicy)(nil)

// NewApiManagementApiPolicy returns knowledge for the
// Microsoft.ApiManagement/service/apis/policies resource.
func NewApiManagementApiPolicy() *ApiManagementApiPolicy {
	return &ApiManagementApiPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/apis/policies",
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

func init() { azwise.Register(NewApiManagementApiPolicy()) }
