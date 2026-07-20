package storage

import (
	nativeresource "github.com/Azure/terraform-provider-azapi/internal/native/resource"
	"github.com/Azure/terraform-provider-azapi/internal/native/services"
)

// Storage account blob service hooks model the ARM lifecycle of the singleton
// Microsoft.Storage/storageAccounts/blobServices/default child: it comes into being
// with its parent storage account and has no ARM create or delete operation (DELETE
// returns 405). Marking it a Singleton default makes a Terraform create update the
// always-present default in place, and a Terraform destroy reset it to Azure's
// baseline (static website disabled, no CORS rules, delete retention disabled)
// instead of failing an ARM delete.
func init() {
	nativeresource.RegisterHooks(StorageAccountBlobService.Name, &nativeresource.Hooks{
		Singleton: &nativeresource.SingletonDefault{
			DefaultBody: map[string]interface{}{
				"properties": map[string]interface{}{
					"staticWebsite":         map[string]interface{}{"enabled": false},
					"cors":                  map[string]interface{}{"corsRules": []interface{}{}},
					"deleteRetentionPolicy": map[string]interface{}{"allowPermanentDelete": false, "enabled": false},
				},
			},
		},
		// A restore policy is only valid alongside a delete-retention policy.
		Relational: []services.RelationalConstraint{
			{Kind: services.RequiredWith, Paths: []string{"properties.restore_policy", "properties.delete_retention_policy"}},
		},
	})
}
