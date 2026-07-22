package iothub

import (
	"time"

	"github.com/wuxu92/azwise"
)

// IotHubDps provides resource knowledge for Microsoft.Devices/provisioningServices.
//
// Contributing Terraform resources (both mutate the single provisioningServices body):
//   - azurerm_iothub_dps
//   - azurerm_iothub_dps_shared_access_policy (properties.authorizationPolicies[*])
//
// Sources:
//   - terraform-provider-azurerm internal/services/iothub/iothub_dps_resource.go
//     (schema L51-204, Create body L229-247, timeouts L44-49, expand L434-528)
//   - terraform-provider-azurerm internal/services/iothub/validate/iot_hub_name.go
//     (IoTHubName L10-19)
//   - go-azure-sdk resource-manager/deviceprovisioningservices/2022-02-05/iotdpsresource:
//     model_provisioningservicedescription.go, model_iotdpspropertiesdescription.go,
//     model_iotdpsskuinfo.go, model_iothubdefinitiondescription.go, model_ipfilterrule.go,
//     constants.go (AllocationPolicy L67-71, PublicNetworkAccess L322-325, IotDpsSku L196-198,
//     IPFilterActionType L111-114, IPFilterTargetType L152-156)
//
// Notes:
//   - name/location/resource_group_name are envelope fields; not emitted as body rules.
//   - sku is a top-level envelope member (sku.name/sku.capacity), NOT under properties.
//   - Folded-into-parent sub-resources (not distinct ARM types):
//       * dps_shared_access_policy maps to properties.authorizationPolicies[*]
//         (array element, also AzureRM-computed) — skipped; present in the Create model,
//         so NOT listed as a ComputedField.
//       * linked_hub maps to properties.iotHubs[*] (array element:
//         connectionString/location/applyAllocationPolicy/allocationWeight 0-1000) — skipped.
//       * ip_filter_rule maps to properties.ipFilterRules[*] (array element:
//         filterName/ipMask/action[Accept|Reject]/target[all|serviceApi|deviceApi]) — skipped.
//   - public_network_access_enabled (bool, default true) maps to the enum
//     properties.publicNetworkAccess (Enabled/Disabled).
type IotHubDps struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*IotHubDps)(nil)

// NewIotHubDps returns knowledge for the Microsoft.Devices/provisioningServices resource.
func NewIotHubDps() *IotHubDps {
	return &IotHubDps{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Devices/provisioningServices",
			ApiVersions:  []string{"2022-02-05"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.enableDataResidency"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[0-9a-zA-Z-]{1,}$`,
					Message:      "DPS name may only contain alphanumeric characters and dashes",
				},
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"S1"},
					Message:       "sku name must be S1",
				},
				{
					PropertyPath:  "properties.allocationPolicy",
					AllowedValues: []string{"Hashed", "GeoLatency", "Static"},
					Message:       "allocation_policy must be one of Hashed, GeoLatency, Static",
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "publicNetworkAccess must be Enabled or Disabled",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "sku.capacity",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(200)),
					Message:      "sku capacity must be between 1 and 200",
				},
			},
			ComputedFields: []string{
				"properties.deviceProvisioningHostName",
				"properties.idScope",
				"properties.serviceOperationsHostName",
				"properties.provisioningState",
				"properties.state",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.allocationPolicy", Value: "Hashed"},
				{PropertyPath: "properties.enableDataResidency", Value: false},
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
			},
			RequiredFields: []string{
				"sku.name",
				"sku.capacity",
			},
		},
	}
}

func init() { azwise.Register(NewIotHubDps()) }
