package authorization

import (
	"fmt"

	"github.com/Azure/terraform-provider-azapi/internal/native/services/config"
	"github.com/Azure/terraform-provider-azapi/internal/native/services/resources"
	"github.com/hashicorp/go-uuid"
)

const (
	readerRoleDefinitionID                      = "acdd72a7-3385-48ef-bd42-f606fba81ae7"
	keyVaultAdministratorRoleDefinitionID       = "00482a5a-887f-4fb3-b363-3b7fe8e74483"
	keyVaultCertificatesOfficerRoleDefinitionID = "a4417e6f-fecd-4de8-b567-7b0420556985"
	keyVaultCryptoOfficerRoleDefinitionID       = "14b46e9e-c2b7-41b4-b07b-48a6ebf60603"
	keyVaultSecretsOfficerRoleDefinitionID      = "b86a8fe4-44ce-4948-aee5-eccb2c155cd7"
)

// RoleAssignmentCfg carries Terraform address metadata, a stable role-assignment GUID,
// and the shared client_config data source used by role-assignment acceptance scenarios.
// Construct it with NewRoleAssignmentCfg, then wrap it in RoleAssignmentCfg_Basic when
// applying. The basic scenario assigns a built-in Reader role at subscription scope to
// the identity currently running the provider.
type RoleAssignmentCfg struct {
	config.ResourceConfigBase
	name         string
	clientConfig config.ClientConfigData
	scopeRef     string
}

// NewRoleAssignmentCfg builds a role-assignment config scoped at the current
// subscription. The generated GUID is the ARM resource name; holding it on the config
// keeps re-applies idempotent within one acceptance run. The label is optional — omit it
// for the single-instance default ("test"), or pass an explicit label when a scope holds
// more than one. The resource type is read from the RoleAssignment descriptor.
func NewRoleAssignmentCfg(clientConfig config.ClientConfigData, label ...string) RoleAssignmentCfg {
	guid, err := uuid.GenerateUUID()
	if err != nil {
		panic(fmt.Sprintf("authorization: generating role assignment GUID: %v", err))
	}
	return RoleAssignmentCfg{
		ResourceConfigBase: config.NewResourceConfigBase(RoleAssignment.Name, label...),
		name:               guid,
		clientConfig:       clientConfig,
	}
}

// KeyVaultAdministratorRoleAssignmentCfg grants Key Vault Administrator at a resource
// group scope to the current provider identity. The role ID is from Microsoft Learn's
// Key Vault data-plane RBAC built-in roles table.
type KeyVaultAdministratorRoleAssignmentCfg RoleAssignmentCfg

func NewKeyVaultAdministratorRoleAssignmentCfg(resourceGroup resources.ResourceGroupCfg, clientConfig config.ClientConfigData, label ...string) KeyVaultAdministratorRoleAssignmentCfg {
	return KeyVaultAdministratorRoleAssignmentCfg(newKeyVaultRoleAssignmentCfg(resourceGroup, clientConfig, "key_vault_administrator", label...))
}

func (r KeyVaultAdministratorRoleAssignmentCfg) Config() string {
	return RoleAssignmentCfg(r).keyVaultDataPlaneRoleConfig(keyVaultAdministratorRoleDefinitionID)
}

// KeyVaultCertificatesOfficerRoleAssignmentCfg grants Key Vault Certificates Officer
// at a resource group scope to the current provider identity.
type KeyVaultCertificatesOfficerRoleAssignmentCfg RoleAssignmentCfg

func NewKeyVaultCertificatesOfficerRoleAssignmentCfg(resourceGroup resources.ResourceGroupCfg, clientConfig config.ClientConfigData, label ...string) KeyVaultCertificatesOfficerRoleAssignmentCfg {
	return KeyVaultCertificatesOfficerRoleAssignmentCfg(newKeyVaultRoleAssignmentCfg(resourceGroup, clientConfig, "key_vault_certificates_officer", label...))
}

func (r KeyVaultCertificatesOfficerRoleAssignmentCfg) Config() string {
	return RoleAssignmentCfg(r).keyVaultDataPlaneRoleConfig(keyVaultCertificatesOfficerRoleDefinitionID)
}

// KeyVaultCryptoOfficerRoleAssignmentCfg grants Key Vault Crypto Officer at a
// resource group scope to the current provider identity.
type KeyVaultCryptoOfficerRoleAssignmentCfg RoleAssignmentCfg

func NewKeyVaultCryptoOfficerRoleAssignmentCfg(resourceGroup resources.ResourceGroupCfg, clientConfig config.ClientConfigData, label ...string) KeyVaultCryptoOfficerRoleAssignmentCfg {
	return KeyVaultCryptoOfficerRoleAssignmentCfg(newKeyVaultRoleAssignmentCfg(resourceGroup, clientConfig, "key_vault_crypto_officer", label...))
}

func (r KeyVaultCryptoOfficerRoleAssignmentCfg) Config() string {
	return RoleAssignmentCfg(r).keyVaultDataPlaneRoleConfig(keyVaultCryptoOfficerRoleDefinitionID)
}

// KeyVaultSecretsOfficerRoleAssignmentCfg grants Key Vault Secrets Officer at a
// resource group scope to the current provider identity.
type KeyVaultSecretsOfficerRoleAssignmentCfg RoleAssignmentCfg

func NewKeyVaultSecretsOfficerRoleAssignmentCfg(resourceGroup resources.ResourceGroupCfg, clientConfig config.ClientConfigData, label ...string) KeyVaultSecretsOfficerRoleAssignmentCfg {
	return KeyVaultSecretsOfficerRoleAssignmentCfg(newKeyVaultRoleAssignmentCfg(resourceGroup, clientConfig, "key_vault_secrets_officer", label...))
}

func (r KeyVaultSecretsOfficerRoleAssignmentCfg) Config() string {
	return RoleAssignmentCfg(r).keyVaultDataPlaneRoleConfig(keyVaultSecretsOfficerRoleDefinitionID)
}

func newKeyVaultRoleAssignmentCfg(resourceGroup resources.ResourceGroupCfg, clientConfig config.ClientConfigData, defaultLabel string, label ...string) RoleAssignmentCfg {
	if len(label) == 0 {
		label = []string{defaultLabel}
	}
	cfg := NewRoleAssignmentCfg(clientConfig, label...)
	cfg.scopeRef = resourceGroup.IDRef()
	return cfg
}

func (r RoleAssignmentCfg) keyVaultDataPlaneRoleConfig(roleDefinitionID string) string {
	return r.config(fmt.Sprintf(`
  properties = {
    principal_id       = %[1]s
    role_definition_id = "/subscriptions/${%[2]s}/providers/Microsoft.Authorization/roleDefinitions/%[3]s"
  }`, r.clientConfig.RefOf("object_id"), r.clientConfig.RefOf("subscription_id"), roleDefinitionID), r.scopeRef)
}

// RoleAssignmentCfg_Basic assigns the built-in Reader role at subscription scope to the
// current provider identity. It exercises the native role-assignment envelope
// (scope_id is the assignment scope) and the two required ARM body properties.
type RoleAssignmentCfg_Basic RoleAssignmentCfg

func (r RoleAssignmentCfg_Basic) Config() string {
	ra := RoleAssignmentCfg(r)
	return ra.config(fmt.Sprintf(`
  properties = {
    principal_id       = %[1]s
    role_definition_id = "/subscriptions/${%[2]s}/providers/Microsoft.Authorization/roleDefinitions/%[3]s"
  }`, ra.clientConfig.RefOf("object_id"), ra.clientConfig.RefOf("subscription_id"), readerRoleDefinitionID))
}

func (r RoleAssignmentCfg) config(body string, scopes ...string) string {
	scope := fmt.Sprintf(`"/subscriptions/${%s}"`, r.clientConfig.RefOf("subscription_id"))
	if len(scopes) > 0 {
		scope = scopes[0]
	} else if r.scopeRef != "" {
		scope = r.scopeRef
	}
	return r.RenderConfig(config.ConfigEnvelope{
		Name:       r.name,
		ParentAttr: "scope_id",
		ParentRef:  scope,
		Body:       body,
	})
}
