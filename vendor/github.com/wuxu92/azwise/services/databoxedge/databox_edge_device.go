package databoxedge

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DataBoxEdgeDevice provides resource knowledge for Microsoft.DataBoxEdge/dataBoxEdgeDevices.
//
// Contributing Terraform resource:
//   - azurerm_databox_edge_device
//
// Sources:
//   - terraform-provider-azurerm internal/services/databoxedge/databox_edge_device_resource.go
//     (schema Arguments L73-95, computed Attributes L97-165, Create timeout L169 + body L193-197,
//     expandDeviceSku L311-325 splitting sku_name "<Name>-<Tier>" into sku.name/sku.tier)
//   - internal/services/databoxedge/validate/databox_edge_name.go (name regex, 3-24 chars)
//   - internal/services/databoxedge/validate/databox_edge_device_sku_name.go
//     (composite "<SkuName>-<SkuTier>" validator using PossibleValuesForSkuName/SkuTier)
//   - go-azure-sdk resource-manager/databoxedge/2022-03-01/devices:
//     model_databoxedgedevice.go, model_sku.go, model_databoxedgedeviceproperties.go,
//     constants.go (SkuName L834-862, SkuTier L956).
//
// Notes:
//   - sku_name is a composite TF field ("<SkuName>-<SkuTier>", e.g. "EdgeP_Base-Standard")
//     that AzureRM splits into two ARM properties: sku.name (SkuName enum) and sku.tier
//     (SkuTier enum). The composite TF validator has no single ARM body field, so it is not
//     emitted; instead the two ARM enum StringRules below cover sku.name and sku.tier directly.
//   - The entire device_properties block is Computed in AzureRM — these are read-only device
//     telemetry surfaced in the GET response and never set on create; listed in ComputedFields.
//   - name and the sku (name+tier) are ForceNew.
type DataBoxEdgeDevice struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DataBoxEdgeDevice)(nil)

// NewDataBoxEdgeDevice returns knowledge for the Microsoft.DataBoxEdge/dataBoxEdgeDevices resource.
func NewDataBoxEdgeDevice() *DataBoxEdgeDevice {
	return &DataBoxEdgeDevice{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DataBoxEdge/dataBoxEdgeDevices",
			ApiVersions:  []string{"2022-03-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "sku.name"},
				{PropertyPath: "sku.tier"},
			},
			StringRules: []azwise.StringRule{
				// name (resource name; empty PropertyPath). 3-24 chars.
				{
					PropertyPath: "",
					Regex:        `^[\da-zA-Z][-\da-zA-Z]{1,22}[\da-zA-Z]$`,
					MinLength:    3,
					MaxLength:    24,
					Message:      "name must be 3-24 characters, begin and end with an alphanumeric character, and contain only alphanumeric characters and hyphens",
				},
				{
					PropertyPath: "sku.name",
					AllowedValues: []string{
						"Gateway", "Edge", "TEA_1Node", "TEA_1Node_UPS", "TEA_1Node_Heater",
						"TEA_1Node_UPS_Heater", "TEA_4Node_Heater", "TEA_4Node_UPS_Heater", "TMA",
						"TDC", "TCA_Small", "GPU", "TCA_Large", "EdgeP_Base", "EdgeP_High",
						"EdgePR_Base", "EdgePR_Base_UPS", "EdgeMR_Mini", "RCA_Small", "RCA_Large",
						"RDC", "Management", "EP2_64_1VPU_W", "EP2_128_1T4_Mx1_W", "EP2_256_2T4_W",
						"EP2_64_Mx1_W", "EP2_128_GPU1_Mx1_W", "EP2_256_GPU2_Mx1",
					},
					Message: "must be a valid Data Box Edge SKU name",
				},
				{
					PropertyPath:  "sku.tier",
					AllowedValues: []string{"Standard"},
					Message:       "must be Standard",
				},
			},
			// Read-only device telemetry (device_properties is Computed in AzureRM).
			ComputedFields: []string{
				"properties.configuredRoleTypes",
				"properties.culture",
				"properties.dataBoxEdgeDeviceStatus",
				"properties.deviceHcsVersion",
				"properties.deviceLocalCapacity",
				"properties.deviceModel",
				"properties.deviceSoftwareVersion",
				"properties.deviceType",
				"properties.nodeCount",
				"properties.serialNumber",
				"properties.timeZone",
				"properties.edgeProfile",
				"properties.resourceMoveDetails",
			},
			RequiredFields: []string{
				"sku.name",
				"sku.tier",
			},
		},
	}
}

// Self-registers into the azwise registry.
func init() { azwise.Register(NewDataBoxEdgeDevice()) }
