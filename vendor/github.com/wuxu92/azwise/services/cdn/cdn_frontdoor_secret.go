package cdn

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CdnFrontDoorSecret provides resource knowledge for
// Microsoft.Cdn/profiles/secrets (Azure Front Door secret).
//
// Contributing Terraform resource: azurerm_cdn_frontdoor_secret.
//
// Sources:
//   - terraform-provider-azurerm internal/services/cdn/cdn_frontdoor_secret_resource.go
//     (schema L43-97, Create body L131-135)
//   - internal/services/cdn/validate/front_door_validation_helpers.go (CdnFrontDoorSecretName)
//   - internal/services/cdn/frontdoorsecretparams/cdn_frontdoor_secret_params.go
//     (polymorphic SecretParameters expansion).
//
// Notes:
//   - name & cdn_frontdoor_profile_id are envelope/parent-owned (ForceNew).
//   - The whole `secret` block is ForceNew and maps to the polymorphic
//     properties.parameters object (discriminated by `type`, e.g. CustomerCertificate).
//   - secret.customer_certificate.key_vault_certificate_id is a Key Vault nested-item
//     URL validated by a semantic validator (keyvault.ValidateNestedItemID) at the
//     azapin layer; it expands into properties.parameters.secretSource, not a scalar
//     body field, so no declarative rule is emitted.
type CdnFrontDoorSecret struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CdnFrontDoorSecret)(nil)

func NewCdnFrontDoorSecret() *CdnFrontDoorSecret {
	return &CdnFrontDoorSecret{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Cdn/profiles/secrets",
			ApiVersions:  []string{"2025-12-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 4 * time.Hour,
				Read:   5 * time.Minute,
				Update: 4 * time.Hour,
				Delete: 6 * time.Hour,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.parameters"},
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.CdnFrontDoorSecretName
				{
					Regex:     `^[a-zA-Z0-9][a-zA-Z0-9-]{0,258}[a-zA-Z0-9]$`,
					MinLength: 2,
					MaxLength: 260,
					Message:   "must be 2-260 characters, begin and end with a letter or number, and contain only letters, numbers and hyphens",
				},
			},
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.deploymentStatus",
				"properties.profileName",
			},
			RequiredFields: []string{"properties.parameters"},
		},
	}
}

func init() { azwise.Register(NewCdnFrontDoorSecret()) }
