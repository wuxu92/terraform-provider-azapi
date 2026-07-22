package domainservices

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ActiveDirectoryDomainService provides resource knowledge for
// Microsoft.AAD/domainServices.
//
// Sources:
//   - terraform-provider-azurerm internal/services/domainservices/active_directory_domain_service_resource.go:33-300
//     (azurerm_active_directory_domain_service schema: ForceNew fields, timeouts,
//     sku/domain_configuration_type enums, sensitive pfx fields, filtered_sync default)
//   - terraform-provider-azurerm internal/services/domainservices/active_directory_domain_service_resource.go:302-457
//     (create/update mapping to domainservices.DomainService / DomainServiceProperties:
//     domainName, sku, filteredSync, domainConfigurationType, replicaSets on create)
//   - terraform-provider-azurerm internal/services/domainservices/active_directory_domain_service_resource.go:542-559
//     (delete: plain DeleteThenPoll, no purge/recover — not soft-delete)
//   - terraform-provider-azurerm internal/services/domainservices/validate/domain_service_name.go:11-24
//     (DomainServiceName FQDN regex on domain_name, not the envelope name)
//   - terraform-provider-azurerm vendor/.../aad/2021-05-01/domainservices/model_domainserviceproperties.go:6-23
//   - terraform-provider-azurerm vendor/.../aad/2021-05-01/domainservices/model_domainsecuritysettings.go:6-14
//   - terraform-provider-azurerm vendor/.../aad/2021-05-01/domainservices/model_notificationsettings.go:6-10
//   - terraform-provider-azurerm vendor/.../aad/2021-05-01/domainservices/model_ldapssettings.go:12-20
//   - terraform-provider-azurerm vendor/.../aad/2021-05-01/domainservices/model_replicaset.go:6-17
//   - terraform-provider-azurerm vendor/.../aad/2021-05-01/domainservices/constants.go
//     (Enabled/Disabled enums: ExternalAccess, FilteredSync, KerberosArmoring,
//     KerberosRc4Encryption, Ldaps, NotifyDcAdmins, NotifyGlobalAdmins, NtlmV1,
//     SyncKerberosPasswords, SyncNtlmPasswords, SyncOnPremPasswords, TlsV1)
//
// Merged/related TF resources (no distinct ARM resource type — see report):
//   - azurerm_active_directory_domain_service_replica_set: manages entries of the
//     domainService properties.replicaSets sub-array via CreateOrUpdate/Update on the
//     parent domainService. No separate ARM type. Its only user-settable value
//     (subnet_id) is an array-element path (properties.replicaSets[*].subnetId), which
//     azwise cannot express as a rule — skipped, documented below.
//   - azurerm_active_directory_domain_service_trust: manages the resourceForestSettings
//     trusts configuration (properties.resourceForestSettings.settings[*]) of the parent
//     domainService, again via Update. No separate ARM type; its fields are all
//     array-element paths — skipped, documented below.
//
// Intentionally skipped here:
//   - name / location / resource_group_name: envelope fields, not
//     Microsoft.AAD/domainServices body properties.
//   - initial_replica_set.subnet_id (ForceNew, commonids.ValidateSubnetID): maps to the
//     array-element path properties.replicaSets[*].subnetId. azwise cannot lower a rule
//     through an array element, so neither the ForceNew nor the subnet-ID validator is
//     emitted here.
//   - secure_ldap.pfx_certificate (azValidate.Base64EncodedString) and
//     notifications.additional_recipients element (validation.StringIsNotWhiteSpace):
//     semantic validators over a sensitive/array-element value; not expressible as a
//     declarative StringRule and left to an azapin customizer if needed.
type ActiveDirectoryDomainService struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ActiveDirectoryDomainService)(nil)

// NewActiveDirectoryDomainService returns knowledge for the
// Microsoft.AAD/domainServices resource.
func NewActiveDirectoryDomainService() *ActiveDirectoryDomainService {
	return &ActiveDirectoryDomainService{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.AAD/domainServices",
			ApiVersions:  []string{"2021-05-01"},
			SoftDelete:   false,
			ForceNew: []azwise.ForceNewRule{
				// name/location/resource_group_name are envelope; the ARM body ForceNew
				// fields are domainName and domainConfigurationType.
				{PropertyPath: "properties.domainName"},
				{PropertyPath: "properties.domainConfigurationType"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 3 * time.Hour,
				Read:   5 * time.Minute,
				Update: 2 * time.Hour,
				Delete: 1 * time.Hour,
			},
			StringRules: []azwise.StringRule{
				{
					// domain_name (validate.DomainServiceName). FQDN whose first label is
					// <=15 chars; maps to properties.domainName.
					PropertyPath: "properties.domainName",
					Regex:        `^(([0-9a-zA-Z])|(([0-9a-zA-Z][0-9a-zA-Z-]{0,28}[0-9a-zA-Z])))(\.[0-9a-zA-Z-]+)+$`,
					Message:      "domain_name must be a valid FQDN and the first element must be 15 or fewer characters",
				},
				{
					// sku is a free-form *string in the SDK (no enum constant); AzureRM
					// restricts it to these three tiers.
					PropertyPath:  "properties.sku",
					AllowedValues: []string{"Standard", "Enterprise", "Premium"},
					Message:       "must be one of Standard, Enterprise or Premium",
				},
				{
					// domainConfigurationType is a free-form *string in the SDK; AzureRM
					// restricts it to these values.
					PropertyPath:  "properties.domainConfigurationType",
					AllowedValues: []string{"FullySynced", "ResourceTrusting"},
					Message:       "must be one of FullySynced or ResourceTrusting",
				},
				{
					PropertyPath:  "properties.filteredSync",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.domainSecuritySettings.kerberosArmoring",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.domainSecuritySettings.kerberosRc4Encryption",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.domainSecuritySettings.ntlmV1",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.domainSecuritySettings.syncKerberosPasswords",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.domainSecuritySettings.syncNtlmPasswords",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.domainSecuritySettings.syncOnPremPasswords",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.domainSecuritySettings.tlsV1",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.notificationSettings.notifyDcAdmins",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.notificationSettings.notifyGlobalAdmins",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.ldapsSettings.ldaps",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
				{
					PropertyPath:  "properties.ldapsSettings.externalAccess",
					AllowedValues: []string{"Enabled", "Disabled"},
					Message:       "must be one of Enabled or Disabled",
				},
			},
			SensitiveFields: []string{
				"properties.ldapsSettings.pfxCertificate",
				"properties.ldapsSettings.pfxCertificatePassword",
			},
			// Read-only ARM response fields (never set on create/update).
			ComputedFields: []string{
				"properties.deploymentId",
				"properties.provisioningState",
				"properties.syncOwner",
				"properties.tenantId",
				"properties.version",
			},
			DefaultValues: []azwise.DefaultValue{
				// filtered_sync_enabled Default:false -> properties.filteredSync Disabled.
				{PropertyPath: "properties.filteredSync", Value: "Disabled"},
			},
			// sku and domainName are Required in AzureRM; the ARM API also needs at least
			// one replica set on create (properties.replicaSets, an array).
			RequiredFields: []string{
				"properties.domainName",
				"properties.sku",
				"properties.replicaSets",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewActiveDirectoryDomainService()) }
