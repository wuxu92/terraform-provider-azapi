package arcresourcebridge

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ArcResourceBridgeAppliance provides resource knowledge for
// Microsoft.ResourceConnector/appliances.
//
// Mirrors azurerm_arc_resource_bridge_appliance. Envelope fields (name,
// resource_group_name, location, identity) live on the operational envelope;
// identity is system-assigned Required + ForceNew and is not repeated here.
//
// Notes:
//   - distro is Required with a single allowed value (AKSEdge) but is NOT ForceNew.
//   - infrastructure_provider (Required + ForceNew) maps to
//     properties.infrastructureConfig.provider.
//   - public_key_base64 (Optional + ForceNew) maps to properties.publicKey; AzureRM
//     sets it via a follow-up update after create, but it is the same ARM body field.
//   - name's Base64EncodedString-style character validator is a StringMatch and is
//     expressed as a Regex rule; the length bound (1..260) is added via MinLength/MaxLength.
//
// Sources:
//   - terraform-provider-azurerm internal/services/arcresourcebridge/arc_resource_bridge_appliance_resource.go
//     (schema L41-87, create L101-161; distro StringInSlice; infrastructure_provider
//     StringInSlice+ForceNew; public_key_base64 ForceNew; timeouts 60m/5m/30m/30m)
//   - go-azure-sdk resource-manager/resourceconnector/2022-10-27/appliances
//     ApplianceProperties (distro/infrastructureConfig.provider/publicKey settable;
//     provisioningState/status/version server-computed); Distro = AKSEdge;
//     Provider = HCI|SCVMM|VMWare
type ArcResourceBridgeAppliance struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ArcResourceBridgeAppliance)(nil)

// NewArcResourceBridgeAppliance returns knowledge for the appliances resource.
func NewArcResourceBridgeAppliance() *ArcResourceBridgeAppliance {
	return &ArcResourceBridgeAppliance{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ResourceConnector/appliances",
			ApiVersions:  []string{"2022-10-27"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.infrastructureConfig.provider"},
				{PropertyPath: "properties.publicKey"},
			},
			StringRules: []azwise.StringRule{
				// name (StringLenBetween(1,260) + StringMatch).
				{MinLength: 1, MaxLength: 260, Regex: `[^+#%&'?/,%\\]+$`, Message: "name must be 1-260 characters and may not contain any of '+', '#', '%', '&', '\\'', '?', '/', ',', '\\'"},
				// distro (StringInSlice).
				{PropertyPath: "properties.distro", AllowedValues: []string{"AKSEdge"}},
				// infrastructure_provider (StringInSlice).
				{PropertyPath: "properties.infrastructureConfig.provider", AllowedValues: []string{"HCI", "SCVMM", "VMWare"}},
			},
			RequiredFields: []string{
				"properties.distro",
				"properties.infrastructureConfig.provider",
			},
			// Server-computed read-only properties surfaced by AzureRM.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.status",
				"properties.version",
			},
		},
	}
}

func init() { azwise.Register(NewArcResourceBridgeAppliance()) }
