package network

import (
	"strings"
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkManagerConnectivityConfiguration provides resource knowledge for
// Microsoft.Network/networkManagers/connectivityConfigurations.
//
// Mirrors azurerm_network_manager_connectivity_configuration. name and the parent
// network_manager_id are envelope-owned (Required + RequiresReplace).
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_manager_connectivity_configuration_resource.go
//     (schema Arguments, CustomizeDiff, expand functions)
//   - go-azure-sdk resource-manager/network/2025-01-01/connectivityconfigurations:
//     model_connectivityconfigurationproperties.go,
//     model_connectivityconfigurationpropertiesconnectivitycapabilities.go, constants.go
//
// Not encoded (deliberate):
//   - applies_to_group.group_connectivity is an array-element enum
//     (appliesToGroups[*].connectivityGroupItem...); azwise cannot lower a path through an
//     array element, so its None/DirectlyConnected enum is documented, not emitted.
//   - CustomizeDiff forbids peering_enforcement_enabled when connectivity_topology is Mesh
//     — a cross-field semantic rule with no representable single ARM body path; it belongs
//     on an azapin customizer, noted here only.
//   - connected_group_address_overlap_enabled is conditionally ForceNew (only Allowed →
//     Disallowed); handled by the CheckForceNew override below rather than an
//     unconditional ForceNew rule.
type NetworkManagerConnectivityConfiguration struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkManagerConnectivityConfiguration)(nil)

func NewNetworkManagerConnectivityConfiguration() *NetworkManagerConnectivityConfiguration {
	return &NetworkManagerConnectivityConfiguration{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkManagers/connectivityConfigurations",
			ApiVersions:  []string{"2025-01-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.connectivityTopology",
					AllowedValues: []string{"HubAndSpoke", "Mesh"},
					Message:       "must be HubAndSpoke or Mesh",
				},
				{
					PropertyPath:  "properties.connectivityCapabilities.connectedGroupPrivateEndpointsScale",
					AllowedValues: []string{"HighScale", "Standard"},
					Message:       "must be HighScale or Standard",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.connectivityCapabilities.connectedGroupPrivateEndpointsScale", Value: "Standard"},
				// connected_group_address_overlap_enabled defaults true → Allowed.
				{PropertyPath: "properties.connectivityCapabilities.connectedGroupAddressOverlap", Value: "Allowed"},
				// peering_enforcement_enabled defaults false → Unenforced.
				{PropertyPath: "properties.connectivityCapabilities.peeringEnforcement", Value: "Unenforced"},
			},
			// applies_to_group (Required) and connectivity_topology (Required).
			RequiredFields: []string{
				"properties.appliesToGroups",
				"properties.connectivityTopology",
			},
		},
	}
}

// CheckForceNew extends the declarative rules with the conditional ForceNew AzureRM
// enforces via CustomizeDiff: changing connectedGroupAddressOverlap from Allowed to
// Disallowed forces replacement, but the reverse is an in-place update.
func (c *NetworkManagerConnectivityConfiguration) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
	if c.BaseKnowledge.CheckForceNew(oldBody, newBody) {
		return true
	}
	const path = "properties.connectivityCapabilities.connectedGroupAddressOverlap"
	oldVal := azwise.ExtractStringValue(oldBody, path)
	newVal := azwise.ExtractStringValue(newBody, path)
	if strings.EqualFold(oldVal, "Allowed") && strings.EqualFold(newVal, "Disallowed") {
		return true
	}
	return false
}

func init() { azwise.Register(NewNetworkManagerConnectivityConfiguration()) }
