package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementPolicy provides resource knowledge for
// Microsoft.ApiManagement/service/policies.
//
// Contributing Terraform resource: azurerm_api_management_policy.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_policy_resource.go:23-74
//     (schema: api_management_id Required+ForceNew; xml_content Optional+Computed and
//     xml_link, ConflictsWith + ExactlyOneOf between the two)
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_policy_resource.go:95-121
//     (create: PolicyContractProperties.Format + .Value; xml_link -> rawxml-link,
//     xml_content -> rawxml)
//   - terraform-provider-azurerm vendor/.../apimanagement/2022-08-01/apipolicy/constants.go:12-28
//     (PolicyContentFormat: rawxml, rawxml-link, xml, xml-link — shared model)
//   - terraform-provider-azurerm vendor/.../apimanagement/2022-08-01/policy/id_service.go:110-119
//     (resource ID segments: .../service/{serviceName}; policy name segment fixed "policy")
//
// ApiVersions includes 2024-05-01 because the newer service schema also exposes this
// policy sub-resource; the 2022-08-01 policy client is what AzureRM uses to write it.
//
// Intentionally skipped here:
//   - api_management_id: the parent APIM service ID, an AzAPI ID segment not a body prop.
//   - xml_content ExactlyOneOf/ConflictsWith xml_link: both map to the single ARM body
//     path properties.value (differing only in properties.format). ExactlyOneOf over one
//     shared path is not expressible; the body always carries exactly one value anyway.
//   - xml_content Optional+Computed -> properties.value has no static ARM default.
type ApiManagementPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementPolicy)(nil)

// NewApiManagementPolicy returns knowledge for the
// Microsoft.ApiManagement/service/policies resource.
func NewApiManagementPolicy() *ApiManagementPolicy {
	return &ApiManagementPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/policies",
			ApiVersions:  []string{"2022-08-01", "2024-05-01"},
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

func init() { azwise.Register(NewApiManagementPolicy()) }
