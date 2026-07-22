package web

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CertificateOrder provides resource knowledge for
// Microsoft.CertificateRegistration/certificateOrders.
//
// Backing AzureRM Terraform resource: azurerm_app_service_certificate_order.
//
// Sources:
//   - terraform-provider-azurerm internal/services/web/app_service_certificate_order_resource.go:25-178
//     (schema: name/location ForceNew; auto_renew Default true; csr ConflictsWith
//     distinguished_name; key_size Default 2048 IntAtLeast(0); product_type Default "Standard"
//     StringInSlice; validity_in_years Default 1 IntBetween(1,3))
//   - terraform-provider-azurerm internal/services/web/app_service_certificate_order_resource.go:180-221,399-404
//     (create mapping to AppServiceCertificateOrderProperties: autoRenew, csr, distinguishedName,
//     keySize, productType, validityInYears; expandProductType maps Standard->
//     StandardDomainValidatedSsl, WildCard->StandardDomainValidatedWildCardSsl)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/certificateregistration/2023-12-01/appservicecertificateorders/id_certificateorder.go:102-120
//     (ID casing: Microsoft.CertificateRegistration/certificateOrders)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/certificateregistration/2023-12-01/appservicecertificateorders/model_appservicecertificateorderproperties.go:12-33
//     (properties JSON tags: autoRenew *bool, csr/distinguishedName *string, keySize *int64,
//     productType CertificateProductType, validityInYears *int64)
//   - terraform-provider-azurerm vendor/github.com/hashicorp/go-azure-sdk/resource-manager/certificateregistration/2023-12-01/appservicecertificateorders/constants.go:154-166
//     (CertificateProductType enum: StandardDomainValidatedSsl, StandardDomainValidatedWildCardSsl)
//
// Intentionally skipped here:
//   - AzureRM computed attributes (certificates, domain_verification_token, status,
//     expiration_time, is_private_key_external, app_service_certificate_not_renewable_reasons,
//     signed_certificate_thumbprint, root_thumbprint, intermediate_thumbprint) are NOT listed as
//     ComputedFields: the 2023-12-01 SDK uses a single AppServiceCertificateOrderProperties model
//     for create and read, so those JSON paths are in the create body; stripping them would
//     discard valid raw ARM input.
//   - csr/distinguished_name are Optional+Computed; azwise leaves them settable and only
//     encodes their mutual-exclusion as a ConflictsWith relational rule.
type CertificateOrder struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CertificateOrder)(nil)

// NewCertificateOrder returns knowledge for the certificateOrders resource.
func NewCertificateOrder() *CertificateOrder {
	return &CertificateOrder{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.CertificateRegistration/certificateOrders",
			ApiVersions:  []string{"2023-12-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					MinLength: 1,
					Message:   "certificate order name must not be empty",
				},
				{
					PropertyPath: "properties.productType",
					AllowedValues: []string{
						"StandardDomainValidatedSsl",
						"StandardDomainValidatedWildCardSsl",
					},
					Message: "must be a valid certificate product type",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.keySize",
					MinValue:     azwise.Ptr(int64(0)),
					Message:      "key size must be non-negative",
				},
				{
					PropertyPath: "properties.validityInYears",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(3)),
					Message:      "validity in years must be between 1 and 3",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.autoRenew", Value: true},
				{PropertyPath: "properties.keySize", Value: int64(2048)},
				{PropertyPath: "properties.productType", Value: "StandardDomainValidatedSsl"},
				{PropertyPath: "properties.validityInYears", Value: int64(1)},
			},
			ConflictsWith: []azwise.RelationalRule{
				{Paths: []string{"properties.csr", "properties.distinguishedName"}},
			},
		},
	}
}

func init() { azwise.Register(NewCertificateOrder()) }
