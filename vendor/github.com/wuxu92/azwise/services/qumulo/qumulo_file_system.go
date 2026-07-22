// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package qumulo

import (
	"time"

	"github.com/wuxu92/azwise"
)

// FileSystem provides resource knowledge for Qumulo.Storage/fileSystems.
//
// Mirrors azurerm_qumulo_file_system. subnet_id maps onto
// properties.delegatedSubnetId, zone onto properties.availabilityZone, and email
// onto properties.userDetails.email; offer/plan/publisher live under
// properties.marketplaceDetails. Every argument is ForceNew.
//
// Sources:
//   - terraform-provider-azurerm internal/services/qumulo/qumulo_file_system_resource.go
//     Arguments() lines 56-131, Create() lines 137-193
//   - Create/Read/Update/Delete timeouts: 90m / 5m / 60m / 30m
//   - go-azure-sdk resource-manager/qumulostorage/2024-06-19/filesystems
//     model_liftrbasestoragefilesystemresourceproperties.go (adminPassword,
//     availabilityZone, delegatedSubnetId, storageSku, marketplaceDetails, userDetails) /
//     model_liftrbasemarketplacedetails.go (offerId, planId, publisherId) /
//     model_liftrbaseuserdetails.go (email)
//   - internal/services/qumulo/validate/file_system_name.go (name regex)
//   - id_filesystem.go: /providers/Qumulo.Storage/fileSystems/{fileSystemName}
//     (note the non-Microsoft "Qumulo.Storage" provider namespace and "fileSystems" casing)
type FileSystem struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*FileSystem)(nil)

// NewFileSystem returns knowledge for the Qumulo fileSystems resource.
func NewFileSystem() *FileSystem {
	return &FileSystem{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Qumulo.Storage/fileSystems",
			ApiVersions:  []string{"2024-06-19", "2025-01-01"},
			ForceNew: []azwise.ForceNewRule{
				// commonschema.Location() is ForceNew; name/resource_group are envelope.
				{PropertyPath: "location"},
				{PropertyPath: "properties.adminPassword"},
				{PropertyPath: "properties.userDetails.email"},
				{PropertyPath: "properties.storageSku"},
				{PropertyPath: "properties.delegatedSubnetId"},
				{PropertyPath: "properties.availabilityZone"},
				{PropertyPath: "properties.marketplaceDetails.offerId"},
				{PropertyPath: "properties.marketplaceDetails.planId"},
				{PropertyPath: "properties.marketplaceDetails.publisherId"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 60 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.adminPassword",
				"properties.availabilityZone",
				"properties.delegatedSubnetId",
				"properties.storageSku",
				"properties.userDetails.email",
				"properties.marketplaceDetails.offerId",
				"properties.marketplaceDetails.planId",
			},
			SensitiveFields: []string{
				"properties.adminPassword",
			},
			DefaultValues: []azwise.DefaultValue{
				// offer_id / plan_id / publisher_id AzureRM schema defaults.
				{PropertyPath: "properties.marketplaceDetails.offerId", Value: "qumulo-saas-mpp"},
				{PropertyPath: "properties.marketplaceDetails.planId", Value: "azure-native-qumulo-v3"},
				{PropertyPath: "properties.marketplaceDetails.publisherId", Value: "qumulo1584033880660"},
			},
			StringRules: []azwise.StringRule{
				// name: validate.FileSystemName (2-15 chars, alphanumeric + hyphen,
				// not starting/ending with a hyphen).
				{
					PropertyPath: "",
					Regex:        `^[a-zA-Z0-9][a-zA-Z0-9-]{0,13}[a-zA-Z0-9]$`,
					Message:      "name must be 2-15 characters of alphanumerics and hyphens, not starting or ending with a hyphen",
				},
				// storage_sku enum. AzureRM uses a StringInSlice pending an SDK enum
				// (azure-rest-api-specs#34017).
				{
					PropertyPath:  "properties.storageSku",
					AllowedValues: []string{"Cold_LRS", "Hot_LRS", "Hot_ZRS"},
					Message:       "storage_sku must be one of Cold_LRS, Hot_LRS, or Hot_ZRS",
				},
			},
			// email uses validation.IsEmailAddress and admin_password uses
			// validate.ValidatePasswordComplexity — semantic checks (email format;
			// >=3 of lower/upper/digit/symbol classes) better ported to customizer
			// validators than expressed declaratively here.
		},
	}
}

func init() { azwise.Register(NewFileSystem()) }
