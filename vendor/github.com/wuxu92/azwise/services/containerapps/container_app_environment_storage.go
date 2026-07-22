package containerapps

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ContainerAppEnvironmentStorage provides resource knowledge for
// Microsoft.App/managedEnvironments/storages.
//
// Mirrors azurerm_container_app_environment_storage.
//
// Sources:
//   - terraform-provider-azurerm internal/services/containerapps/container_app_environment_storage_resource.go
//     schema (Arguments 47-110) + Create (116-180): name/account/share/access_mode/nfs ForceNew,
//     access_mode enum, access_key sensitive, azureFile vs nfsAzureFile ConflictsWith/RequiredWith;
//     timeouts 30m/5m/30m/30m.
//   - go-azure-sdk resource-manager/containerapps/2025-07-01/managedenvironmentsstorages
//     AzureFileProperties (accessMode/accountKey/accountName/shareName), NfsAzureFileProperties
//     (accessMode/server/shareName), constants.go AccessMode enum.
//
// NOTE: share_name and access_mode are Required in Terraform but live under whichever of
// properties.azureFile / properties.nfsAzureFile is present, so they are not declared as
// universal RequiredFields (that would reject the variant that omits the other sub-object).
type ContainerAppEnvironmentStorage struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ContainerAppEnvironmentStorage)(nil)

func NewContainerAppEnvironmentStorage() *ContainerAppEnvironmentStorage {
	return &ContainerAppEnvironmentStorage{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.App/managedEnvironments/storages",
			ApiVersions:  []string{"2025-07-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.azureFile.accountName"},
				{PropertyPath: "properties.azureFile.shareName"},
				{PropertyPath: "properties.azureFile.accessMode"},
				{PropertyPath: "properties.nfsAzureFile.server"},
				{PropertyPath: "properties.nfsAzureFile.shareName"},
				{PropertyPath: "properties.nfsAzureFile.accessMode"},
			},
			StringRules: []azwise.StringRule{
				// access_mode -> AccessMode enum (applies to both storage variants).
				{
					PropertyPath:  "properties.azureFile.accessMode",
					AllowedValues: []string{"ReadOnly", "ReadWrite"},
					Message:       "access mode must be ReadOnly or ReadWrite",
				},
				{
					PropertyPath:  "properties.nfsAzureFile.accessMode",
					AllowedValues: []string{"ReadOnly", "ReadWrite"},
					Message:       "access mode must be ReadOnly or ReadWrite",
				},
			},
			// access_key is Sensitive.
			SensitiveFields: []string{
				"properties.azureFile.accountKey",
			},
			// account_name conflicts with nfs_server_url; account_name requires access_key.
			ConflictsWith: []azwise.RelationalRule{
				{Paths: []string{"properties.azureFile.accountName", "properties.nfsAzureFile.server"}, Message: "an Azure File storage account cannot be combined with an NFS server"},
			},
			RequiredWith: []azwise.RelationalRule{
				{Paths: []string{"properties.azureFile.accountName", "properties.azureFile.accountKey"}},
			},
		},
	}
}

func init() { azwise.Register(NewContainerAppEnvironmentStorage()) }
