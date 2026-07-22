package cdn

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CdnFrontDoorCustomDomain provides resource knowledge for
// Microsoft.Cdn/profiles/customDomains (Azure Front Door custom domain).
//
// Contributing Terraform resource: azurerm_cdn_frontdoor_custom_domain.
//
// Sources:
//   - terraform-provider-azurerm internal/services/cdn/cdn_frontdoor_custom_domain_resource.go
//     (schema L68-205, Create body L266-281)
//   - internal/services/cdn/validate/front_door_custom_domain_name.go (name regex)
//   - go-azure-sdk resource-manager/cdn/2025-12-01/afddomains:
//     model_afddomainproperties.go, model_afddomainhttpsparameters.go,
//     model_afddomainhttpscustomizedciphersuiteset.go, model_domainvalidationproperties.go,
//     constants.go (AfdCertificateType / AfdMinimumTlsVersion / AfdCipherSuiteSetType).
//
// Notes:
//   - name & cdn_frontdoor_profile_id are envelope/parent-owned (ForceNew).
//   - host_name is a body ForceNew property (properties.hostName).
//   - dns_zone_id -> properties.azureDnsZone.id and tls.cdn_frontdoor_secret_id ->
//     properties.tlsSettings.secret.id are ResourceReference IDs validated by
//     semantic validators (DnsZoneID / FrontDoorSecretID) at the azapin layer, not here.
//   - tls.cipher_suite.custom_ciphers.tls12/tls13 map to arrays
//     properties.tlsSettings.customizedCipherSuiteSet.cipherSuiteSetForTls12 /
//     cipherSuiteSetForTls13 (array-of-enum) — not expressible as a scalar StringRule,
//     so no declarative rule is emitted (AtLeastOneOf between them is a nested
//     array-element relation, also out of scope).
type CdnFrontDoorCustomDomain struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CdnFrontDoorCustomDomain)(nil)

func NewCdnFrontDoorCustomDomain() *CdnFrontDoorCustomDomain {
	return &CdnFrontDoorCustomDomain{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cdn/profiles/customDomains",
			ApiVersions:  []string{"2025-12-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 12 * time.Hour,
				Read:   5 * time.Minute,
				Update: 24 * time.Hour,
				Delete: 12 * time.Hour,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.hostName"},
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.FrontDoorCustomDomainName
				{
					Regex:     `^[a-zA-Z0-9][a-zA-Z0-9-]{0,258}[a-zA-Z0-9]$`,
					MinLength: 2,
					MaxLength: 260,
					Message:   "must be 2-260 characters, begin and end with a letter or number, and contain only letters, numbers and hyphens",
				},
				// tls.certificate_type → properties.tlsSettings.certificateType
				{
					PropertyPath:  "properties.tlsSettings.certificateType",
					AllowedValues: []string{"AzureFirstPartyManagedCertificate", "CustomerCertificate", "ManagedCertificate"},
				},
				// tls.minimum_version → properties.tlsSettings.minimumTlsVersion
				{
					PropertyPath:  "properties.tlsSettings.minimumTlsVersion",
					AllowedValues: []string{"TLS13", "TLS12", "TLS10"},
				},
				// tls.cipher_suite.type → properties.tlsSettings.cipherSuiteSetType
				{
					PropertyPath:  "properties.tlsSettings.cipherSuiteSetType",
					AllowedValues: []string{"Customized", "TLS12_2023", "TLS12_2022", "TLS10_2019"},
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.tlsSettings.certificateType", Value: "ManagedCertificate"},
				{PropertyPath: "properties.tlsSettings.minimumTlsVersion", Value: "TLS12"},
			},
			ComputedFields: []string{
				"properties.deploymentStatus",
				"properties.provisioningState",
				"properties.domainValidationState",
				"properties.profileName",
				"properties.validationProperties",
			},
			// tls block is Required; certificateType is a required (non-pointer) SDK
			// field that AzureRM always sends (defaulted to ManagedCertificate).
			RequiredFields: []string{
				"properties.hostName",
				"properties.tlsSettings",
				"properties.tlsSettings.certificateType",
			},
		},
	}
}

func init() { azwise.Register(NewCdnFrontDoorCustomDomain()) }
