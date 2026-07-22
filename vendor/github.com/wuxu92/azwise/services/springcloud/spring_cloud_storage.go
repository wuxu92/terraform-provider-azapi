package springcloud

import (
	"time"

	"github.com/wuxu92/azwise"
)

// SpringCloudStorage provides resource knowledge for
// Microsoft.AppPlatform/Spring/storages.
//
// Mirrors azurerm_spring_cloud_storage.
//
// Sources:
//   - internal/services/springcloud/spring_cloud_storage_resource.go:39-90
//     (Timeouts :39-44 30m/5m/30m/30m; name :52-57 ForceNew; storage_account_name :66-70;
//     storage_account_key :72-76 Required)
//   - go-azure-sdk .../appplatform model_storageaccount.go (AccountKey json "accountKey",
//     AccountName "accountName", StorageType "storageType"), constants.go StorageType
//     (StorageAccount)
type SpringCloudStorage struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*SpringCloudStorage)(nil)

// NewSpringCloudStorage returns knowledge for the Spring Cloud storage resource.
func NewSpringCloudStorage() *SpringCloudStorage {
	return &SpringCloudStorage{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AppPlatform/Spring/storages",
			ApiVersions:  []string{"2024-01-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.accountName",
				"properties.accountKey",
				"properties.storageType",
			},
			StringRules: []azwise.StringRule{
				// storageType discriminator: only StorageAccount is supported today.
				{
					PropertyPath:  "properties.storageType",
					AllowedValues: []string{"StorageAccount"},
				},
			},
			SensitiveFields: []string{"properties.accountKey"},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.storageType", Value: "StorageAccount"},
			},
		},
	}
}

func init() { azwise.Register(NewSpringCloudStorage()) }
