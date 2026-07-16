package azwise

import "time"

// DataFactory provides resource knowledge for Microsoft.DataFactory/factories.
//
// Sources:
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_resource.go:32-224
//     (azurerm_data_factory schema: ForceNew, timeouts, validators, defaults,
//     github/vsts ConflictsWith, CustomizeDiff conditional-ForceNew + CMK check)
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_resource.go:226-335
//     (create mapping to factories.Factory / FactoryProperties: publicNetworkAccess,
//     encryption, purviewConfiguration, globalParameters; separate repo + managed-vnet APIs)
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_resource.go:337-443
//     (read/flatten: read-only createTime/provisioningState/version, publicNetworkAccess default)
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_resource.go:445-460
//     (delete: plain Delete, no purge/recover — not soft-delete)
//   - terraform-provider-azurerm internal/services/datafactory/data_factory_resource.go:462-616
//     (global_parameter + github/vsts repoConfiguration expand/flatten)
//   - terraform-provider-azurerm internal/services/datafactory/validate/datafactory.go:25-34
//     (DataFactoryName regex) and :52-64 (CMKIdentityIdRequiredAtCreation)
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/factories/model_factory.go:11-21
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/factories/model_factoryproperties.go:14-23
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/factories/model_encryptionconfiguration.go:6-11
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/factories/model_cmkidentitydefinition.go:6-8
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/factories/model_purviewconfiguration.go:6-8
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/factories/model_factorygithubconfiguration.go:13-27
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/factories/model_factoryvstsconfiguration.go:13-26
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/factories/model_githubclientsecret.go:6-9
//   - terraform-provider-azurerm vendor/.../datafactory/2018-06-01/factories/constants.go:12-77
//     (GlobalParameterType, PublicNetworkAccess enum values)
//
// Intentionally skipped here:
//   - resource_group_name: AzureRM marks it (envelope), but it is an AzAPI ID
//     segment, not a Microsoft.DataFactory/factories body property.
//   - identity: commonschema.SystemAssignedUserAssignedIdentityOptional maps to the
//     top-level envelope `identity` block (Factory.Identity), not a properties.* body
//     field; no azwise rule applies.
//   - github_configuration ConflictsWith vsts_configuration (and the reverse): both
//     lower to the discriminated properties.repoConfiguration union
//     (FactoryGitHubConfiguration / FactoryVSTSConfiguration). The generator already
//     emits an AtMostOneOf over the two variant blocks, so re-emitting a RelationalRule
//     would duplicate that constraint — deliberately omitted.
//   - managed_virtual_network_enabled: ForceNewIfChange(true->false) in CustomizeDiff,
//     but it is a *separate* ARM API resource
//     (Microsoft.DataFactory/factories/managedVirtualNetworks/default, configured via
//     the ManagedVirtualNetworks client), not a factory body property — a sub-service
//     concern that does not belong on this parent's knowledge file.
//   - global_parameter[].type StringInSlice(PossibleValuesForGlobalParameterType):
//     maps to properties.globalParameters.<key>.type, a map-keyed path
//     (map[string]GlobalParameterSpecification). azwise cannot express a map-value
//     enum, so the [Array,Bool,Float,Int,Object,String] rule is omitted.
//   - customer_managed_key_id keyvault.ValidateNestedItemID: a composite Key Vault key
//     URI that azurerm splits across properties.encryption.vaultBaseUrl / .keyName /
//     .keyVersion. There is no single ARM body field to validate, so the URI validator
//     is not mappable (see skill "non-mappable" case) and is omitted.
//   - github/vsts publishing_enabled Default:true -> properties.repoConfiguration.disablePublish
//     default false: lives inside the discriminated repoConfiguration union whose paths
//     do not resolve through the SDK interface field; omitted as a DefaultValue.
type DataFactory struct {
	BaseKnowledge
}

var _ ResourceKnowledge = (*DataFactory)(nil)

// NewDataFactory returns knowledge for the Microsoft.DataFactory/factories resource.
func NewDataFactory() *DataFactory {
	return &DataFactory{
		BaseKnowledge: BaseKnowledge{
			ResourceType: "Microsoft.DataFactory/factories",
			ApiVersions:  []string{"2018-06-01"},
			SoftDelete:   false,
			ForceNew: []ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
			},
			TimeoutsConfig: &Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []StringRule{
				{
					// Envelope resource name (validate.DataFactoryName). AzureRM enforces
					// only the regex; the 3-63 length reflects Azure's documented factory
					// naming limits.
					Regex:     `^[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*$`,
					MinLength: 3,
					MaxLength: 63,
					Message:   "must be 3-63 characters, alphanumeric segments separated by single hyphens (see Azure Data Factory naming rules)",
				},
				{
					PropertyPath:  "properties.publicNetworkAccess",
					AllowedValues: []string{"Disabled", "Enabled"},
					Message:       "must be one of Disabled or Enabled",
				},
				{
					// purview_id (account.ValidateAccountID) — a Microsoft.Purview account
					// resource ID. Expressed as a regex here; ideally a shared
					// AzureResourceID() validator in a customizer.
					PropertyPath: "properties.purviewConfiguration.purviewResourceId",
					Regex:        `(?i)^/subscriptions/[^/]+/resourceGroups/[^/]+/providers/Microsoft\.Purview/accounts/[^/]+$`,
					Message:      "must be a Microsoft.Purview account resource ID",
				},
				{
					// customer_managed_key_identity_id (commonids.ValidateUserAssignedIdentityID).
					PropertyPath: "properties.encryption.identity.userAssignedIdentity",
					Regex:        `(?i)^/subscriptions/[^/]+/resourceGroups/[^/]+/providers/Microsoft\.ManagedIdentity/userAssignedIdentities/[^/]+$`,
					Message:      "must be a Microsoft.ManagedIdentity user-assigned identity resource ID",
				},
			},
			// GitHub repo integration carries a client secret referencing a Key Vault
			// secret (FactoryGitHubConfiguration.ClientSecret -> GitHubClientSecret).
			// AzureRM does not surface these fields, but the ARM API accepts them and
			// they bear credential references, so mark them sensitive. Paths live under
			// the discriminated properties.repoConfiguration union (not resolvable by
			// azwise_validate's struct walk, hence not path-checked).
			SensitiveFields: []string{
				"properties.repoConfiguration.clientSecret.byoaSecretAkvUrl",
				"properties.repoConfiguration.clientSecret.byoaSecretName",
			},
			// Read-only ARM response fields (never set on create/update).
			ComputedFields: []string{
				"properties.createTime",
				"properties.provisioningState",
				"properties.version",
			},
			DefaultValues: []DefaultValue{
				// public_network_enabled Default:true -> publicNetworkAccess Enabled.
				{PropertyPath: "properties.publicNetworkAccess", Value: "Enabled"},
			},
			// CMKIdentityIdRequiredAtCreation (validate/datafactory.go:52-64): configuring
			// customer-managed-key encryption requires a user-assigned identity to reach
			// the key vault. keyName is a required member whenever the encryption block is
			// present, so its presence is a faithful proxy for "CMK configured".
			RequiredWith: []RelationalRule{
				{
					Paths:   []string{"properties.encryption.keyName", "properties.encryption.identity.userAssignedIdentity"},
					Message: "customer-managed-key encryption requires a user-assigned identity (properties.encryption.identity.userAssignedIdentity)",
				},
			},
		},
	}
}
