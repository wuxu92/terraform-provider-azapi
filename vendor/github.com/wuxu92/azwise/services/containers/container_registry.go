package containers

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ContainerRegistry provides resource knowledge for Microsoft.ContainerRegistry/registries.
//
// Contributing Terraform resource: azurerm_container_registry.
//
// Sources:
//   - terraform-provider-azurerm internal/services/containers/container_registry_resource.go
//     (schema L36-384, CustomizeDiff L290-375)
//   - terraform-provider-azurerm internal/services/containers/validate/container_registry_name.go
//   - go-azure-sdk resource-manager/containerregistry/2025-11-01/registries:
//     model_registryproperties.go, model_sku.go, model_policies.go,
//     model_retentionpolicy.go, model_networkruleset.go, model_iprule.go,
//     model_encryptionproperty.go, constants.go (enums).
//
// Notes:
//   - name is ForceNew but is an envelope property, not an ARM body field, so it is not
//     emitted as a body ForceNew rule.
//   - Many CustomizeDiff constraints are SKU-conditional (georeplications/quarantine/
//     retention/encryption/zone_redundancy require Premium; anonymous_pull requires
//     Standard/Premium) — value-conditional and not expressible declaratively, left as notes.
//   - admin_username / admin_password are Computed and returned by the ListCredentials API,
//     not the registry body, so they are not encoded here.
//   - georeplications are a separate ARM API (Microsoft.ContainerRegistry/registries/
//     replications), so their fields are not part of this body.
type ContainerRegistry struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ContainerRegistry)(nil)

// NewContainerRegistry returns knowledge for the registries resource.
func NewContainerRegistry() *ContainerRegistry {
	return &ContainerRegistry{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ContainerRegistry/registries",
			ApiVersions:  []string{"2025-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// zone_redundancy_enabled is ForceNew → properties.zoneRedundancy (Enabled/Disabled).
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.zoneRedundancy"},
			},
			StringRules: []azwise.StringRule{
				// ── Resource name ── validate.ContainerRegistryName
				{
					Regex:     `^[a-zA-Z0-9]+$`,
					MinLength: 5,
					MaxLength: 50,
					Message:   "alpha numeric characters only, 5-50 chars",
				},
				// ── sku → sku.name (full ARM SDK set)
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"Basic", "Classic", "Premium", "Standard"},
					Message:       "must be one of Basic, Classic, Premium or Standard",
				},
				// ── public_network_access_enabled → properties.publicNetworkAccess
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
				// ── network_rule_bypass_option → properties.networkRuleBypassOptions
				{
					PropertyPath:  "properties.networkRuleBypassOptions",
					AllowedValues: []string{"AzureServices", "None"},
					Message:       "must be one of AzureServices or None",
				},
				// ── role_assignment_mode → properties.roleAssignmentMode
				{
					PropertyPath:  "properties.roleAssignmentMode",
					AllowedValues: []string{"AbacRepositoryPermissions", "LegacyRegistryPermissions"},
					Message:       "must be one of AbacRepositoryPermissions or LegacyRegistryPermissions",
				},
				// ── zone_redundancy_enabled → properties.zoneRedundancy
				{
					PropertyPath:  "properties.zoneRedundancy",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
				// ── network_rule_set.default_action → properties.networkRuleSet.defaultAction
				{
					PropertyPath:  "properties.networkRuleSet.defaultAction",
					AllowedValues: []string{"Allow", "Deny"},
					Message:       "must be one of Allow or Deny",
				},
			},
			IntRules: []azwise.IntRule{
				// ── retention_policy_in_days → properties.policies.retentionPolicy.days
				{
					PropertyPath: "properties.policies.retentionPolicy.days",
					MinValue:     azwise.Ptr(int64(0)),
					MaxValue:     azwise.Ptr(int64(365)),
				},
			},
			DefaultValues: []azwise.DefaultValue{
				// admin_enabled Default false.
				{PropertyPath: "properties.adminUserEnabled", Value: false},
				// public_network_access_enabled Default true → Enabled.
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
				// network_rule_bypass_option Default AzureServices.
				{PropertyPath: "properties.networkRuleBypassOptions", Value: "AzureServices"},
				// network_rule_bypass_for_tasks_enabled Default false.
				{PropertyPath: "properties.networkRuleBypassAllowedForTasks", Value: false},
				// role_assignment_mode Default LegacyRegistryPermissions.
				{PropertyPath: "properties.roleAssignmentMode", Value: "LegacyRegistryPermissions"},
				// zone_redundancy_enabled Default false → Disabled.
				{PropertyPath: "properties.zoneRedundancy", Value: "Disabled"},
			},
			// sku.name is Required for creation.
			RequiredFields: []string{"sku.name"},
			// Read-only properties returned by GET but not settable in the create body.
			ComputedFields: []string{
				"properties.loginServer",
				"properties.creationDate",
				"properties.dataEndpointHostNames",
				"properties.privateEndpointConnections",
				"properties.provisioningState",
				"properties.status",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewContainerRegistry()) }
