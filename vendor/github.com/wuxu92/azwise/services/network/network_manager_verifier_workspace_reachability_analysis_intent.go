package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetworkManagerVerifierWorkspaceReachabilityAnalysisIntent provides resource knowledge
// for Microsoft.Network/networkManagers/verifierWorkspaces/reachabilityAnalysisIntents.
//
// Mirrors azurerm_network_manager_verifier_workspace_reachability_analysis_intent. This is
// an immutable resource (AzureRM sdk.Resource with no Update — every argument is ForceNew).
// name and the parent verifier_workspace_id are envelope-owned (Required + RequiresReplace);
// the remaining body fields are listed in ForceNew below.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_manager_verifier_workspace_reachability_analysis_intent_resource.go
//   - go-azure-sdk resource-manager/network/2025-01-01/reachabilityanalysisintents:
//     model_reachabilityanalysisintentproperties.go, model_iptraffic.go, constants.go
//
// Not encoded (deliberate):
//   - source_resource_id / destination_resource_id validate with resource-ID validators
//     (PublicIP, Subnet, VirtualMachine, SqlServer, StorageAccount, CosmosDB) — semantic
//     rules that belong on an azapin customizer, noted only.
//   - ip_traffic.protocols is an array-of-enum body field (ipTraffic.protocols,
//     []NetworkProtocol); azwise StringRule.AllowedValues cannot target an array field, so
//     the per-element enum (Any/ICMP/TCP/UDP) is documented, not emitted.
//   - ip_traffic destination_ports / source_ports validate with validate.IpTrafficPort and
//     destination_ips / source_ips with IsCIDR/IsIPAddress — per-element semantic rules
//     inside arrays; belong on an azapin customizer, noted only.
type NetworkManagerVerifierWorkspaceReachabilityAnalysisIntent struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetworkManagerVerifierWorkspaceReachabilityAnalysisIntent)(nil)

func NewNetworkManagerVerifierWorkspaceReachabilityAnalysisIntent() *NetworkManagerVerifierWorkspaceReachabilityAnalysisIntent {
	return &NetworkManagerVerifierWorkspaceReachabilityAnalysisIntent{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkManagers/verifierWorkspaces/reachabilityAnalysisIntents",
			ApiVersions:  []string{"2025-01-01"},
			// Immutable resource: all body arguments are ForceNew.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.sourceResourceId"},
				{PropertyPath: "properties.destinationResourceId"},
				{PropertyPath: "properties.ipTraffic"},
				{PropertyPath: "properties.description"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.sourceResourceId",
				"properties.destinationResourceId",
				"properties.ipTraffic",
				"properties.ipTraffic.destinationIps",
				"properties.ipTraffic.destinationPorts",
				"properties.ipTraffic.protocols",
				"properties.ipTraffic.sourceIps",
				"properties.ipTraffic.sourcePorts",
			},
		},
	}
}

func init() {
	azwise.Register(NewNetworkManagerVerifierWorkspaceReachabilityAnalysisIntent())
}
