package paloalto

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Firewall provides resource knowledge for PaloAltoNetworks.Cloudngfw/firewalls.
//
// This ONE file is the union of the six AzureRM next-generation-firewall TF resources
// that all map to the firewalls ARM type:
//   - azurerm_palo_alto_next_generation_firewall_virtual_hub_local_rulestack
//   - azurerm_palo_alto_next_generation_firewall_virtual_hub_panorama
//   - azurerm_palo_alto_next_generation_firewall_virtual_hub_strata_cloud_manager
//   - azurerm_palo_alto_next_generation_firewall_virtual_network_local_rulestack
//   - azurerm_palo_alto_next_generation_firewall_virtual_network_panorama
//   - azurerm_palo_alto_next_generation_firewall_virtual_network_strata_cloud_manager
//
// Only UNIVERSAL knowledge (true for every firewall body regardless of kind) is
// emitted. Kind-specific shapes are deliberately NOT unioned, because doing so would
// corrupt validation for the other kinds:
//   - networkProfile.networkType value: VWAN for virtual_hub_*, VNET for
//     virtual_network_* (the enum StringRule below is safe — it only constrains the
//     allowed set, never forces a specific value).
//   - properties.associatedRulestack (local_rulestack kinds only),
//     properties.panoramaConfig (panorama kinds only),
//     properties.strataCloudManagerConfig (strata kinds only) — none are required or
//     ForceNew across all kinds, so they carry no shared rule.
//
// Universal facts (hardcoded identically by AzureRM for all six kinds):
//   marketplaceDetails.publisherId = "paloaltonetworks", planData.billingCycle = MONTHLY,
//   marketplace_offer_id default "pan_swfw_cloud_ngfw" (ForceNew), plan_id default
//   "panw-cngfw-payg".
//
// Sources (terraform-provider-azurerm internal/services/paloalto/):
//   - palo_alto_next_generation_firewall_virtual_hub_local_rulestack_resource.go
//     Arguments() L56-103, Create() L109-181 (marketplaceDetails.publisherId="paloaltonetworks",
//     planData.billingCycle=MONTHLY)
//   - palo_alto_next_generation_firewall_virtual_hub_panorama_resource.go L74-103,144-153
//   - palo_alto_next_generation_firewall_virtual_hub_strata_cloud_manager_resource.go L74-85,141-145
//   - palo_alto_next_generation_firewall_virtual_network_local_rulestack_resource.go L71-93,152-156
//   - palo_alto_next_generation_firewall_virtual_network_panorama_resource.go L74-96,144-148
//   - palo_alto_next_generation_firewall_virtual_network_strata_cloud_manager_resource.go L74-85,143-147
//   - name ForceNew (validate.NextGenerationFirewallName ^[a-zA-Z0-9-]{1,128}$, no
//     leading/trailing dash); marketplace_offer_id ForceNew (StringIsNotEmpty);
//     plan_id StringLenBetween(1,50); timeouts widest across kinds 3h/5m/3h/3h
//   - go-azure-sdk resource-manager/paloaltonetworks/2025-10-08/firewallresources
//     model_firewalldeploymentproperties.go, model_marketplacedetails.go,
//     model_plandata.go, model_networkprofile.go, model_dnssettings.go, constants.go
//     (BillingCycle, NetworkType, EgressNat, DNSProxy, EnabledDNSType)
type Firewall struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Firewall)(nil)

// NewFirewall returns knowledge for the firewalls resource.
func NewFirewall() *Firewall {
	return &Firewall{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "PaloAltoNetworks.Cloudngfw/firewalls",
			ApiVersions:  []string{"2025-10-08"},
			ForceNew: []azwise.ForceNewRule{
				// marketplace_offer_id is ForceNew for all six kinds.
				{PropertyPath: "properties.marketplaceDetails.offerId"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 3 * time.Hour,
				Read:   5 * time.Minute,
				Update: 3 * time.Hour,
				Delete: 3 * time.Hour,
			},
			// Required in the ARM model (non-pointer) for every firewall kind.
			RequiredFields: []string{
				"properties.marketplaceDetails.offerId",
				"properties.marketplaceDetails.publisherId",
				"properties.planData.planId",
				"properties.planData.billingCycle",
				"properties.networkProfile.networkType",
				"properties.networkProfile.enableEgressNat",
				"properties.networkProfile.publicIps",
			},
			// Values AzureRM hardcodes/defaults identically for all six kinds.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.marketplaceDetails.offerId", Value: "pan_swfw_cloud_ngfw"},
				{PropertyPath: "properties.marketplaceDetails.publisherId", Value: "paloaltonetworks"},
				{PropertyPath: "properties.planData.planId", Value: "panw-cngfw-payg"},
				{PropertyPath: "properties.planData.billingCycle", Value: "MONTHLY"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "", // resource name
					Regex:        `^[a-zA-Z0-9]([a-zA-Z0-9-]{0,126}[a-zA-Z0-9])?$`,
					MaxLength:    128,
					Message:      "may only contain alphanumeric characters and dashes, must be 1-128 characters and cannot start or end with a dash",
				},
				{
					PropertyPath: "properties.marketplaceDetails.offerId",
					MinLength:    1,
					Message:      "marketplace offer id must not be empty",
				},
				{
					PropertyPath: "properties.planData.planId",
					MinLength:    1,
					MaxLength:    50,
					Message:      "plan id must be 1-50 characters",
				},
				{
					PropertyPath:  "properties.planData.billingCycle",
					AllowedValues: []string{"MONTHLY", "WEEKLY"},
					Message:       "must be one of MONTHLY, WEEKLY",
				},
				{
					PropertyPath:  "properties.networkProfile.networkType",
					AllowedValues: []string{"VNET", "VWAN"},
					Message:       "must be one of VNET, VWAN",
				},
				{
					PropertyPath:  "properties.networkProfile.enableEgressNat",
					AllowedValues: []string{"ENABLED", "DISABLED"},
					Message:       "must be one of ENABLED, DISABLED",
				},
				{
					PropertyPath:  "properties.dnsSettings.enableDnsProxy",
					AllowedValues: []string{"ENABLED", "DISABLED"},
					Message:       "must be one of ENABLED, DISABLED",
				},
				{
					PropertyPath:  "properties.dnsSettings.enabledDnsType",
					AllowedValues: []string{"AZURE", "CUSTOM"},
					Message:       "must be one of AZURE, CUSTOM",
				},
			},
		},
	}
}

func init() { azwise.Register(NewFirewall()) }
