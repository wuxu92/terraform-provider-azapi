package mongocluster

import (
	"strings"
	"time"

	"github.com/wuxu92/azwise"
)

// MongoCluster provides resource knowledge for
// Microsoft.DocumentDB/mongoClusters (azurerm_mongo_cluster).
//
// It overrides CheckForceNew to add AzureRM's value-conditional replacement that
// the static ForceNew list cannot express: the Data API mode can only be turned
// on after the cluster exists, so an Enabled->Disabled change forces replacement.
//
// Sources:
//   - AzureRM internal/services/mongocluster/mongo_cluster_resource.go
//     :89-97   (name StringMatch regex, 3-40 chars)
//     :103-291 (schema: create_mode, customer_managed_key, preview_features,
//     restore, shard_count, source_location, source_server_id, compute_tier,
//     high_availability_mode, public_network_access, storage_size_in_gb,
//     storage_type, version enums/ranges/defaults/ForceNew)
//     :327-464 (Create: ARM body mapping administrator/authConfig/encryption/
//     restoreParameters/sharding/replicaParameters/compute/highAvailability/
//     storage/dataApi/serverVersion)
//     :713-808 (CustomizeDiff: create_mode change ForceNew, data_api_mode_enabled
//     Enabled->Disabled ForceNew, identity add/remove ForceNew, conditional
//     required fields per create_mode)
//     :329,468,591,695 (timeouts: create 60m, update 60m, read 5m, delete 60m)
//   - go-azure-sdk resource-manager/mongocluster/2025-09-01/mongoclusters:
//     model_mongoclusterproperties.go (ARM body paths),
//     model_storageproperties.go / model_computeproperties.go /
//     model_shardingproperties.go / model_administratorproperties.go /
//     model_authconfigproperties.go / model_highavailabilityproperties.go /
//     model_dataapiproperties.go / model_encryptionproperties.go,
//     constants.go (enum PossibleValuesFor* value sets)
//
// Not encoded (deliberate):
//   - authentication_methods maps to properties.authConfig.allowedModes, an array
//     of AuthenticationMode enums. Its O+C default (["NativeAuth"]) and the
//     per-element enum are array-element paths azwise cannot lower, so no rule and
//     no DefaultValue is emitted for it.
//   - customer_managed_key (properties.encryption) and restore
//     (properties.restoreParameters) are MaxItems:1 lists representing single ARM
//     objects, not arrays, so no ArrayRule is emitted; they are still ForceNew.
//   - storage_type default "PremiumSSD" and the conditional dataApi default are
//     nested under objects that only exist when their siblings are set, so
//     injecting them as DefaultValues could fabricate partial objects; left out.
//   - The create_mode=Default conditional requireds (administrator_username,
//     compute_tier, storage_size_in_gb, high_availability_mode, shard_count,
//     version) are mode-conditional and would corrupt GeoReplica /
//     PointInTimeRestore bodies if listed unconditionally, so RequiredFields is
//     empty (only name/location/resource_group_name are schema-Required and those
//     are envelope-owned).
//   - identity add/remove ForceNew (resource :800-804) is a presence-toggle over
//     the identity envelope, not a body-path value change, so it is left as a
//     residual hook rather than a declarative rule.
type MongoCluster struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*MongoCluster)(nil)

// CheckForceNew extends BaseKnowledge with AzureRM's value-conditional
// replacement (mongo_cluster_resource.go:789-793): the Data API mode can only be
// enabled after creation, so moving properties.dataApi.mode from Enabled to
// Disabled forces replacement.
func (s *MongoCluster) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
	if s.BaseKnowledge.CheckForceNew(oldBody, newBody) {
		return true
	}

	oldMode, _ := azwise.ExtractNestedValue(oldBody, "properties.dataApi.mode").(string)
	newMode, _ := azwise.ExtractNestedValue(newBody, "properties.dataApi.mode").(string)
	if strings.EqualFold(oldMode, "Enabled") && strings.EqualFold(newMode, "Disabled") {
		return true
	}

	return false
}

// NewMongoCluster returns knowledge for the mongoClusters resource.
func NewMongoCluster() *MongoCluster {
	return &MongoCluster{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.DocumentDB/mongoClusters",
			ApiVersions:  []string{"2025-09-01"},
			SoftDelete:   false,
			// location (commonschema.Location) replaces the cluster on change; the
			// remaining fields are unconditionally ForceNew in the AzureRM schema,
			// plus create_mode which the CustomizeDiff forces on any change.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "location"},
				{PropertyPath: "properties.createMode"},
				{PropertyPath: "properties.administrator.userName"},
				{PropertyPath: "properties.encryption"},
				{PropertyPath: "properties.previewFeatures"},
				{PropertyPath: "properties.restoreParameters"},
				{PropertyPath: "properties.sharding.shardCount"},
				{PropertyPath: "properties.replicaParameters.sourceLocation"},
				{PropertyPath: "properties.replicaParameters.sourceResourceId"},
				{PropertyPath: "properties.storage.type"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name: 3-40 chars, lowercase letters, digits, hyphens,
					// starting and ending with a lowercase letter or digit.
					Regex:     `^[a-z\d]([-a-z\d]{1,38}[a-z\d])$`,
					MinLength: 3,
					MaxLength: 40,
					Message:   "`name` must be between 3 and 40 characters and contain only lowercase letters, numbers and hyphens, starting and ending with a lowercase letter or number",
				},
				{
					PropertyPath:  "properties.createMode",
					AllowedValues: []string{"Default", "GeoReplica", "PointInTimeRestore", "Replica"},
					Message:       "must be one of Default, GeoReplica, PointInTimeRestore or Replica",
				},
				{
					PropertyPath:  "properties.compute.tier",
					AllowedValues: []string{"Free", "M10", "M20", "M25", "M30", "M40", "M50", "M60", "M80", "M200"},
					Message:       "must be a supported compute tier",
				},
				{
					PropertyPath:  "properties.highAvailability.targetMode",
					AllowedValues: []string{"Disabled", "SameZone", "ZoneRedundantPreferred"},
					Message:       "must be one of Disabled, SameZone or ZoneRedundantPreferred",
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "must be Disabled or Enabled",
				},
				{
					PropertyPath:  "properties.storage.type",
					AllowedValues: []string{"PremiumSSD", "PremiumSSDv2"},
					Message:       "must be PremiumSSD or PremiumSSDv2",
				},
				{
					PropertyPath:  "properties.serverVersion",
					AllowedValues: []string{"5.0", "6.0", "7.0", "8.0"},
					Message:       "must be a supported MongoDB server version",
				},
			},
			IntRules: []azwise.IntRule{
				{
					PropertyPath: "properties.sharding.shardCount",
					MinValue:     azwise.Ptr(int64(1)),
					Message:      "must be at least 1",
				},
				{
					PropertyPath: "properties.storage.sizeGb",
					MinValue:     azwise.Ptr(int64(32)),
					MaxValue:     azwise.Ptr(int64(32768)),
					Message:      "must be between 32 and 32768 GB",
				},
			},
			// administrator_password is Sensitive and maps to a real body path.
			SensitiveFields: []string{
				"properties.administrator.password",
			},
			// Response-only properties absent from the create/update payload.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.clusterStatus",
				"properties.connectionString",
				"properties.infrastructureVersion",
				"properties.privateEndpointConnections",
				"properties.replica",
			},
			// Only the always-set top-level defaults are safe to inject; nested
			// object defaults (storage.type, dataApi.mode) are conditional.
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.createMode", Value: "Default"},
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
			},
			RequiredFields: []string{},
		},
	}
}

func init() { azwise.Register(NewMongoCluster()) }
