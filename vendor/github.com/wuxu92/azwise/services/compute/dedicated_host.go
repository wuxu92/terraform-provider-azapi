package compute

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DedicatedHost provides resource knowledge for Microsoft.Compute/hostGroups/hosts.
//
// Contributing Terraform resource: azurerm_dedicated_host.
//
// Sources:
//   - terraform-provider-azurerm internal/services/compute/dedicated_host_resource.go
//     (schema L52-158, Create body L186-201)
//   - terraform-provider-azurerm internal/services/compute/validate/dedicated_host_name.go
//     (delegates to dedicated_host_group_name.go regex)
//   - go-azure-sdk resource-manager/compute/2024-03-01/dedicatedhosts:
//     model_dedicatedhost.go, model_dedicatedhostproperties.go, model_sku.go, constants.go.
//
// Notes:
//   - name/location are envelope-owned; dedicated_host_group_id is the parent
//     reference (envelope) — none are body fields.
//   - sku_name maps to the top-level ARM `sku.name`, not a properties.* field.
type DedicatedHost struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DedicatedHost)(nil)

func NewDedicatedHost() *DedicatedHost {
	return &DedicatedHost{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Compute/hostGroups/hosts",
			ApiVersions:  []string{"2024-03-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "sku.name"},                       // sku_name
				{PropertyPath: "properties.platformFaultDomain"}, // platform_fault_domain
			},
			RequiredFields: []string{
				"sku.name",
				"properties.platformFaultDomain",
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.DedicatedHostName (→ DedicatedHostGroupName regex)
				{
					Regex:   `^[^_\W][\w-.]{0,78}[\w]$`,
					Message: "must be 2-80 chars, contain only word characters, hyphens and periods, and not start with an underscore",
				},
				// sku_name → sku.name ── StringInSlice (SKU family/type identifiers)
				{
					PropertyPath: "sku.name",
					AllowedValues: []string{
						"DADSv5-Type1", "DASv4-Type1", "DASv4-Type2", "DASv5-Type1", "DCSv2-Type1",
						"DDSv4-Type1", "DDSv4-Type2", "DDSv5-Type1", "DSv3-Type1", "DSv3-Type2",
						"DSv3-Type3", "DSv3-Type4", "DSv4-Type1", "DSv4-Type2", "DSv5-Type1",
						"EADSv5-Type1", "EASv4-Type1", "EASv4-Type2", "EASv5-Type1", "EDSv4-Type1",
						"EDSv4-Type2", "EDSv5-Type1", "ESv3-Type1", "ESv3-Type2", "ESv3-Type3",
						"ESv3-Type4", "ESv4-Type1", "ESv4-Type2", "ESv5-Type1", "FSv2-Type2",
						"FSv2-Type3", "FSv2-Type4", "FXmds-Type1", "LSv2-Type1", "LSv3-Type1",
						"MDMSv2MedMem-Type1", "MDSv2MedMem-Type1", "MMSv2MedMem-Type1", "MS-Type1",
						"MSm-Type1", "MSmv2-Type1", "MSv2-Type1", "MSv2MedMem-Type1", "NVASv4-Type1",
						"NVSv3-Type1",
					},
					Message: "must be a supported dedicated host SKU identifier",
				},
				// license_type → properties.licenseType ── full ARM SDK enum set
				{
					PropertyPath:  "properties.licenseType",
					AllowedValues: []string{"None", "Windows_Server_Hybrid", "Windows_Server_Perpetual"},
					Message:       "must be one of None, Windows_Server_Hybrid, or Windows_Server_Perpetual",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.autoReplaceOnFailure", Value: true},
				{PropertyPath: "properties.licenseType", Value: "None"},
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDedicatedHost()) }
