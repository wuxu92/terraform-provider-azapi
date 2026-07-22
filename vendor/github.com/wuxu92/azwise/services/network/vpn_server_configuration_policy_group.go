package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VpnServerConfigurationPolicyGroup provides resource knowledge for
// Microsoft.Network/vpnServerConfigurations/configurationPolicyGroups.
//
// Mirrors azurerm_vpn_server_configuration_policy_group. name and
// vpn_server_configuration_id (parent) are envelope/parent references.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/vpn_server_configuration_policy_group_resource.go
//     (schema 42-100, expand 132-138, timeouts 35-40)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualwans:
//     id_configurationpolicygroup.go:110 (segment casing),
//     model_vpnserverconfigurationpolicygroupproperties.go
//
// Not encoded (deliberate):
//   - policy is a Required set expanding to properties.policyMembers[*]; its
//     per-member enum (type AADGroupId/CertificateGroupId/RadiusAzureGroupId) lives
//     under an array element and cannot be lowered by azwise/azapin.
type VpnServerConfigurationPolicyGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VpnServerConfigurationPolicyGroup)(nil)

// NewVpnServerConfigurationPolicyGroup returns knowledge for the configurationPolicyGroups resource.
func NewVpnServerConfigurationPolicyGroup() *VpnServerConfigurationPolicyGroup {
	return &VpnServerConfigurationPolicyGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/vpnServerConfigurations/configurationPolicyGroups",
			ApiVersions:  []string{"2025-01-01"},
			// is_default is ForceNew; policy and priority are updatable.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.isDefault"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.priority",
					MinValue:     azwise.Ptr(int64(0)),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.isDefault", Value: false},
				{PropertyPath: "properties.priority", Value: float64(0)},
			},
			RequiredFields: []string{
				"properties.policyMembers",
			},
		},
	}
}

func init() { azwise.Register(NewVpnServerConfigurationPolicyGroup()) }
