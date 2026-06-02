package azwise

import (
	"strings"
	"time"
)

// StorageAccount provides resource knowledge for Microsoft.Storage/storageAccounts.
// It overrides CheckForceNew to implement conditional SKU zone-migration logic.
type StorageAccount struct {
	BaseKnowledge
}

var _ ResourceKnowledge = (*StorageAccount)(nil)

// CheckForceNew overrides BaseKnowledge — only cross-zone SKU migration triggers replacement.
// Changing between zonal (ZRS/GZRS/RAGZRS) and non-zonal (LRS/GRS/RAGRS) requires recreation.
func (s *StorageAccount) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
	oldSku := strings.ToUpper(extractStringValue(oldBody, "sku.name"))
	newSku := strings.ToUpper(extractStringValue(newBody, "sku.name"))
	if oldSku == "" || newSku == "" || oldSku == newSku {
		return false
	}
	zonal := map[string]bool{
		"STANDARD_ZRS":     true,
		"STANDARD_GZRS":   true,
		"STANDARD_RAGZRS": true,
	}
	nonZonal := map[string]bool{
		"STANDARD_LRS":   true,
		"STANDARD_GRS":   true,
		"STANDARD_RAGRS": true,
	}
	return (zonal[oldSku] && nonZonal[newSku]) || (nonZonal[oldSku] && zonal[newSku])
}

// NewStorageAccount returns a StorageAccount knowledge instance.
func NewStorageAccount() *StorageAccount {
	return &StorageAccount{
		BaseKnowledge: BaseKnowledge{
			ResourceType: "Microsoft.Storage/storageAccounts",
			TimeoutsConfig: &Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 60 * time.Minute,
			},
			StringRules: []StringRule{
				{
					// Resource name: lowercase alphanumeric only, 3-24 chars
					Regex:     `^[a-z0-9]{3,24}$`,
					MinLength: 3,
					MaxLength: 24,
					Message:   "must be lowercase alphanumeric, 3-24 characters",
				},
				{
					PropertyPath: "sku.name",
					AllowedValues: []string{
						"Standard_LRS", "Standard_GRS", "Standard_RAGRS",
						"Standard_ZRS", "Standard_GZRS", "Standard_RAGZRS",
						"Premium_LRS", "Premium_ZRS",
					},
					Message: "must be a valid storage SKU (e.g. Standard_LRS, Premium_ZRS)",
				},
				{
					PropertyPath: "kind",
					AllowedValues: []string{
						"BlobStorage", "BlockBlobStorage", "FileStorage",
						"Storage", "StorageV2",
					},
					Message: "must be a valid storage kind (e.g. StorageV2, BlobStorage)",
				},
				{
					PropertyPath:  "properties.accessTier",
					AllowedValues: []string{"Hot", "Cool", "Premium"},
					Message:       "must be Hot, Cool, or Premium",
				},
			},
			SensitiveFields: []string{
				"properties.primaryAccessKey",
				"properties.secondaryAccessKey",
			},
		},
	}
}
