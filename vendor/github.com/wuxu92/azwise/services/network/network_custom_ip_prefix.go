package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CustomIPPrefix provides resource knowledge for Microsoft.Network/customIPPrefixes.
//
// Mirrors azurerm_custom_ip_prefix. cidr expands into properties.cidr; the
// parent reference and validation messages map under properties. Create/Delete
// run long (BYOIP commissioning), reflected in the timeouts.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/custom_ip_prefix_resource.go
//     (Arguments lines 59-140, Create expand lines 204-232, timeouts 9h/17h)
//   - go-azure-sdk resource-manager/network/2025-01-01/customipprefixes:
//     id_customipprefix.go (ARM type segment "customIPPrefixes"),
//     model_customipprefixpropertiesformat.go (cidr, customIpPrefixParent,
//     signedMessage, authorizationMessage json tags)
//
// Not encoded (deliberate):
//   - commissioning_enabled / internet_advertising_disabled drive a
//     CommissionedState state machine (properties.commissionedState transitions);
//     they are procedural, not a single settable body value, so left unencoded.
//   - roa_validity_end_date is combined with subscription id and cidr into
//     properties.authorizationMessage (a computed composite string), not a direct
//     mapping.
//   - cidr carries a net.ParseCIDR ValidateFunc and roa_validity_end_date a date
//     parse; these are semantic validators without a declarative regex here.
type CustomIPPrefix struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CustomIPPrefix)(nil)

// NewCustomIPPrefix returns knowledge for the customIPPrefixes resource.
func NewCustomIPPrefix() *CustomIPPrefix {
	return &CustomIPPrefix{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/customIPPrefixes",
			ApiVersions:  []string{"2025-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "properties.cidr"},
				{PropertyPath: "properties.customIpPrefixParent.id"}, // parent_custom_ip_prefix_id
				{PropertyPath: "properties.signedMessage"},           // wan_validation_signed_message
				{PropertyPath: "properties.authorizationMessage"},    // roa_validity_end_date (composite)
				{PropertyPath: "zones"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 9 * time.Hour,
				Read:   5 * time.Minute,
				Update: 17 * time.Hour,
				Delete: 17 * time.Hour,
			},
			RequiredFields: []string{
				"properties.cidr",
			},
		},
	}
}

func init() { azwise.Register(NewCustomIPPrefix()) }
