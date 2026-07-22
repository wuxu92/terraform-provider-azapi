package netapp

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetAppAccount provides resource knowledge for Microsoft.NetApp/netAppAccounts.
//
// Mirrors azurerm_netapp_account. The azurerm_netapp_account_encryption resource
// mutates the SAME ARM resource (netAppAccounts): it folds customer-managed-key
// encryption into properties.encryption of the account body, so its constraints
// are amended into this file rather than emitted separately.
//
// Sources:
//   - internal/services/netapp/netapp_account_resource.go
//     (schema 50-165: name ForceNew + AccountName validator; location/identity/tags;
//     active_directory block Optional MaxItems 1; timeouts Create/Update/Delete 30m Read 5m;
//     create 169-227 → AccountProperties{ActiveDirectories}).
//   - internal/services/netapp/netapp_account_encryption_resource.go
//     (schema 43-91: user_assigned_identity_id, system_assigned_identity_principal_id,
//     encryption_key, federated_client_id, cross_tenant_key_vault_resource_id;
//     expandEncryption 304-359 → AccountEncryption{KeySource, Identity{UserAssignedIdentity,
//     FederatedClientId}, KeyVaultProperties{KeyName, KeyVaultUri, KeyVaultResourceId}}).
//   - internal/services/netapp/validate/account_name.go (regex ^[-_\da-zA-Z]{3,64}$).
//   - go-azure-sdk resource-manager/netapp/2026-01-01/netappaccounts:
//     model_accountproperties.go (activeDirectories/disableShowmount/encryption/
//     multiAdStatus/nfsV4IDDomain; provisioningState read-only),
//     model_accountencryption.go (identity/keySource/keyVaultProperties),
//     constants.go (KeySource Microsoft.KeyVault/Microsoft.NetApp),
//     id_netappaccount.go (type segment casing "netAppAccounts").
type NetAppAccount struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetAppAccount)(nil)

// NewNetAppAccount returns knowledge for the netAppAccounts resource.
func NewNetAppAccount() *NetAppAccount {
	return &NetAppAccount{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.NetApp/netAppAccounts",
			ApiVersions:  []string{"2026-01-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "location"},
			},
			StringRules: []azwise.StringRule{
				{
					// AzureRM AccountName: 3-64 chars, letters/numbers/underscore/hyphen.
					Regex:     `^[-_\da-zA-Z]{3,64}$`,
					MinLength: 3,
					MaxLength: 64,
					Message:   "must be 3-64 characters and contain only letters, numbers, underscores and hyphens",
				},
				{
					// Folded in from azurerm_netapp_account_encryption (expandEncryption).
					PropertyPath:  "properties.encryption.keySource",
					AllowedValues: []string{"Microsoft.KeyVault", "Microsoft.NetApp"},
				},
			},
			// Server-populated, read-only ARM property.
			ComputedFields: []string{
				"properties.provisioningState",
			},
			// NOTE: active_directory maps to properties.activeDirectories[*] (an array of
			// objects). Its element-level validators (domain/smb_server_name regexes,
			// IPv4 dns_servers/kdc_ip, RequiredWith ldap_over_tls<->server_root_ca_certificate,
			// defaults organizational_unit="CN=Computers"/site_name="Default-First-Site-Name")
			// are array-element paths and cannot be expressed as declarative rules here.
			// NOTE: the encryption resource's user_assigned_identity_id ConflictsWith
			// system_assigned_identity_principal_id is a Terraform-config-level constraint;
			// system_assigned_identity_principal_id has no ARM body encryption path (it is
			// read from the account's SystemAssigned identity), so it is not mappable.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewNetAppAccount()) }
