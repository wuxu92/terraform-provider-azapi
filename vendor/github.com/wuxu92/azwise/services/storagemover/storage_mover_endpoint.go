package storagemover

import (
	"time"

	"github.com/wuxu92/azwise"
)

// StorageMoverEndpoint provides resource knowledge for
// Microsoft.StorageMover/storageMovers/endpoints.
//
// MERGES two AzureRM TF resources that both map to the SAME ARM type
// (storageMovers/endpoints), discriminated by properties.endpointType:
//   - azurerm_storage_mover_source_endpoint → endpointType "NfsMount"
//     (properties.host, properties.export, properties.nfsVersion)
//   - azurerm_storage_mover_target_endpoint → endpointType "AzureStorageBlobContainer"
//     (properties.blobContainerName, properties.storageAccountResourceId)
//
// Only knowledge universal to every body is unioned at the top level (name
// validation, timeouts, the endpointType discriminator enum, provisioningState
// computed field). The nfsVersion enum is safe to include because it fires only
// when that field is present (NfsMount bodies). Kind-specific ForceNew
// (host/export/nfs_version for source; storage_account_id/storage_container_name for
// target), kind-specific RequiredFields, and the source-only nfs_version default are
// NOT unioned, since they would corrupt validation for the other kind.
//
// Sources:
//   - internal/services/storagemover/storage_mover_source_endpoint_resource.go
//     (Arguments 56-107: name ForceNew StringMatch ^[0-9a-zA-Z][-_0-9a-zA-Z]{0,63}$;
//     host Required ForceNew; export Optional ForceNew; nfs_version Optional ForceNew
//     Default "NFSauto" StringInSlice(NFSauto,NFSv4,NFSv3); description Optional;
//     create 141-147 → NfsMountEndpointProperties{Export, Host, NfsVersion}).
//   - internal/services/storagemover/storage_mover_target_endpoint_resource.go
//     (Arguments 57-96: storage_account_id Required ForceNew; storage_container_name
//     Required ForceNew; create 130-135 → AzureStorageBlobContainerEndpointProperties{
//     BlobContainerName, StorageAccountResourceId}).
//   - go-azure-sdk resource-manager/storagemover/2025-07-01/endpoints:
//     model_endpoint.go (properties discriminated union), model_nfsmountendpointproperties.go
//     (export/host/nfsVersion/endpointType), model_azurestorageblobcontainerendpointproperties.go
//     (blobContainerName/storageAccountResourceId/endpointType),
//     constants.go (EndpointType AzureMultiCloudConnector/AzureStorageBlobContainer/
//     AzureStorageNfsFileShare/AzureStorageSmbFileShare/NfsMount; NfsVersion NFSauto/NFSv4/NFSv3),
//     id_endpoint.go (type segment casing "storageMovers/endpoints").
type StorageMoverEndpoint struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*StorageMoverEndpoint)(nil)

// NewStorageMoverEndpoint returns knowledge for the storageMovers/endpoints resource.
func NewStorageMoverEndpoint() *StorageMoverEndpoint {
	return &StorageMoverEndpoint{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.StorageMover/storageMovers/endpoints",
			ApiVersions:  []string{"2025-07-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// Universal ForceNew only. name is ForceNew across both kinds.
			// (host/export/nfs_version and storage_account_id/storage_container_name are
			// kind-specific and omitted.)
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			StringRules: []azwise.StringRule{
				{
					// name: validation.StringMatch.
					Regex:     `^[0-9a-zA-Z][-_0-9a-zA-Z]{0,63}$`,
					MinLength: 1,
					MaxLength: 64,
					Message:   "must be 1-64 characters, begin with a letter or number, and contain only letters, numbers, dashes and underscores",
				},
				{
					// Discriminator, present in every body.
					PropertyPath: "properties.endpointType",
					AllowedValues: []string{
						"AzureMultiCloudConnector",
						"AzureStorageBlobContainer",
						"AzureStorageNfsFileShare",
						"AzureStorageSmbFileShare",
						"NfsMount",
					},
				},
				{
					// nfs_version — fires only when the NfsMount properties.nfsVersion is set.
					PropertyPath:  "properties.nfsVersion",
					AllowedValues: []string{"NFSauto", "NFSv4", "NFSv3"},
				},
			},
			// endpointType is required (non-omitempty) in every body.
			RequiredFields: []string{
				"properties.endpointType",
			},
			// Server-populated, read-only ARM property.
			ComputedFields: []string{
				"properties.provisioningState",
			},
			// NOTE: kind-specific RequiredFields are intentionally omitted: source requires
			// properties.host, target requires properties.blobContainerName +
			// properties.storageAccountResourceId; unioning them would wrongly force every
			// kind to carry the others' fields.
			// NOTE: nfs_version has AzureRM Default "NFSauto" but only for the NfsMount kind;
			// it is not added to DefaultValues because injecting it into an
			// AzureStorageBlobContainer body would corrupt the target endpoint.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewStorageMoverEndpoint()) }
