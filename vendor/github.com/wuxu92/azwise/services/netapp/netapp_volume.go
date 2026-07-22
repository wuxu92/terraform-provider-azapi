package netapp

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetAppVolume provides resource knowledge for
// Microsoft.NetApp/netAppAccounts/capacityPools/volumes.
//
// Mirrors azurerm_netapp_volume. (The volume_group_* resources create volumes too,
// but through the distinct Microsoft.NetApp/netAppAccounts/volumeGroups ARM type —
// see netapp_volume_group.go.)
//
// Sources:
//   - internal/services/netapp/netapp_volume_resource.go
//     (schema 58-456: name/zone/volume_path/subnet_id/create_from_snapshot_resource_id/
//     accept_grow.../kerberos_enabled/smb_continuous_availability_enabled/
//     smb3_protocol_encryption_enabled/security_style/data_protection_replication/
//     azure_vmware_data_store_enabled/encryption_key_source/key_vault_private_endpoint_id
//     all ForceNew; service_level/network_features/security_style enums;
//     storage_quota_in_gb IntBetween(50,1048576) GiB; throughput_in_mibps FloatAtLeast(1.0);
//     cool_access block coolness_period_in_days IntBetween(2,183); timeouts C/U/D 60m R 5m;
//     create 619-891 → VolumeProperties mapping; CustomizeDiff 457-578).
//   - internal/services/netapp/validate/volume_name.go (regex ^[a-zA-Z][-_\da-zA-Z]{0,63}$).
//   - go-azure-sdk resource-manager/netapp/2026-01-01/volumes:
//     model_volumeproperties.go (creationToken/subnetId/usageThreshold required; many read-only),
//     constants.go (ServiceLevel, NetworkFeatures Basic/Standard/Basic_Standard/Standard_Basic,
//     SecurityStyle ntfs/unix, EncryptionKeySource, CoolAccessRetrievalPolicy/TieringPolicy,
//     AcceptGrowCapacityPoolForShortTermCloneSplit, SmbNonBrowsable/SmbAccessBasedEnumeration/
//     AvsDataStore Disabled/Enabled), id_volume.go (type segment casing "volumes").
type NetAppVolume struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetAppVolume)(nil)

// NewNetAppVolume returns knowledge for the volumes resource.
func NewNetAppVolume() *NetAppVolume {
	return &NetAppVolume{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.NetApp/netAppAccounts/capacityPools/volumes",
			ApiVersions:  []string{"2026-01-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "zones"},
				{PropertyPath: "properties.creationToken"},
				{PropertyPath: "properties.subnetId"},
				{PropertyPath: "properties.snapshotId"},
				{PropertyPath: "properties.acceptGrowCapacityPoolForShortTermCloneSplit"},
				{PropertyPath: "properties.kerberosEnabled"},
				{PropertyPath: "properties.smbContinuouslyAvailable"},
				{PropertyPath: "properties.smbEncryption"},
				{PropertyPath: "properties.securityStyle"},
				{PropertyPath: "properties.dataProtection.replication"},
				{PropertyPath: "properties.avsDataStore"},
				{PropertyPath: "properties.encryptionKeySource"},
				{PropertyPath: "properties.keyVaultPrivateEndpointResourceId"},
			},
			// creationToken, subnetId and usageThreshold are required (no omitempty in model).
			RequiredFields: []string{
				"properties.creationToken",
				"properties.subnetId",
				"properties.usageThreshold",
			},
			StringRules: []azwise.StringRule{
				{
					// AzureRM VolumeName: 1-64 chars, start with a letter.
					Regex:     `^[a-zA-Z][-_\da-zA-Z]{0,63}$`,
					MinLength: 1,
					MaxLength: 64,
					Message:   "must be 1-64 characters, start with a letter, and contain only letters, numbers, underscores and hyphens",
				},
				{
					PropertyPath:  "properties.serviceLevel",
					AllowedValues: []string{"Flexible", "Premium", "Standard", "StandardZRS", "Ultra"},
				},
				{
					PropertyPath:  "properties.networkFeatures",
					AllowedValues: []string{"Basic", "Basic_Standard", "Standard", "Standard_Basic"},
				},
				{
					PropertyPath:  "properties.securityStyle",
					AllowedValues: []string{"ntfs", "unix"},
				},
				{
					PropertyPath:  "properties.encryptionKeySource",
					AllowedValues: []string{"Microsoft.KeyVault", "Microsoft.NetApp"},
				},
				{
					PropertyPath:  "properties.acceptGrowCapacityPoolForShortTermCloneSplit",
					AllowedValues: []string{"Accepted", "Declined"},
				},
				{
					PropertyPath:  "properties.coolAccessRetrievalPolicy",
					AllowedValues: []string{"Default", "Never", "OnRead"},
				},
				{
					PropertyPath:  "properties.coolAccessTieringPolicy",
					AllowedValues: []string{"Auto", "SnapshotOnly"},
				},
			},
			IntRules: []azwise.IntRule{
				{
					// cool_access.coolness_period_in_days IntBetween(2,183).
					PropertyPath: "properties.coolnessPeriod",
					MinValue:     azwise.Ptr(int64(2)),
					MaxValue:     azwise.Ptr(int64(183)),
				},
			},
			FloatRules: []azwise.FloatRule{
				{
					// throughput_in_mibps FloatAtLeast(1.0).
					PropertyPath: "properties.throughputMibps",
					MinValue:     azwise.Ptr(float64(1.0)),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.snapshotDirectoryVisible", Value: true},
				{PropertyPath: "properties.smbNonBrowsable", Value: "Disabled"},
				{PropertyPath: "properties.smbAccessBasedEnumeration", Value: "Disabled"},
				{PropertyPath: "properties.avsDataStore", Value: "Disabled"},
				{PropertyPath: "properties.isLargeVolume", Value: false},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.mountTargets",
				"properties.fileSystemId",
				"properties.baremetalTenantId",
				"properties.networkSiblingSetId",
				"properties.provisionedAvailabilityZone",
				"properties.actualThroughputMibps",
				"properties.cloneProgress",
				"properties.encrypted",
				"properties.t2Network",
				"properties.storageToNetworkProximity",
				"properties.effectiveNetworkFeatures",
				"properties.inheritedSizeInBytes",
				"properties.capacityPoolResourceId",
				"properties.originatingResourceId",
			},
			// NOTE: storage_quota_in_gb (IntBetween 50-1048576 GiB) is converted to
			// properties.usageThreshold in BYTES (GiB * 1024^3) before send, so a declarative
			// GiB IntRule would be wrong-unit and is intentionally omitted.
			// NOTE: protocols maps to properties.protocolTypes[*] (array); export_policy_rule
			// maps to properties.exportPolicy.rules[*] (array of objects) — element-level
			// validators (rule_index 1-5, allowed_clients CIDR, protocol enum) are
			// array-element paths and cannot be expressed declaratively here.
			// NOTE: CustomizeDiff enforces value-conditional invariants that are not declarative:
			// service_level/pool_name must change together (else ForceNew); large_volume_enabled
			// vs storage_quota_in_gb size bands; short-term-clone incompatibility with large
			// volumes and cool_access; NFSv3->NFSv4.1 conversion restrictions; cross-zone
			// replication requires a zone; ntfs/unix security_style vs protocol constraints.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewNetAppVolume()) }
