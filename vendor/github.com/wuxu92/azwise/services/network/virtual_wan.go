package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// VirtualWan provides resource knowledge for Microsoft.Network/virtualWans.
//
// Mirrors azurerm_virtual_wan. name and resource_group_name are envelope-owned
// (Required + RequiresReplace); only body/location knowledge is encoded here.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/virtual_wan_resource.go
//     (resourceVirtualWan schema lines 43-86, expand lines 111-120, timeouts 36-41)
//   - go-azure-sdk resource-manager/network/2025-01-01/virtualwans:
//     model_virtualwanproperties.go, constants.go (OfficeTrafficCategory 1080-1094)
type VirtualWan struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*VirtualWan)(nil)

// NewVirtualWan returns knowledge for the virtualWans resource.
func NewVirtualWan() *VirtualWan {
	return &VirtualWan{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/virtualWans",
			ApiVersions:  []string{"2025-01-01"},
			// location (commonschema.Location) replaces the resource on change.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.office365LocalBreakoutCategory",
					AllowedValues: []string{"All", "None", "Optimize", "OptimizeAndAllow"},
					Message:       "must be one of All, None, Optimize, OptimizeAndAllow",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.disableVpnEncryption", Value: false},
				{PropertyPath: "properties.allowBranchToBranchTraffic", Value: true},
				{PropertyPath: "properties.office365LocalBreakoutCategory", Value: "None"},
				{PropertyPath: "properties.type", Value: "Standard"},
			},
		},
	}
}

func init() { azwise.Register(NewVirtualWan()) }
