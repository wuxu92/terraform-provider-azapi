// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package machinelearning

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Datastores provides resource knowledge for
// Microsoft.MachineLearningServices/workspaces/dataStores.
//
// One ARM type, three AzureRM resources discriminated by `properties.datastoreType`
// (see datastore.DatastoreResource.Properties, a discriminated union):
//   - azurerm_machine_learning_datastore_blobstorage      (datastoreType = "AzureBlob")
//   - azurerm_machine_learning_datastore_datalake_gen2     (datastoreType = "AzureDataLakeGen2")
//   - azurerm_machine_learning_datastore_fileshare         (datastoreType = "AzureFile")
//
// The three share an identical shape for the universal knowledge encoded here:
//   - name validation (validate.DataStoreName) is the same for all three.
//   - service_data_auth_identity enum (properties.serviceDataAccessAuthIdentity) is the
//     same StringInSlice on all three.
//   - `name` and `description` are ForceNew on all three; tags use TagsForceNew.
//   - is_default (properties.isDefault) defaults to false on all three.
//
// account_key / shared_access_signature are credentials nested under
// properties.credentials.secrets.* and are Sensitive; they are represented as
// SensitiveFields. The at-least-one-of(account_key, shared_access_signature) rule is a
// CustomizeDiff gated on service_data_auth_identity == None and cannot be expressed as a
// static RelationalRule, so it is documented here and left out.
//
// Sources:
//   - AzureRM internal/services/machinelearning/machine_learning_datastore_blobstorage_resource.go
//     (schema :61-124, Create :144-234, CustomizeDiff :126-142)
//   - AzureRM internal/services/machinelearning/machine_learning_datastore_datalake_gen2_resource.go
//     (datastoreType AzureDataLakeGen2 :177)
//   - AzureRM internal/services/machinelearning/machine_learning_datastore_fileshare_resource.go
//     (datastoreType AzureFile :191)
//   - AzureRM internal/services/machinelearning/validate/datastore_name.go
//   - go-azure-sdk .../machinelearningservices/2025-06-01/datastore:
//     model_datastoreresource.go, model_datastore.go, model_azureblobdatastore.go,
//     constants.go (DatastoreType, ServiceDataAccessAuthIdentity)
type Datastores struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Datastores)(nil)

func NewDatastores() *Datastores {
	return &Datastores{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.MachineLearningServices/workspaces/dataStores",
			ApiVersions:  []string{"2025-06-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.description"}, // description ForceNew on all three
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). validate.DataStoreName:
					// 1-255 chars, first char alphanumeric, then alphanumeric/hyphen/underscore.
					Regex:     `^[a-zA-Z0-9][a-zA-Z0-9\-_]{0,254}$`,
					MinLength: 1,
					MaxLength: 255,
					Message:   "must be 1-255 characters, start with a letter or digit, and contain only alphanumeric characters, hyphens, and underscores",
				},
				{
					// service_data_auth_identity (same enum on all three). Full ARM SDK set.
					PropertyPath:  "properties.serviceDataAccessAuthIdentity",
					AllowedValues: []string{"None", "WorkspaceSystemAssignedIdentity", "WorkspaceUserAssignedIdentity"},
					Message:       "must be one of None, WorkspaceSystemAssignedIdentity, or WorkspaceUserAssignedIdentity",
				},
			},
			SensitiveFields: []string{
				"properties.credentials.secrets.key",      // account_key (AzureBlob/AzureFile)
				"properties.credentials.secrets.sasToken", // shared_access_signature (AzureBlob)
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.isDefault", Value: false},                           // is_default default false
				{PropertyPath: "properties.serviceDataAccessAuthIdentity", Value: "None"},      // default None
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewDatastores()) }
