package cdn

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CdnEndpointCustomDomain provides resource knowledge for
// Microsoft.Cdn/profiles/endpoints/customDomains (classic CDN custom domain).
//
// Contributing Terraform resource: azurerm_cdn_endpoint_custom_domain.
//
// Sources:
//   - AzureRM internal/services/cdn/cdn_endpoint_custom_domain_resource.go:28-169
//     (schema: ForceNew name/host_name, name validator, timeouts, https blocks)
//   - AzureRM internal/services/cdn/cdn_endpoint_custom_domain_resource.go:171-246
//     (create expand: only properties.hostName is written to the resource body)
//   - AzureRM internal/services/cdn/validate/custom_domain.go:12 (name regex)
//   - azure-sdk-for-go services/cdn/mgmt/2020-09-01/cdn/models.go:2660
//     (CustomDomainPropertiesParameters json tags)
//
// Notes on constraints that cannot be expressed declaratively:
//   - The `cdn_managed_https` and `user_managed_https` blocks (certificate_type,
//     protocol_type, tls_version, key_vault_secret_id) are NOT part of the
//     customDomain resource body. AzureRM applies them via the separate
//     EnableCustomHTTPS POST action (enableArmCdnEndpointCustomDomainHttps), so their
//     enums and the cdn_managed_https<->user_managed_https ConflictsWith relation live
//     on that action, not on this resource — omitted here rather than emitted against
//     non-existent body paths.
type CdnEndpointCustomDomain struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CdnEndpointCustomDomain)(nil)

// NewCdnEndpointCustomDomain returns a CdnEndpointCustomDomain knowledge instance.
func NewCdnEndpointCustomDomain() *CdnEndpointCustomDomain {
	return &CdnEndpointCustomDomain{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cdn/profiles/endpoints/customDomains",
			ApiVersions:  []string{"2020-09-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.hostName"},
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 12 * time.Hour,
				Read:   5 * time.Minute,
				Update: 24 * time.Hour,
				Delete: 12 * time.Hour,
			},
			StringRules: []azwise.StringRule{
				{
					// resource name — validate.CdnEndpointCustomDomainName() StringMatch.
					PropertyPath: "",
					Regex:        `^[a-zA-Z0-9]+(-*[a-zA-Z0-9])*$`,
					Message:      "must be alphanumeric and may contain interior hyphens",
				},
				{
					// host_name — validation.StringIsNotEmpty.
					PropertyPath: "properties.hostName",
					MinLength:    1,
					Message:      "host name must not be empty",
				},
			},
			RequiredFields: []string{"properties.hostName"},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewCdnEndpointCustomDomain()) }
