package paloalto

import (
	"time"

	"github.com/wuxu92/azwise"
)

// LocalRulestack provides resource knowledge for PaloAltoNetworks.Cloudngfw/localRulestacks.
//
// Mirrors azurerm_palo_alto_local_rulestack. AzureRM hardcodes defaultMode=NONE and
// scope=LOCAL on create, and defaults every securityServices.* profile to "None"; the
// user-facing profile fields accept "Custom" or "BestPractice". The six security
// profile knobs are plain ARM strings (no SDK enum), so the allowed set is
// Custom/BestPractice/None (None is the provider default).
//
// The outbound_trust/untrust certificate association TF resources have no ARM type of
// their own — they mutate this rulestack's properties.securityServices
// .outboundTrustCertificate / .outboundUnTrustCertificate in place, so no separate
// knowledge file is emitted for them.
//
// Sources:
//   - terraform-provider-azurerm internal/services/paloalto/palo_alto_local_rulestack_resource.go
//     Arguments() L54-126 (name ForceNew envelope; location ForceNew;
//     {vulnerability,anti_spyware,anti_virus,url_filtering,file_blocking}_profile +
//     dns_subscription enum Custom/BestPractice -> properties.securityServices.*;
//     description -> properties.description), Create() L136-210 (defaultMode=NONE,
//     scope=LOCAL hardcoded; securityServices.* default "None"), timeouts 30m/5m/30m/30m
//   - palo_alto_local_rulestack_outbound_trust_certificate_association_resource.go L85-93
//     (folds into properties.securityServices.outboundTrustCertificate)
//   - go-azure-sdk resource-manager/paloaltonetworks/2025-10-08/localrulestackresources
//     model_rulestackproperties.go, model_securityservices.go, constants.go
type LocalRulestack struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*LocalRulestack)(nil)

// NewLocalRulestack returns knowledge for the localRulestacks resource.
func NewLocalRulestack() *LocalRulestack {
	securityProfileValues := []string{"Custom", "BestPractice", "None"}
	securityProfileMsg := "must be one of Custom, BestPractice (None is the service default)"

	return &LocalRulestack{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "PaloAltoNetworks.Cloudngfw/localRulestacks",
			ApiVersions:  []string{"2025-10-08"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// AzureRM hardcodes these on create; a user may still set them via AzAPI.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.defaultMode", Value: "NONE"},
				{PropertyPath: "properties.scope", Value: "LOCAL"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "", // resource name
					Regex:         `^[a-zA-Z0-9]([a-zA-Z0-9-]{0,126}[a-zA-Z0-9])?$`,
					MaxLength:     128,
					Message:       "may only contain alphanumeric characters and dashes, must be 1-128 characters and cannot start or end with a dash",
				},
				{PropertyPath: "properties.securityServices.vulnerabilityProfile", AllowedValues: securityProfileValues, Message: securityProfileMsg},
				{PropertyPath: "properties.securityServices.antiSpywareProfile", AllowedValues: securityProfileValues, Message: securityProfileMsg},
				{PropertyPath: "properties.securityServices.antiVirusProfile", AllowedValues: securityProfileValues, Message: securityProfileMsg},
				{PropertyPath: "properties.securityServices.urlFilteringProfile", AllowedValues: securityProfileValues, Message: securityProfileMsg},
				{PropertyPath: "properties.securityServices.fileBlockingProfile", AllowedValues: securityProfileValues, Message: securityProfileMsg},
				{PropertyPath: "properties.securityServices.dnsSubscription", AllowedValues: securityProfileValues, Message: securityProfileMsg},
				// defaultMode/scope are hardcoded by AzureRM but settable via AzAPI (full ARM enum set).
				{PropertyPath: "properties.defaultMode", AllowedValues: []string{"FIREWALL", "IPS", "NONE"}, Message: "must be one of FIREWALL, IPS, NONE"},
				{PropertyPath: "properties.scope", AllowedValues: []string{"GLOBAL", "LOCAL"}, Message: "must be one of GLOBAL, LOCAL"},
			},
		},
	}
}

func init() { azwise.Register(NewLocalRulestack()) }
