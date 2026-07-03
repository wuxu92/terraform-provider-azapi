package azwise

import "time"

// StorageAccountBlobService provides resource knowledge for
// Microsoft.Storage/storageAccounts/blobServices.
//
// In AzureRM these settings live in the `blob_properties` block of
// azurerm_storage_account, but ARM models them as a separate sub-service
// resource (blobServices/default), so the knowledge belongs in its own file
// (azwise "Sub-service API separation" rule). All property paths are relative
// to the blobServices body and were verified against the ARM SDK model.
//
// Sources:
//   - terraform-provider-azurerm internal/services/storage/storage_account_resource.go:439-532
//     (blob_properties schema: validators, defaults, MaxItems)
//   - .../storage_account_resource.go:2811-2909 (expandAccountBlobServiceProperties: Terraform→ARM mapping)
//   - .../internal/services/storage/validate/storage_blob_properties_default_service_version.go (defaultServiceVersion enum)
//   - go-azure-sdk resource-manager/storage/2025-08-01/blobservices/model_blobservicepropertiesproperties.go (ARM json paths)
type StorageAccountBlobService struct {
	BaseKnowledge
}

var _ ResourceKnowledge = (*StorageAccountBlobService)(nil)

// NewStorageAccountBlobService returns knowledge for the blobServices sub-resource.
func NewStorageAccountBlobService() *StorageAccountBlobService {
	return &StorageAccountBlobService{
		BaseKnowledge: BaseKnowledge{
			ResourceType: "Microsoft.Storage/storageAccounts/blobServices",
			ApiVersions:  []string{"2025-08-01"},
			// blobServices is a settings sub-resource; every property is updatable
			// in place, so there are no ForceNew body fields. (Replacement of the
			// parent storage account is handled by the envelope's storage_account_id.)
			SoftDelete: false,
			TimeoutsConfig: &Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []StringRule{
				{
					// validate.BlobPropertiesDefaultServiceVersion — a fixed enum of
					// Azure Storage REST API versions (date strings). Mirrors AzureRM exactly.
					PropertyPath: "properties.defaultServiceVersion",
					AllowedValues: []string{
						"2008-10-27", "2009-04-14", "2009-07-17", "2009-09-19",
						"2011-08-28", "2012-02-12", "2013-08-15", "2014-02-14",
						"2015-02-21", "2015-04-05", "2015-07-08", "2015-12-11",
						"2016-05-31", "2017-04-17", "2017-07-29", "2017-11-09",
						"2018-03-28", "2018-11-09", "2019-02-02", "2019-07-07",
						"2019-12-12", "2020-02-10", "2020-04-08", "2020-06-12",
						"2020-10-02", "2020-12-06", "2021-02-12", "2021-04-10",
						"2021-06-08", "2021-08-06", "2021-10-04", "2021-12-02",
						"2022-11-02", "2023-01-03", "2023-11-03",
					},
					Message: "must be a valid Azure Storage service version (e.g. 2023-01-03)",
				},
			},
			IntRules: []IntRule{
				{
					// change_feed_retention_in_days — validation.IntBetween(1, 146000)
					PropertyPath: "properties.changeFeed.retentionInDays",
					MinValue:     ptr(int64(1)),
					MaxValue:     ptr(int64(146000)),
					Message:      "must be between 1 and 146000 days",
				},
				{
					// delete_retention_policy.days — validation.IntBetween(1, 365)
					PropertyPath: "properties.deleteRetentionPolicy.days",
					MinValue:     ptr(int64(1)),
					MaxValue:     ptr(int64(365)),
					Message:      "must be between 1 and 365 days",
				},
				{
					// container_delete_retention_policy.days — validation.IntBetween(1, 365)
					PropertyPath: "properties.containerDeleteRetentionPolicy.days",
					MinValue:     ptr(int64(1)),
					MaxValue:     ptr(int64(365)),
					Message:      "must be between 1 and 365 days",
				},
				{
					// restore_policy.days — validation.IntBetween(1, 365)
					PropertyPath: "properties.restorePolicy.days",
					MinValue:     ptr(int64(1)),
					MaxValue:     ptr(int64(365)),
					Message:      "must be between 1 and 365 days",
				},
			},
			ArrayRules: []ArrayRule{
				{
					// cors_rule — helpers.SchemaStorageAccountCorsRule sets MaxItems: 5.
					// Runtime-only (the static generator does not bake array bounds).
					PropertyPath: "properties.cors.corsRules",
					MaxItems:     5,
					Message:      "a maximum of 5 CORS rules can be configured",
				},
			},
			// AzureRM schema defaults for the blob_properties block.
			DefaultValues: []DefaultValue{
				{PropertyPath: "properties.isVersioningEnabled", Value: false},
				{PropertyPath: "properties.changeFeed.enabled", Value: false},
				{PropertyPath: "properties.lastAccessTimeTrackingPolicy.enable", Value: false},
				{PropertyPath: "properties.deleteRetentionPolicy.allowPermanentDelete", Value: false},
				{PropertyPath: "properties.deleteRetentionPolicy.days", Value: 7},
				{PropertyPath: "properties.containerDeleteRetentionPolicy.days", Value: 7},
			},
			// restore_policy requires delete_retention_policy — AzureRM
			// storage_account_resource.go:522 (RequiredWith). Enabling RestorePolicy
			// without a DeleteRetentionPolicy is rejected (point-in-time restore
			// prerequisite).
			RequiredWith: []RelationalRule{
				{Paths: []string{"properties.restorePolicy", "properties.deleteRetentionPolicy"}},
			},
		},
	}
}
