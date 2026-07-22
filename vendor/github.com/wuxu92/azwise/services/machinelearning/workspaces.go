// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package machinelearning

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Workspaces provides resource knowledge for Microsoft.MachineLearningServices/workspaces.
//
// One ARM type, three AzureRM resources discriminated by the top-level `kind`:
//   - azurerm_machine_learning_workspace   (kind = "Default" / "FeatureStore")
//   - azurerm_ai_foundry                   (kind = "Hub")
//   - azurerm_ai_foundry_project           (kind = "Project")
//
// Only knowledge that is universal (or harmless) across every kind is encoded here.
// Kind-specific rules are intentionally NOT unioned because they would corrupt
// validation for the other contributors:
//   - RequiredFields (workspace requires applicationInsights/keyVault/storageAccount;
//     ai_foundry requires keyVault/storageAccount; ai_foundry_project requires only
//     hubResourceId) — none are required for EVERY body, so RequiredFields is empty.
//   - DefaultValues (sku.name=Basic, systemDatastoresAuthMode=AccessKey,
//     v1LegacyMode=false, publicNetworkAccess=Enabled, kind default) only apply to the
//     Default kind body; GetDefaultValues would inject them into Hub/Project bodies that
//     never carry those fields, so DefaultValues is empty.
//   - ForceNew on properties.applicationInsights and properties.containerRegistry is
//     excluded: those are ForceNew for azurerm_machine_learning_workspace but freely
//     updatable for azurerm_ai_foundry — a genuine conflict.
//
// ForceNew that IS safe to union: `name` (ForceNew for all three) plus properties that
// only the Default/Hub kinds carry and force-new consistently (keyVault, storageAccount,
// encryption, hbiWorkspace, enableServiceSideCMKEncryption, provisionNetworkNow) — these
// never appear in a Project body, so the rule simply never fires there.
//
// Sources:
//   - AzureRM internal/services/machinelearning/machine_learning_workspace_resource.go
//     (schema :56-281, Create :285-395, encryption expand :697-719, feature_store :755-777,
//     managed_network :808-818, serverless_compute :835-851)
//   - AzureRM internal/services/machinelearning/ai_foundry_resource.go
//     (schema :102-239, Create :241-348, kind="Hub" :287)
//   - AzureRM internal/services/machinelearning/ai_foundry_project_resource.go
//     (kind="Project" :177, HubResourceId :179, name regex :85-87)
//   - AzureRM internal/services/machinelearning/validate/workspace_name.go (name rule)
//   - go-azure-sdk .../machinelearningservices/2025-06-01/workspaces:
//     model_workspace.go, model_workspaceproperties.go, model_encryptionproperty.go,
//     constants.go (IsolationMode, PublicNetworkAccess, SystemDatastoresAuthMode)
type Workspaces struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Workspaces)(nil)

func NewWorkspaces() *Workspaces {
	return &Workspaces{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.MachineLearningServices/workspaces",
			ApiVersions:  []string{"2025-06-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.keyVault"},                       // key_vault_id (Default + Hub)
				{PropertyPath: "properties.storageAccount"},                 // storage_account_id (Default + Hub)
				{PropertyPath: "properties.encryption"},                     // encryption block (Default + Hub)
				{PropertyPath: "properties.hbiWorkspace"},                   // high_business_impact(_enabled) (Default + Hub)
				{PropertyPath: "properties.enableServiceSideCMKEncryption"}, // service_side_encryption_enabled (Default only)
				{PropertyPath: "properties.provisionNetworkNow"},            // managed_network.provision_on_creation_enabled (Default only)
				// NOTE: properties.applicationInsights and properties.containerRegistry are
				// ForceNew for azurerm_machine_learning_workspace but updatable for
				// azurerm_ai_foundry — conflicting, so deliberately NOT unioned here.
			},
			SoftDelete: false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute, // ai_foundry (Hub) uses 60m; workspace/project 30m
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath). validate.WorkspaceName and the
					// ai_foundry_project name regex are equivalent: first char alphanumeric,
					// then 2-32 of alphanumeric/underscore/hyphen (3-33 chars total).
					Regex:     `^[a-zA-Z0-9][a-zA-Z0-9_-]{2,32}$`,
					MinLength: 3,
					MaxLength: 33,
					Message:   "must be 3-33 characters, start with a letter or digit, and contain only alphanumeric characters, underscores, and hyphens",
				},
				{
					// public_network_access(_enabled). Only fires when the field is present.
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "must be one of Disabled or Enabled",
				},
				{
					// storage_account_access_type (Default kind only). Full ARM SDK set.
					PropertyPath:  "properties.systemDatastoresAuthMode",
					AllowedValues: []string{"AccessKey", "Identity", "UserDelegationSAS"},
					Message:       "must be one of AccessKey, Identity, or UserDelegationSAS",
				},
				{
					// managed_network.isolation_mode (Default + Hub). Full ARM SDK set.
					PropertyPath:  "properties.managedNetwork.isolationMode",
					AllowedValues: []string{"AllowInternetOutbound", "AllowOnlyApprovedOutbound", "Disabled"},
					Message:       "must be one of AllowInternetOutbound, AllowOnlyApprovedOutbound, or Disabled",
				},
				{
					// sku_name (Default kind only). AzureRM restricts to Basic.
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Basic"},
					Message:       "must be Basic",
				},
				// NOTE: `kind` is intentionally not constrained: its valid values differ per
				// contributing resource (Default/FeatureStore vs Hub vs Project). A single
				// AllowedValues set would reject legitimate values for the others.
			},
			// discovery_url and workspace_id are server-computed (schema Computed:true).
			ComputedFields: []string{
				"properties.discoveryUrl",
				"properties.workspaceId",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewWorkspaces()) }
