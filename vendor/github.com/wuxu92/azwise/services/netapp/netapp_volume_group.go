package netapp

import (
	"time"

	"github.com/wuxu92/azwise"
)

// NetAppVolumeGroup provides resource knowledge for
// Microsoft.NetApp/netAppAccounts/volumeGroups.
//
// One ARM type shared by two typed Terraform resources — azurerm_netapp_volume_group_oracle
// and azurerm_netapp_volume_group_sap_hana — which differ only in the hardcoded
// groupMetaData.applicationType (ORACLE vs SAP-HANA) and per-kind volume counts. Only
// knowledge universal to both is unioned here; kind-specific volume shapes are left out.
//
// Sources:
//   - internal/services/netapp/netapp_volume_group_oracle_resource.go
//     (schema 45-321: name ForceNew+VolumeGroupName; group_description Required ForceNew;
//     application_identifier Required ForceNew regex ^[a-zA-Z][\w-]{2,11}$; volume list
//     MinItems 2 MaxItems 12; create 375-... → VolumeGroupDetails{GroupMetaData{...,
//     ApplicationType ORACLE}, Volumes}; timeouts Create 90m Update/Delete 120m Read 5m).
//   - internal/services/netapp/netapp_volume_group_sap_hana_resource.go
//     (schema 44-...: same shape; volume list MinItems 2 MaxItems 5;
//     create 388-... → ApplicationType SAP-HANA).
//   - internal/services/netapp/validate/volume_grupo_name.go (regex ^[a-zA-Z][-_\da-zA-Z]{0,63}$).
//   - go-azure-sdk resource-manager/netapp/2026-01-01/volumegroups:
//     model_volumegroupproperties.go (groupMetaData/volumes; provisioningState read-only),
//     model_volumegroupmetadata.go (applicationIdentifier/applicationType/groupDescription;
//     volumesCount read-only), constants.go (ApplicationType ORACLE/SAP-HANA),
//     id_volumegroup.go (type segment casing "volumeGroups").
type NetAppVolumeGroup struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*NetAppVolumeGroup)(nil)

// NewNetAppVolumeGroup returns knowledge for the volumeGroups resource.
func NewNetAppVolumeGroup() *NetAppVolumeGroup {
	return &NetAppVolumeGroup{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.NetApp/netAppAccounts/volumeGroups",
			ApiVersions:  []string{"2026-01-01"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 90 * time.Minute,
				Read:   5 * time.Minute,
				Update: 120 * time.Minute,
				Delete: 120 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.groupMetaData.groupDescription"},
				{PropertyPath: "properties.groupMetaData.applicationIdentifier"},
				{PropertyPath: "properties.groupMetaData.applicationType"},
			},
			// group_description and application_identifier are Required for both kinds.
			RequiredFields: []string{
				"properties.groupMetaData.groupDescription",
				"properties.groupMetaData.applicationIdentifier",
			},
			StringRules: []azwise.StringRule{
				{
					// AzureRM VolumeGroupName: 1-64 chars, start with a letter.
					Regex:     `^[a-zA-Z][-_\da-zA-Z]{0,63}$`,
					MinLength: 1,
					MaxLength: 64,
					Message:   "must be 1-64 characters, start with a letter, and contain only letters, numbers, underscores and hyphens",
				},
				{
					// application_identifier: 3-12 chars, alphanumerics/hyphens/underscores,
					// must begin with a letter (identical validator on both kinds).
					PropertyPath: "properties.groupMetaData.applicationIdentifier",
					Regex:        `^[a-zA-Z][\w-]{2,11}$`,
					Message:      "must be 3-12 characters, begin with a letter, and contain only letters, numbers, hyphens and underscores",
				},
				{
					// Both kinds are valid ARM applicationType values.
					PropertyPath:  "properties.groupMetaData.applicationType",
					AllowedValues: []string{"ORACLE", "SAP-HANA"},
				},
			},
			// Server-populated, read-only ARM properties.
			ComputedFields: []string{
				"properties.provisioningState",
				"properties.groupMetaData.volumesCount",
			},
			// NOTE: the volume list maps to properties.volumes[*] (array of full volume
			// objects). Its element-level validators (service_level/security_style/
			// network_features enums, storage_quota_in_gb ranges, volume_spec_name per-kind
			// allowed values, per-volume ForceNew fields) are array-element paths and cannot
			// be expressed as declarative rules here. The per-kind MinItems/MaxItems on the
			// volume list (Oracle 2-12, SAP-HANA 2-5) are also not unioned — they differ by
			// contributing resource.
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewNetAppVolumeGroup()) }
