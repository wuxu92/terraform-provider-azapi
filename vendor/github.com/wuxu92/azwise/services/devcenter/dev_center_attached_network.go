package devcenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DevCenterAttachedNetwork provides resource knowledge for
// Microsoft.DevCenter/devCenters/attachedNetworks.
//
// Contributing Terraform resource: azurerm_dev_center_attached_network.
//
// Sources:
//   - terraform-provider-azurerm internal/services/devcenter/dev_center_attached_network_resource.go
//     (schema L43-56)
//   - go-azure-sdk resource-manager/devcenter/2025-02-01/attachednetworkconnections:
//     id_devcenterattachednetwork.go (type segment "attachedNetworks", parent "devCenters"),
//     model_attachednetworkconnectionproperties.go.
//
// Key mappings:
//   - network_connection_id → properties.networkConnectionId (Required, ForceNew)
//
// Notes:
//   - name is envelope-owned; dev_center_id is the parent resource, not a body property.
//   - No Update: every argument is ForceNew.
type DevCenterAttachedNetwork struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DevCenterAttachedNetwork)(nil)

func NewDevCenterAttachedNetwork() *DevCenterAttachedNetwork {
	return &DevCenterAttachedNetwork{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DevCenter/devCenters/attachedNetworks",
			ApiVersions:  []string{"2025-02-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.networkConnectionId"},
			},
			RequiredFields: []string{
				"properties.networkConnectionId",
			},
		},
	}
}

func init() { azwise.Register(NewDevCenterAttachedNetwork()) }
