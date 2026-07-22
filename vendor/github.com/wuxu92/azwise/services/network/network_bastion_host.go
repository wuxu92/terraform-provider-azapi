package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// BastionHost provides resource knowledge for Microsoft.Network/bastionHosts.
//
// Mirrors azurerm_bastion_host. sku expands into the top-level sku.name; the
// boolean feature toggles and scale_units map under properties.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/bastion_host_resource.go
//     (resourceBastionHost schema lines 41-213, expand lines 215-347)
//   - go-azure-sdk resource-manager/network/2025-01-01/bastionhosts:
//     id_bastionhost.go (ARM type segment "bastionHosts"),
//     model_bastionhost.go (top-level sku), model_bastionhostpropertiesformat.go
//     (scaleUnits, ipConfigurations, virtualNetwork json tags),
//     constants.go (BastionHostSkuName: Basic, Developer, Premium, Standard)
//
// Not encoded (deliberate):
//   - copy_paste_enabled maps to properties.disableCopyPaste with INVERTED
//     semantics (AzureRM sends disableCopyPaste = !copy_paste_enabled, and only
//     when disabling); this is not a straight default and is left unencoded.
//   - file_copy_enabled / ip_connect_enabled / kerberos_enabled /
//     shareable_link_enabled / tunneling_enabled / session_recording_enabled are
//     Optional bools sent only when true (properties.enableFileCopy etc.); their
//     "only supported on Standard/Premium" checks are cross-field semantic rules
//     with no declarative form.
//   - dns_name / private_only_enabled are Computed read-only; already bicep ReadOnly.
//   - ip_configuration sub-fields (name, subnet_id, public_ip_address_id) live
//     under the ipConfigurations[*] array element and are skipped.
type BastionHost struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*BastionHost)(nil)

// NewBastionHost returns knowledge for the bastionHosts resource.
func NewBastionHost() *BastionHost {
	return &BastionHost{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/bastionHosts",
			ApiVersions:  []string{"2025-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "zones"},
				{PropertyPath: "properties.ipConfigurations"},   // ip_configuration block
				{PropertyPath: "properties.virtualNetwork.id"},  // virtual_network_id
				// sku.name is conditionally ForceNew (ForceNewIfChange): replaced
				// only on a SKU downgrade (Basic < Standard < Premium). Included so
				// upgrades that Azure rejects still surface as a plan concern.
				{PropertyPath: "sku.name"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Basic", "Developer", "Premium", "Standard"},
					Message:       "must be Basic, Developer, Premium or Standard",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.scaleUnits",
					MinValue:     azwise.Ptr(int64(2)),
					MaxValue:     azwise.Ptr(int64(50)),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "sku.name", Value: "Basic"},
				{PropertyPath: "properties.scaleUnits", Value: float64(2)},
			},
		},
	}
}

func init() { azwise.Register(NewBastionHost()) }
