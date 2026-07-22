package devcenter

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DevCenterProjectPool provides resource knowledge for
// Microsoft.DevCenter/projects/pools.
//
// Contributing Terraform resource: azurerm_dev_center_project_pool.
//
// Sources:
//   - terraform-provider-azurerm internal/services/devcenter/dev_center_project_pool_resource.go
//     (schema L55-111, Create body L148-174)
//   - go-azure-sdk resource-manager/devcenter/2025-02-01/pools:
//     id_pool.go (type segment "pools", parent "projects"),
//     model_poolproperties.go, model_stopondisconnectconfiguration.go,
//     constants.go (LocalAdminStatus, SingleSignOnStatus, VirtualNetworkType).
//
// Key mappings:
//   - dev_box_definition_name → properties.devBoxDefinitionName (Required)
//   - dev_center_attached_network_name → properties.networkConnectionName (Required)
//   - local_administrator_enabled (bool) → properties.localAdministrator (Enabled/Disabled)
//   - single_sign_on_enabled (bool) → properties.singleSignOnStatus (Enabled/Disabled)
//   - managed_virtual_network_regions → properties.managedVirtualNetworkRegions (MaxItems 1)
//   - stop_on_disconnect_grace_period_minutes → properties.stopOnDisconnect.gracePeriodMinutes (60-480)
//
// Notes:
//   - name/location/tags are envelope-owned; dev_center_project_id is the parent.
//   - properties.licenseType is hardcoded by AzureRM to "Windows_Client" (server default),
//     properties.virtualNetworkType is derived from managed_virtual_network_regions presence;
//     neither is a user-facing required field.
type DevCenterProjectPool struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DevCenterProjectPool)(nil)

func NewDevCenterProjectPool() *DevCenterProjectPool {
	return &DevCenterProjectPool{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DevCenter/projects/pools",
			ApiVersions:  []string{"2025-02-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath:  "properties.localAdministrator",
					AllowedValues: []string{"Disabled", "Enabled"},
				},
				{
					PropertyPath:  "properties.singleSignOnStatus",
					AllowedValues: []string{"Disabled", "Enabled"},
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.stopOnDisconnect.gracePeriodMinutes",
					MinValue:     azwise.Ptr(int64(60)),
					MaxValue:     azwise.Ptr(int64(480)),
				},
			},
			ArrayRules: []azwise.ArrayRule{
				{PropertyPath: "properties.managedVirtualNetworkRegions", MaxItems: 1},
			},
			DefaultValues: []azwise.DefaultValue{
				// AzureRM defaults single_sign_on_enabled = false.
				{PropertyPath: "properties.singleSignOnStatus", Value: "Disabled"},
			},
			RequiredFields: []string{
				"properties.devBoxDefinitionName",
				"properties.networkConnectionName",
				"properties.localAdministrator",
			},
		},
	}
}

func init() { azwise.Register(NewDevCenterProjectPool()) }
