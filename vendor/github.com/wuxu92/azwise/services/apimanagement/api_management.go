package apimanagement

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ApiManagementService provides resource knowledge for
// Microsoft.ApiManagement/service.
//
// Mirrors azurerm_api_management. This is a large resource; the knowledge below
// was assembled block-by-block from the AzureRM schema and Create/Update
// expand logic.
//
// Sub-service API separation — the following AzureRM blocks are NOT part of the
// apimanagementservice body and are managed as separate ARM sub-resources, so
// their rules belong in their own knowledge files (out of scope here):
//   - sign_in       -> Microsoft.ApiManagement/service/portalsettings (signin)
//   - sign_up       -> Microsoft.ApiManagement/service/portalsettings (signup)
//   - delegation    -> Microsoft.ApiManagement/service/portalsettings (delegation)
//   - tenant_access -> Microsoft.ApiManagement/service/tenant/access
//
// One-to-many / dictionary mapping — the `security` and `protocols` blocks are
// flattened by AzureRM into the properties.customProperties string dictionary
// (special TLS/cipher keys). This cannot be expressed with the declarative rule
// types, so those fields are intentionally omitted here.
//
// Sources:
//   - terraform-provider-azurerm internal/services/apimanagement/api_management_resource.go
//     resource (lines 69-108, timeouts 3h/5m/3h/3h, conditional ForceNew CustomizeDiff),
//     schema (lines 110-744), Create expand (lines 746-1029)
//   - go-azure-sdk apimanagement/2024-05-01/apimanagementservice
//     ApiManagementServiceProperties / ApiManagementServiceSkuProperties /
//     CertificateConfiguration / AdditionalLocation + SkuType/VirtualNetworkType/
//     StoreName/PublicNetworkAccess constants
//   - apimanagement/validate: ApiManagementServiceName, ApiManagementServicePublisherName,
//     ApiManagementServicePublisherEmail, ApimSkuName
//
// TODO: conditional ForceNew (CustomizeDiff) cannot be expressed declaratively:
//   - virtual_network_type: ForceNew unless transitioning from None to Internal/External
//   - virtual_network_configuration: ForceNew when changing an existing subnet
//   - sku_name: ForceNew only when crossing the V2 <-> non-V2 boundary
//
// TODO: additional_location capacity has IntBetween(0,50) in AzureRM, but the ARM
// path properties.additionalLocations[*].sku.capacity is an array-element numeric
// path which IntRules cannot evaluate; not represented.
type ApiManagementService struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ApiManagementService)(nil)

func NewApiManagementService() *ApiManagementService {
	return &ApiManagementService{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ApiManagement/service",
			ApiVersions:  []string{"2024-05-01", "2022-08-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"}, // name (SchemaApiManagementName, ForceNew)
			},
			RequiredFields: []string{
				"properties.publisherName",
				"properties.publisherEmail",
				"sku.name",
				"sku.capacity",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.virtualNetworkType", Value: "None"},       // virtual_network_type default None
				{PropertyPath: "properties.enableClientCertificate", Value: false},   // client_certificate_enabled default false
				{PropertyPath: "properties.disableGateway", Value: false},            // gateway_disabled default false
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},   // public_network_access_enabled default true
			},
			ComputedFields: []string{
				"properties.gatewayUrl",
				"properties.gatewayRegionalUrl",
				"properties.managementApiUrl",
				"properties.portalUrl",
				"properties.developerPortalUrl",
				"properties.scmUrl",
				"properties.publicIPAddresses",
				"properties.privateIPAddresses",
				"properties.outboundPublicIPAddresses",
				"properties.provisioningState",
				"properties.targetProvisioningState",
				"properties.createdAtUtc",
				"properties.platformVersion",
				"properties.developerPortalStatus",
				"properties.legacyPortalStatus",
				"properties.privateEndpointConnections",
			},
			SensitiveFields: []string{
				"properties.certificates[*].encodedCertificate",  // certificate.encoded_certificate (Sensitive)
				"properties.certificates[*].certificatePassword", // certificate.certificate_password (Sensitive)
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 3 * time.Hour,
				Read:   5 * time.Minute,
				Update: 3 * time.Hour,
				Delete: 3 * time.Hour,
			},
			// API Management supports soft-delete; AzureRM recovers/purges via the
			// RecoverSoftDeleted / PurgeSoftDeleteOnDestroy feature flags.
			SoftDelete: true,
			ArrayRules: []azwise.ArrayRule{
				{PropertyPath: "properties.certificates", MaxItems: 10, Message: "at most 10 certificates are supported"},
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── ApiManagementServiceName
				{
					Regex:     `^[0-9a-zA-Z-]{1,50}$`,
					MinLength: 1,
					MaxLength: 50,
					Message:   "may only contain alphanumeric characters and dashes up to 50 characters in length",
				},
				// ── publisher_name → properties.publisherName ── ApiManagementServicePublisherName (max 100)
				{PropertyPath: "properties.publisherName", MinLength: 1, MaxLength: 100, Message: "publisher_name may only be up to 100 characters in length"},
				// ── publisher_email → properties.publisherEmail ── ApiManagementServicePublisherEmail
				{PropertyPath: "properties.publisherEmail", Regex: `^[\S*]{1,100}$`, Message: "publisher_email may only be up to 100 characters in length"},
				// ── sku_name → sku.name ── SkuType enum (capacity carried separately in sku.capacity)
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Basic", "BasicV2", "Consumption", "Developer", "Isolated", "Premium", "Standard", "StandardV2"},
					Message:       "sku.name must be a valid API Management SKU tier",
				},
				// ── virtual_network_type → properties.virtualNetworkType ── VirtualNetworkType enum
				{
					PropertyPath:  "properties.virtualNetworkType",
					AllowedValues: []string{"None", "External", "Internal"},
					Message:       "virtual_network_type must be one of: None, External, Internal",
				},
				// ── public_network_access_enabled → properties.publicNetworkAccess ── PublicNetworkAccess enum
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "publicNetworkAccess must be one of: Enabled, Disabled",
				},
				// ── min_api_version → properties.apiVersionConstraint.minApiVersion ── StringIsNotEmpty
				{PropertyPath: "properties.apiVersionConstraint.minApiVersion", MinLength: 1, Message: "min_api_version must not be empty"},
				// ── certificate.store_name → properties.certificates[*].storeName ── StoreName enum
				{
					PropertyPath:  "properties.certificates[*].storeName",
					AllowedValues: []string{"CertificateAuthority", "Root"},
					Message:       "certificate store_name must be one of: CertificateAuthority, Root",
				},
			},
		},
	}
}

func init() { azwise.Register(NewApiManagementService()) }
