package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementPolicyFragment provides resource knowledge for
// Microsoft.ApiManagement/service/policyFragments.
//
// Mirrors azurerm_api_management_policy_fragment. The name segment is ForceNew.
// format defaults to "xml" and value carries the policy fragment content.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_policy_fragment_resource.go
//     schema (lines 66-91), Create body (lines 126-132), timeouts (30m/5m/30m/30m)
//   - go-azure-sdk apimanagement/2022-08-01/policyfragment
//     PolicyFragmentContractProperties + PolicyFragmentContentFormat constants
type ApiManagementPolicyFragment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementPolicyFragment)(nil)

func NewApiManagementPolicyFragment() *ApiManagementPolicyFragment {
	return &ApiManagementPolicyFragment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/policyFragments",
			ApiVersions:  []string{"2022-08-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"}, // name (SchemaApiManagementChildName, ForceNew)
			},
			RequiredFields: []string{
				"properties.value",
			},
			// format defaults to "xml" in AzureRM schema.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.format", Value: "xml"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── ApiManagementChildName
				{
					Regex:   `^[a-zA-Z0-9]([a-zA-Z0-9-_]{0,78}[a-zA-Z0-9])?$`,
					Message: "may only contain alphanumeric characters, underscores and dashes up to 80 characters in length",
				},
				// ── format → properties.format ── PolicyFragmentContentFormat enum
				{
					PropertyPath:  "properties.format",
					AllowedValues: []string{"xml", "rawxml"},
					Message:       "format must be one of: xml, rawxml",
				},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementPolicyFragment()) }
