package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementProductPolicy provides resource knowledge for
// Microsoft.ApiManagement/service/products/policies.
//
// Mirrors azurerm_api_management_product_policy. The policy is set from either
// xml_content or xml_link (mutually exclusive in AzureRM); both map to the single
// ARM property properties.value (with properties.format carrying the content
// format). The policy resource name is always "policy" (fixed ARM segment).
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_product_policy_resource.go
//     schema (lines 49-69), CreateUpdate body (lines 96-117), timeouts (30m/5m/30m/30m)
//   - go-azure-sdk apimanagement/2022-08-01/productpolicy PolicyContractProperties
type ApiManagementProductPolicy struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementProductPolicy)(nil)

func NewApiManagementProductPolicy() *ApiManagementProductPolicy {
	return &ApiManagementProductPolicy{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/products/policies",
			ApiVersions:  []string{"2022-08-01"},
			// xml_content is Required in effect (either xml_content or xml_link);
			// both expand into properties.value.
			RequiredFields: []string{
				"properties.value",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── properties.format ── PolicyContentFormat enum (rawxml, rawxml-link, xml, xml-link)
				{
					PropertyPath:  "properties.format",
					AllowedValues: []string{"xml", "xml-link", "rawxml", "rawxml-link"},
					Message:       "format must be one of: xml, xml-link, rawxml, rawxml-link",
				},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementProductPolicy()) }
