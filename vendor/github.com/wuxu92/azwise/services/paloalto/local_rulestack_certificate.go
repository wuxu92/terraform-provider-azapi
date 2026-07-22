package paloalto

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LocalRulestackCertificate provides resource knowledge for
// PaloAltoNetworks.Cloudngfw/localRulestacks/certificates.
//
// Mirrors azurerm_palo_alto_local_rulestack_certificate. certificateSelfSigned is a
// required BooleanEnum ("TRUE"/"FALSE") and certificateSignerResourceId is ForceNew.
// AzureRM enforces ExactlyOneOf(self_signed, key_vault_certificate_id) at the schema
// level; that maps to "certificateSelfSigned==TRUE XOR certificateSignerResourceId set",
// which is not expressible as an azwise RelationalRule because certificateSelfSigned is
// always present in the body (documented, not emitted).
//
// Sources:
//   - terraform-provider-azurerm internal/services/paloalto/palo_alto_local_rulestack_certificate_resource.go
//     Arguments() L45-92 (name -> ruleName segment, not ForceNew; rulestack_id ForceNew
//     parent; audit_comment -> properties.auditComment; description -> properties.description;
//     key_vault_certificate_id ForceNew -> properties.certificateSignerResourceId;
//     self_signed ForceNew -> properties.certificateSelfSigned), Create() L102-169,
//     timeouts 30m/5m/30m/30m
//   - go-azure-sdk resource-manager/paloaltonetworks/2025-10-08/certificateobjectlocalrulestackresources
//     model_certificateobject.go, constants.go (BooleanEnum)
type LocalRulestackCertificate struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LocalRulestackCertificate)(nil)

// NewLocalRulestackCertificate returns knowledge for the certificates resource.
func NewLocalRulestackCertificate() *LocalRulestackCertificate {
	return &LocalRulestackCertificate{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "PaloAltoNetworks.Cloudngfw/localRulestacks/certificates",
			ApiVersions:  []string{"2025-10-08"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.certificateSelfSigned"},
				{PropertyPath: "properties.certificateSignerResourceId"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.certificateSelfSigned",
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "", // resource name
					Regex:        `^[a-zA-Z0-9]([a-zA-Z0-9-]{0,126}[a-zA-Z0-9])?$`,
					MaxLength:    128,
					Message:      "may only contain alphanumeric characters and dashes, must be 1-128 characters and cannot start or end with a dash",
				},
				{
					PropertyPath:  "properties.certificateSelfSigned",
					AllowedValues: []string{"TRUE", "FALSE"},
					Message:       "must be one of TRUE, FALSE",
				},
			},
		},
	}
}

func init() { azwise.Register(NewLocalRulestackCertificate()) }
