package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementApiTagDescription provides resource knowledge for
// Microsoft.ApiManagement/service/apis/tagDescriptions.
//
// Mirrors azurerm_api_management_api_tag_description. `api_tag_id` is an
// envelope/parent reference (ForceNew) that identifies both the API and the tag.
// The ARM body carries optional descriptive fields:
//   - description                        -> properties.description
//   - external_documentation_url         -> properties.externalDocsUrl
//   - external_documentation_description  -> properties.externalDocsDescription
//
// external_documentation_url uses AzureRM's IsURLWithHTTPorHTTPS semantic
// validator, which is not expressible as a declarative StringRule and belongs in
// an azapin URL validator on properties.externalDocsUrl.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_api_tag_description_resource.go
//     schema (lines 41-64); Create body (lines 98-109); Timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/apimanagement/2022-08-01/apitagdescription
//     TagDescriptionBaseProperties: description, externalDocsUrl, externalDocsDescription.
type ApiManagementApiTagDescription struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementApiTagDescription)(nil)

// NewApiManagementApiTagDescription returns knowledge for the
// apis/tagDescriptions resource.
func NewApiManagementApiTagDescription() *ApiManagementApiTagDescription {
	return &ApiManagementApiTagDescription{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service/apis/tagDescriptions",
			ApiVersions:  []string{"2022-08-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementApiTagDescription()) }
