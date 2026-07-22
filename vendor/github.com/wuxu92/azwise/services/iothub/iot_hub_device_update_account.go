package iothub

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IotHubDeviceUpdateAccount provides resource knowledge for
// Microsoft.DeviceUpdate/accounts.
//
// Contributing Terraform resource: azurerm_iothub_device_update_account.
//
// Sources:
//   - terraform-provider-azurerm internal/services/iothub/iothub_device_update_account_resource.go
//     (schema L36-79, Create body L122-135, timeout L95)
//   - terraform-provider-azurerm internal/services/iothub/validate/iot_hub_device_update_account_name.go
//     (IotHubDeviceUpdateAccountName L10-21)
//   - go-azure-sdk resource-manager/deviceupdate/2022-10-01/deviceupdates:
//     model_account.go, model_accountproperties.go (Sku, PublicNetworkAccess, HostName),
//     constants.go (SKU L366-369, PublicNetworkAccess L284-287)
//
// Notes:
//   - name/location/resource_group_name/identity are envelope fields; not emitted as body rules.
//   - sku maps to properties.sku (enum Free/Standard), is ForceNew, default Standard.
//   - public_network_access_enabled (bool, default true) maps to the enum
//     properties.publicNetworkAccess (Enabled/Disabled).
//   - host_name is read-only (properties.hostName), absent from the Create model.
type IotHubDeviceUpdateAccount struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IotHubDeviceUpdateAccount)(nil)

// NewIotHubDeviceUpdateAccount returns knowledge for the Microsoft.DeviceUpdate/accounts resource.
func NewIotHubDeviceUpdateAccount() *IotHubDeviceUpdateAccount {
	return &IotHubDeviceUpdateAccount{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DeviceUpdate/accounts",
			ApiVersions:  []string{"2022-10-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.sku"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[A-Za-z0-9]+(-[A-Za-z0-9]+)*$`,
					MinLength:    3,
					MaxLength:    24,
					Message:      "account name must be 3-24 characters, start with an alphanumeric, contain only alphanumeric characters and dashes, with no consecutive dashes",
				},
				{
					PropertyPath:  "properties.sku",
					AllowedValues: []string{"Free", "Standard"},
					Message:       "sku must be Free or Standard",
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "publicNetworkAccess must be Enabled or Disabled",
				},
			},
			ComputedFields: []string{
				"properties.hostName",
				"properties.provisioningState",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.sku", Value: "Standard"},
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
			},
		},
	}
}

func init() { azwise.Register(NewIotHubDeviceUpdateAccount()) }
