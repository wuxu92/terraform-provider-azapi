// Copyright (c) HashiCorp, Inc.

package healthcare

import (
	"time"

	"github.com/wuxu92/azwise"
)

// DicomService provides resource knowledge for
// Microsoft.HealthcareApis/workspaces/dicomservices.
//
// Contributing TF resource:
//   - azurerm_healthcare_dicom_service
//
// Sources:
//   - AzureRM internal/services/healthcare/healthcare_dicom_service_resource.go
//     (schema :32-215, Create :217-292, expand storage/cors :537-609)
//   - AzureRM internal/services/healthcare/validate/dicom_name.go (name rule)
//   - go-azure-sdk .../healthcareapis/2024-03-31/dicomservices:
//     id_dicomservice.go (segment casing: workspaces/dicomServices),
//     model_dicomservice.go, model_dicomserviceproperties.go,
//     model_corsconfiguration.go, model_storageconfiguration.go,
//     model_encryption.go, model_encryptioncustomermanagedkeyencryption.go, constants.go
type DicomService struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*DicomService)(nil)

func NewDicomService() *DicomService {
	return &DicomService{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.HealthcareApis/workspaces/dicomservices",
			ApiVersions:  []string{"2024-03-31"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
				{PropertyPath: "properties.enableDataPartitions"},                    // data_partitions_enabled
				{PropertyPath: "properties.storageConfiguration"},                    // storage block
				{PropertyPath: "properties.storageConfiguration.fileSystemName"},     // storage.file_system_name
				{PropertyPath: "properties.storageConfiguration.storageResourceId"},  // storage.storage_account_id
				// NOTE: workspace_id is the parent resource reference (not a body path); omitted.
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 90 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). validate.DicomServiceName:
					// 3-24 chars, starts/ends alphanumeric, dashes allowed.
					Regex:   `^[0-9a-zA-Z][-0-9a-zA-Z]{1,22}[0-9a-zA-Z]$`,
					Message: "must be 3-24 characters, start and end with a letter or number, and contain only letters, numbers, and dashes",
				},
				{
					// encryption_key_url -> validation.IsURLWithHTTPS.
					PropertyPath: "properties.encryption.customerManagedKeyEncryption.keyEncryptionKeyUrl",
					Regex:        `^https://`,
					Message:      "encryption_key_url must be a valid HTTPS URL",
				},
				// NOTE: cors.allowed_origins / allowed_headers / allowed_methods elements
				// use validation.StringIsNotEmpty; these are array-element paths
				// (properties.corsConfiguration.origins[*], .headers[*], .methods[*]) and
				// are not representable as declarative StringRules — intentionally omitted.
			},
			IntRules: []azwise.IntRule{
				// cors.max_age_in_seconds -> validation.IntBetween(0, 99998).
				{PropertyPath: "properties.corsConfiguration.maxAge", MinValue: azwise.Ptr(int64(0)), MaxValue: azwise.Ptr(int64(99998)), Message: "max_age_in_seconds must be between 0 and 99998"},
			},
			DefaultValues: []azwise.DefaultValue{
				// public_network_access_enabled default true -> PublicNetworkAccessEnabled.
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				// data_partitions_enabled default false.
				{PropertyPath: "properties.enableDataPartitions", Value: false},
				// cors.allow_credentials default false.
				{PropertyPath: "properties.corsConfiguration.allowCredentials", Value: false},
			},
			// Azure-populated, read-only fields absent from the create body.
			ComputedFields: []string{
				"properties.serviceUrl",
				"properties.authenticationConfiguration",
				"properties.privateEndpointConnections",
				"properties.provisioningState",
				"properties.eventState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDicomService()) }
