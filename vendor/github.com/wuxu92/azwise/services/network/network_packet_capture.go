package network

import (
	"time"

	"github.com/wuxu92/azwise"
)

// PacketCapture provides resource knowledge for
// Microsoft.Network/networkWatchers/packetCaptures.
//
// This single ARM resource type is contributed by three AzureRM resources, all
// backed by the go-azure-sdk packetcaptures client:
//   - azurerm_network_packet_capture (deprecated; target_resource_id)
//   - azurerm_virtual_machine_packet_capture (virtual_machine_id, targetType AzureVM)
//   - azurerm_virtual_machine_scale_set_packet_capture (virtual_machine_scale_set_id,
//     targetType AzureVMSS, plus machine_scope)
//
// A packet capture is immutable: every field is ForceNew and only Create / Read /
// Delete exist (no Update). Only knowledge universal to all three contributors is
// encoded here; value constraints on optional sub-objects (targetType) fire only
// when that sub-object is present and are therefore safe to union.
//
// Sources:
//   - terraform-provider-azurerm internal/services/network/network_packet_capture_resource.go
//     (schema lines 25-152, expand lines 182-191)
//   - .../virtual_machine_packet_capture_resource.go (schema lines 25-151, expand 186-196)
//   - .../virtual_machine_scale_set_packet_capture_resource.go (schema lines 27-186, expand 220-234)
//   - go-azure-sdk resource-manager/network/2025-01-01/packetcaptures:
//     id_packetcapture.go (ARM type segments "networkWatchers"/"packetCaptures"),
//     model_packetcaptureparameters.go / model_packetcapturestoragelocation.go
//     (target, timeLimitInSeconds, bytesToCapturePerPacket, totalBytesPerSession,
//     storageLocation.filePath/storageId json tags),
//     constants.go (PacketCaptureTargetType: AzureVM, AzureVMSS)
//
// Not encoded (deliberate):
//   - filter is an array (properties.filters); the protocol enum (Any/TCP/UDP)
//     lives under filters[*].protocol and cannot be resolved through an array
//     element, so it is skipped.
//   - machine_scope (properties.scope) is VMSS-only and its exclude/include lists
//     are ConflictsWith at the array-element level; not encoded on the shared type.
//   - storage_location.storage_path is Computed read-only (storageLocation.storagePath).
type PacketCapture struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*PacketCapture)(nil)

// NewPacketCapture returns knowledge for the networkWatchers/packetCaptures resource.
func NewPacketCapture() *PacketCapture {
	return &PacketCapture{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/networkWatchers/packetCaptures",
			ApiVersions:  []string{"2025-01-01"},
			// Every settable field is ForceNew across all three contributors.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.target"},
				{PropertyPath: "properties.storageLocation"},
				{PropertyPath: "properties.bytesToCapturePerPacket"},
				{PropertyPath: "properties.totalBytesPerSession"},
				{PropertyPath: "properties.timeLimitInSeconds"},
				{PropertyPath: "properties.filters"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Set only by the VM / VMSS contributors; fires only when present.
					PropertyPath:  "properties.targetType",
					AllowedValues: []string{"AzureVM", "AzureVMSS"},
					Message:       "must be AzureVM or AzureVMSS",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.timeLimitInSeconds",
					MinValue:     azwise.Ptr(int64(1)),
					MaxValue:     azwise.Ptr(int64(18000)),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.timeLimitInSeconds", Value: float64(18000)},
				{PropertyPath: "properties.bytesToCapturePerPacket", Value: float64(0)},
				{PropertyPath: "properties.totalBytesPerSession", Value: float64(1073741824)},
			},
			RequiredFields: []string{
				"properties.target",
				"properties.storageLocation",
			},
			// storage_location requires at least one of file_path / storage_account_id
			// (both under the single storageLocation object).
			AtLeastOneOf: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.storageLocation.filePath", "properties.storageLocation.storageId"},
					Message: "at least one of file_path or storage_account_id must be set",
				},
			},
		},
	}
}

func init() { azwise.Register(NewPacketCapture()) }
