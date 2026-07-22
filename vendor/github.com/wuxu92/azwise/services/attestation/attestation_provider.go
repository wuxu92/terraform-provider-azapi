package attestation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AttestationProvider provides resource knowledge for
// Microsoft.Attestation/attestationProviders.
//
// Mirrors azurerm_attestation_provider. name/location/resource_group_name are
// operational-envelope fields; name is ForceNew in AzureRM but is not an ARM-body
// property, so it is not encoded as a body ForceNew rule here.
//
// The only management-plane body input is properties.policySigningCertificates,
// which AzureRM builds from the PEM in policy_signing_certificate_data (ForceNew).
// That argument is a PEM string that AzureRM decodes and wraps into a JSONWebKeySet
// object, so it is validated at the schema/customizer layer (validate.IsCert) rather
// than as a declarative StringRule; the ARM path itself is still ForceNew.
//
// The open_enclave_policy_base64 / sgx_enclave_policy_base64 / tpm_policy_base64 /
// sev_snp_policy_base64 arguments are DATA-PLANE settings (applied through the
// attestation data-plane client, not the ARM PUT body) and have no management-plane
// ARM path, so they are intentionally not represented here. AzureRM's CustomizeDiff
// preventing their removal is likewise a data-plane concern that azwise cannot map.
//
// Sources:
//   - terraform-provider-azurerm internal/services/attestation/attestation_provider_resource.go
//     schema + Create/Update (policy_signing_certificate_data ForceNew; policy_*_base64
//     applied via data-plane client; timeouts 30m/5m/30m/30m)
//   - go-azure-sdk resource-manager/attestation/2020-10-01/attestationproviders
//     AttestationServiceCreationParams.Properties.PolicySigningCertificates is the only
//     settable body field; StatusResult (attestUri/trustModel/status/
//     privateEndpointConnections) is read-only.
type AttestationProvider struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AttestationProvider)(nil)

// NewAttestationProvider returns knowledge for the attestationProviders resource.
func NewAttestationProvider() *AttestationProvider {
	return &AttestationProvider{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Attestation/attestationProviders",
			ApiVersions:  []string{"2020-10-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// AzureRM marks policy_signing_certificate_data ForceNew; it maps to the
			// policySigningCertificates JSONWebKeySet in the creation body.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.policySigningCertificates"},
			},
			// Read-only status properties returned by GET but absent from the create body.
			ComputedFields: []string{
				"properties.attestUri",
				"properties.trustModel",
				"properties.status",
				"properties.privateEndpointConnections",
			},
		},
	}
}

func init() { azwise.Register(NewAttestationProvider()) }
