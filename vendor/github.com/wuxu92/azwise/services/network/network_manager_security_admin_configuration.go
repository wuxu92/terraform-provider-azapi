package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkManagerSecurityAdminConfiguration provides resource knowledge for
// Microsoft.Network/networkManagers/securityAdminConfigurations.
//
// Mirrors azurerm_network_manager_security_admin_configuration. name and the parent
// network_manager_id are envelope-owned (Required + RequiresReplace).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_manager_security_admin_configuration_resource.go
//   - go-azure-sdk resource-manager/network/2025-01-01/securityadminconfigurations:
//     model_securityadminconfigurationpropertiesformat.go, constants.go
//
// Not encoded (deliberate):
//   - apply_on_network_intent_policy_based_services is an array-of-enum body field
//     (properties.applyOnNetworkIntentPolicyBasedServices, []NetworkIntentPolicyBasedService);
//     azwise StringRule.AllowedValues cannot target an array field, so the per-element enum
//     (All/AllowRulesOnly/None) is documented, not emitted. The MaxItems:1 constraint IS
//     emitted as an ArrayRule below.
type NetworkManagerSecurityAdminConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkManagerSecurityAdminConfiguration)(nil)

func NewNetworkManagerSecurityAdminConfiguration() *NetworkManagerSecurityAdminConfiguration {
	return &NetworkManagerSecurityAdminConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkManagers/securityAdminConfigurations",
			ApiVersions:  []string{"2025-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ArrayRules: []azwise.ArrayRule{
				{
					PropertyPath: "properties.applyOnNetworkIntentPolicyBasedServices",
					MaxItems:     1,
				},
			},
		},
	}
}

func init() { azwise.Register(NewNetworkManagerSecurityAdminConfiguration()) }
