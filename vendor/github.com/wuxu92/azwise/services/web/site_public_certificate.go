package web

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SitePublicCertificate provides resource knowledge for
// Microsoft.Web/sites/publicCertificates.
//
// Backing AzureRM Terraform resource: azurerm_app_service_public_certificate.
//
// Sources:
//   - terraform-provider-azurerm internal/services/web/app_service_public_certificate_resource.go:23-103
//     (schema: app_service_name/certificate_name/certificate_location/blob all Required ForceNew;
//     certificate_location StringInSlice(PossibleValuesForPublicCertificateLocation);
//     blob StringIsBase64; create mapping certificate_location->
//     properties.publicCertificateLocation, blob->properties.blob)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/id_publiccertificate.go:109-112
//     (ID casing: Microsoft.Web/sites/publicCertificates)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/model_publiccertificateproperties.go:6-10
//     (properties JSON tags: blob *string, publicCertificateLocation *PublicCertificateLocation)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/web/2023-12-01/webapps/constants.go:1677-1680
//     (PublicCertificateLocation enum: CurrentUserMy, LocalMachineMy, Unknown)
//
// Intentionally skipped here:
//   - blob validation.StringIsBase64: base64 format check over properties.blob; a declarative
//     regex cannot faithfully express base64 padding, so it is not approximated.
//   - No Update; every schema field is ForceNew.
type SitePublicCertificate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SitePublicCertificate)(nil)

// NewSitePublicCertificate returns knowledge for the sites/publicCertificates resource.
func NewSitePublicCertificate() *SitePublicCertificate {
	return &SitePublicCertificate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Web/sites/publicCertificates",
			ApiVersions:  []string{"2023-12-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.publicCertificateLocation"},
				{PropertyPath: "properties.blob"},
			},
			RequiredFields: []string{
				"properties.publicCertificateLocation",
				"properties.blob",
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "properties.publicCertificateLocation",
					AllowedValues: []string{
						"CurrentUserMy",
						"LocalMachineMy",
						"Unknown",
					},
					Message: "must be a valid public certificate location",
				},
			},
		},
	}
}

func init() { azwise.Register(NewSitePublicCertificate()) }
