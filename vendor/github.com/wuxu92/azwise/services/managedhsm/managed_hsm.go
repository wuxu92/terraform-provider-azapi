package managedhsm

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ManagedHSM provides resource knowledge for Microsoft.KeyVault/managedHSMs.
//
// Contributing Terraform resource: azurerm_key_vault_managed_hardware_security_module.
//
// Sources:
//   - terraform-provider-azurerm internal/services/managedhsm/key_vault_managed_hardware_security_module_resource.go
//     (schema L38-184, Create body L186-266, network_acls expand L446-457, CustomizeDiff L557-571)
//   - terraform-provider-azurerm internal/services/managedhsm/validate/managed_hsm_name.go
//     (ManagedHardwareSecurityModuleName L12-26)
//   - go-azure-sdk resource-manager/keyvault/2026-02-01/managedhsms:
//     model_managedhsm.go, model_managedhsmproperties.go, model_managedhsmsku.go,
//     model_mhsmnetworkruleset.go, constants.go
//
// Notes:
//   - name/location/resource_group_name/tags are envelope-owned; not emitted as body rules.
//   - AzureRM restricts sku_name to Standard_B1, but azwise emits the full ManagedHsmSkuName
//     SDK enum since AzAPI sends raw ARM values.
//   - sku.family is hardcoded to "B" by AzureRM (Create L222); emitted as a required field
//     with a default value.
//   - public_network_access_enabled is a bool that AzureRM maps one-to-many onto the
//     properties.publicNetworkAccess enum (Enabled/Disabled); default Enabled.
//   - tenant_id (properties.tenantId) and each admin_object_ids element use validation.IsUUID.
//     tenant_id is a semantic UUID check better expressed as an azapin customizer validator
//     (typegraph.Validator(validators.UUID) on properties.tenantId); admin_object_ids is an
//     array-element path (properties.initialAdminObjectIds[*]) so it is not emitted here.
//   - security_domain_key_vault_certificate_ids / security_domain_quorum /
//     security_domain_encrypted_data drive the data-plane security-domain download
//     (securityDomainDownload, kermit keyvault 7.4 client), not the ARM body; omitted. Their
//     conditional ForceNew (CustomizeDiff L557-571) is likewise data-plane, not emitted.
//   - The sibling TF resources key/key_rotation_policy/role_assignment/role_definition are
//     Managed HSM DATA-PLANE resources (jackofallops/kermit keyvault 7.4 client), not ARM
//     resource-manager types, so they have no ARM resource body and are intentionally skipped.
type ManagedHSM struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ManagedHSM)(nil)

// NewManagedHSM returns knowledge for the managedHSMs resource.
func NewManagedHSM() *ManagedHSM {
	return &ManagedHSM{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.KeyVault/managedHSMs",
			ApiVersions:  []string{"2026-02-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 60 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 60 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "sku.name"},
				{PropertyPath: "properties.initialAdminObjectIds"},
				{PropertyPath: "properties.tenantId"},
				{PropertyPath: "properties.enablePurgeProtection"},
				{PropertyPath: "properties.softDeleteRetentionInDays"},
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[a-zA-Z][-a-zA-Z\d]{1,22}[a-zA-Z\d]$`,
					Message:      "name must begin with a letter, end with a letter or number, contain only alphanumeric characters and hyphens (no consecutive hyphens), and be between 3 and 24 characters long",
				},
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Custom_B6", "Custom_B32", "Custom_C42", "Custom_C10", "Standard_B1"},
					Message:       "sku_name must be one of Custom_B6, Custom_B32, Custom_C42, Custom_C10 or Standard_B1",
				},
				{
					PropertyPath:  "properties.networkAcls.defaultAction",
					AllowedValues: []string{"Allow", "Deny"},
					Message:       "network_acls.default_action must be Allow or Deny",
				},
				{
					PropertyPath:  "properties.networkAcls.bypass",
					AllowedValues: []string{"None", "AzureServices"},
					Message:       "network_acls.bypass must be None or AzureServices",
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "public network access must be Enabled or Disabled",
				},
			},
			IntRules: []azwise.IntRule{
				{PropertyPath: "properties.softDeleteRetentionInDays", MinValue: azwise.Ptr(int64(7)), MaxValue: azwise.Ptr(int64(90))},
			},
			ComputedFields: []string{
				"properties.hsmUri",
				"properties.provisioningState",
				"properties.statusMessage",
				"properties.scheduledPurgeDate",
				"properties.securityDomainProperties",
				"properties.privateEndpointConnections",
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "sku.family", Value: "B"},
				{PropertyPath: "properties.softDeleteRetentionInDays", Value: 90},
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
			},
			RequiredFields: []string{
				"sku.name",
				"sku.family",
				"properties.initialAdminObjectIds",
				"properties.tenantId",
			},
		},
	}
}

func init() { azwise.Register(NewManagedHSM()) }
