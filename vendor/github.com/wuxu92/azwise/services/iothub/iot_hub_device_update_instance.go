package iothub

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IotHubDeviceUpdateInstance provides resource knowledge for
// Microsoft.DeviceUpdate/accounts/instances.
//
// Contributing Terraform resource: azurerm_iothub_device_update_instance.
//
// Sources:
//   - terraform-provider-azurerm internal/services/iothub/iothub_device_update_instance_resource.go
//     (schema L40-93, Create body L148-161, timeout L113)
//   - terraform-provider-azurerm internal/services/iothub/validate/iot_hub_device_update_instance_name.go
//     (IotHubDeviceUpdateInstanceName L10-21)
//   - go-azure-sdk resource-manager/deviceupdate/2022-10-01/deviceupdates:
//     model_instanceproperties.go, model_diagnosticstorageproperties.go,
//     model_iothubsettings.go
//
// Notes:
//   - name/device_update_account_id/resource_group_name are envelope / parent-reference fields.
//   - iothub_id maps to properties.iotHubs[*].resourceId (array element) — skipped.
//   - diagnostic_storage_account maps to properties.diagnosticStorageProperties.* :
//     connection_string (Sensitive) -> connectionString, id -> resourceId.
//   - diagnostic_enabled (bool, default false) maps to properties.enableDiagnostics.
type IotHubDeviceUpdateInstance struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IotHubDeviceUpdateInstance)(nil)

// NewIotHubDeviceUpdateInstance returns knowledge for the accounts/instances resource.
func NewIotHubDeviceUpdateInstance() *IotHubDeviceUpdateInstance {
	return &IotHubDeviceUpdateInstance{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DeviceUpdate/accounts/instances",
			ApiVersions:  []string{"2022-10-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[A-Za-z0-9]+(-[A-Za-z0-9]+)*$`,
					MinLength:    3,
					MaxLength:    24,
					Message:      "instance name must be 3-24 characters, start with an alphanumeric, contain only alphanumeric characters and dashes, with no consecutive dashes",
				},
			},
			SensitiveFields: []string{
				"properties.diagnosticStorageProperties.connectionString",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.enableDiagnostics", Value: false},
			},
			ComputedFields: []string{
				"properties.provisioningState",
			},
			RequiredFields: []string{
				"properties.iotHubs",
			},
		},
	}
}

func init() { azwise.Register(NewIotHubDeviceUpdateInstance()) }
