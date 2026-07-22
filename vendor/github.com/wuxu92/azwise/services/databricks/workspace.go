package databricks

import (
	"strings"
	"time"

	"github.com/wuxu92/azwise"
)

// Workspace provides resource knowledge for Microsoft.Databricks/workspaces.
//
// Contributing TF resources:
//   - azurerm_databricks_workspace
//   - azurerm_databricks_workspace_customer_managed_key (applies managed-services CMK to
//     the workspace body: properties.encryption.entities.managedServices) — no distinct
//     ARM type, folded in here.
//   - azurerm_databricks_workspace_root_dbfs_customer_managed_key (applies root DBFS CMK to
//     the same workspace encryption body) — no distinct ARM type, folded in here.
//
// It overrides CheckForceNew to add the conditional SKU-to-`trial` replacement rule that a
// static ForceNew list cannot express (AzureRM CustomizeDiff, databricks_workspace_resource.go:403-410).
//
// Sources:
//   - AzureRM internal/services/databricks/databricks_workspace_resource.go
//     (schema :60-374, CustomizeDiff :376-454, Create :460-753, expand params :1359-1470)
//   - AzureRM internal/services/databricks/validate/workspace_name.go (name rule)
//   - go-azure-sdk .../databricks/2026-01-01/workspaces: model_workspace.go,
//     model_workspaceproperties.go, model_workspacecustomparameters.go,
//     model_workspacepropertiesencryption.go, model_managedidentityconfiguration.go,
//     model_sku.go, constants.go (enum values)
type Workspace struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Workspace)(nil)

// CheckForceNew extends BaseKnowledge with the conditional replacement rule for the
// SKU: a workspace can be upgraded between standard/premium in place, but changing the
// SKU to `trial` forces recreation (databricks_workspace_resource.go:403-410).
func (w *Workspace) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
	if w.BaseKnowledge.CheckForceNew(oldBody, newBody) {
		return true
	}

	oldSku := azwise.ExtractStringValue(oldBody, "sku.name")
	newSku := azwise.ExtractStringValue(newBody, "sku.name")
	if oldSku != "" && newSku != "" && !strings.EqualFold(oldSku, newSku) && strings.EqualFold(newSku, "trial") {
		return true
	}

	return false
}

func NewWorkspace() *Workspace {
	return &Workspace{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Databricks/workspaces",
			ApiVersions:  []string{"2023-09-01", "2025-01-01", "2025-06-01", "2026-01-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.managedResourceGroupId"}, // managed_resource_group_name
				{PropertyPath: "properties.parameters.requireInfrastructureEncryption.value"}, // infrastructure_encryption_enabled
				{PropertyPath: "properties.parameters.loadBalancerId.value"},                  // load_balancer_backend_address_pool_id (also loadBalancerBackendPoolName)
				{PropertyPath: "properties.parameters.amlWorkspaceId.value"},                  // custom_parameters.machine_learning_workspace_id
				{PropertyPath: "properties.parameters.natGatewayName.value"},                  // custom_parameters.nat_gateway_name
				{PropertyPath: "properties.parameters.publicIpName.value"},                    // custom_parameters.public_ip_name
				{PropertyPath: "properties.parameters.customPublicSubnetName.value"},          // custom_parameters.public_subnet_name
				{PropertyPath: "properties.parameters.customPrivateSubnetName.value"},         // custom_parameters.private_subnet_name
				{PropertyPath: "properties.parameters.customVirtualNetworkId.value"},          // custom_parameters.virtual_network_id
				{PropertyPath: "properties.parameters.storageAccountName.value"},              // custom_parameters.storage_account_name
				{PropertyPath: "properties.parameters.vnetAddressPrefix.value"},               // custom_parameters.vnet_address_prefix
				// NOTE: sku -> `trial` is conditionally ForceNew; handled in CheckForceNew.
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
					// Resource name (empty PropertyPath). validate.WorkspaceName:
					// 3-64 chars, alphanumeric plus underscore and hyphen.
					Regex:     `^[a-zA-Z0-9_-]*$`,
					MinLength: 3,
					MaxLength: 64,
					Message:   "must be 3-64 characters and contain only alphanumeric characters, underscores, and hyphens",
				},
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"standard", "premium", "trial"},
					Message:       "must be one of standard, premium, or trial",
				},
				{
					PropertyPath:  "properties.requiredNsgRules",
					AllowedValues: []string{"AllRules", "NoAzureDatabricksRules", "NoAzureServiceRules"},
					Message:       "must be one of AllRules, NoAzureDatabricksRules, or NoAzureServiceRules",
				},
				// NOTE: compliance_security_profile_standards validates an array element
				// (properties.enhancedSecurityCompliance.complianceSecurityProfile.complianceStandards[*],
				// enum HIPAA/PCI_DSS/FEDRAMP_MODERATE/IRAP_PROTECTED/FEDRAMP_HIGH/FEDRAMP_IL5/
				// ITAR_EAR/CYBER_ESSENTIAL_PLUS/CANADA_PROTECTED_B/ISMAP/HITRUST/K_FSI/
				// GERMANY_C5/GERMANY_TISAX). Array-element enum paths are not representable
				// as a declarative StringRule and are intentionally omitted.
			},
			SensitiveFields: []string{
				"properties.managedDiskIdentity.principalId",
				"properties.managedDiskIdentity.tenantId",
				"properties.storageAccountIdentity.principalId",
				"properties.storageAccountIdentity.tenantId",
			},
			// Azure-populated, read-only fields absent from the user-provided create body.
			ComputedFields: []string{
				"properties.workspaceId",
				"properties.workspaceUrl",
				"properties.diskEncryptionSetId",
				"properties.provisioningState",
				"properties.managedDiskIdentity",
				"properties.storageAccountIdentity",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},                           // public_network_access_enabled = true
				{PropertyPath: "properties.parameters.prepareEncryption.value", Value: false},                // customer_managed_key_enabled = false
				{PropertyPath: "properties.parameters.requireInfrastructureEncryption.value", Value: false},  // infrastructure_encryption_enabled = false
				{PropertyPath: "properties.parameters.enableNoPublicIp.value", Value: true},                  // custom_parameters.no_public_ip = true
			},
			RequiredFields: []string{
				"sku.name",
				// AzureRM auto-generates a managed resource group name when omitted, but the
				// ARM API requires managedResourceGroupId to be present (databricks_workspace_resource.go:520-526).
				"properties.managedResourceGroupId",
			},
			// default_storage_firewall_enabled <-> access_connector_id are mutually RequiredWith
			// (schema :135-145). Both map cleanly to ARM body paths.
			RequiredWith: []azwise.RelationalRule{
				{Paths: []string{"properties.defaultStorageFirewall", "properties.accessConnector.id"}},
				{Paths: []string{"properties.accessConnector.id", "properties.defaultStorageFirewall"}},
				// NOTE: managed_disk_cmk_rotation_to_latest_version_enabled RequiredWith
				// managed_disk_cmk_key_vault_key_id (schema :304-308) is NOT encoded: the key
				// vault key ID is a composite Key Vault key URI that ARM splits across
				// keyName/keyVaultUri/keyVersion under
				// properties.encryption.entities.managedDisk.keyVaultProperties — there is no
				// single ARM body field to reference, so the relation is non-mappable.
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewWorkspace()) }
