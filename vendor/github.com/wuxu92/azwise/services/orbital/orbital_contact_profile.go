// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package orbital

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ContactProfile provides resource knowledge for Microsoft.Orbital/contactProfiles.
//
// Mirrors azurerm_orbital_contact_profile. network_configuration_subnet_id maps
// onto properties.networkConfiguration.subnetId; auto_tracking maps onto
// properties.autoTrackingConfiguration.
//
// Sources:
//   - terraform-provider-azurerm internal/services/orbital/orbital_contact_profile_resource.go
//     Arguments() lines 45-96, Create() lines 110-176 (ContactProfilesProperties:
//     AutoTrackingConfiguration, EventHubUri, Links, MinimumElevationDegrees,
//     MinimumViableContactDuration, NetworkConfiguration.SubnetId)
//   - Create/Read/Update/Delete timeouts: 30m / 5m / 30m / 30m
//   - go-azure-sdk resource-manager/orbital/2022-11-01/contactprofile
//     model_contactprofilesproperties.go /
//     model_contactprofilespropertiesnetworkconfiguration.go (subnetId) /
//     constants.go (AutoTrackingConfiguration: disabled, sBand, xBand)
//   - id_contactprofile.go: /providers/Microsoft.Orbital/contactProfiles/{contactProfileName}
type ContactProfile struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ContactProfile)(nil)

// NewContactProfile returns knowledge for the Orbital contactProfiles resource.
func NewContactProfile() *ContactProfile {
	return &ContactProfile{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Orbital/contactProfiles",
			ApiVersions:  []string{"2022-11-01"},
			ForceNew: []azwise.ForceNewRule{
				// commonschema.Location() is ForceNew; name/resource_group are envelope.
				{PropertyPath: "location"},
				// links (ContactProfileLinkSchema) is ForceNew.
				{PropertyPath: "properties.links"},
				// network_configuration_subnet_id is ForceNew.
				{PropertyPath: "properties.networkConfiguration.subnetId"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				// auto_tracking, minimum_variable_contact_duration,
				// network_configuration_subnet_id, links are Required in TF.
				"properties.autoTrackingConfiguration",
				"properties.minimumViableContactDuration",
				"properties.networkConfiguration.subnetId",
				"properties.links",
			},
			StringRules: []azwise.StringRule{
				// name: validation.StringIsNotEmpty.
				{
					PropertyPath: "",
					MinLength:    1,
					Message:      "contact profile name must not be empty",
				},
				// auto_tracking enum. Full SDK AutoTrackingConfiguration set.
				{
					PropertyPath:  "properties.autoTrackingConfiguration",
					AllowedValues: []string{"disabled", "sBand", "xBand"},
					Message:       "auto_tracking must be one of disabled, sBand, or xBand",
				},
				// minimum_variable_contact_duration: validation.StringIsNotEmpty (ISO8601 duration).
				{
					PropertyPath: "properties.minimumViableContactDuration",
					MinLength:    1,
					Message:      "minimum_variable_contact_duration must not be empty",
				},
				// event_hub_uri: validation.StringIsNotEmpty (Optional).
				{
					PropertyPath: "properties.eventHubUri",
					MinLength:    1,
					Message:      "event_hub_uri must not be empty when set",
				},
			},
			// network_configuration_subnet_id uses commonids.ValidateSubnetID — a
			// semantic subnet resource-ID check on properties.networkConfiguration.subnetId,
			// best expressed as an AzureResourceID customizer validator.
			// links[*].channels[*] endpoint/protocol/direction/polarization enums and
			// bandwidth/center_frequency floats are all array-element paths and are not
			// expressible as scalar rules here.
		},
	}
}

func init() { azwise.Register(NewContactProfile()) }
