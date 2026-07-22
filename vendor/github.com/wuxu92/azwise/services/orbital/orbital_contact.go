// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package orbital

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Contact provides resource knowledge for Microsoft.Orbital/spacecrafts/contacts.
//
// Mirrors azurerm_orbital_contact. The contact is a child of a spacecraft
// (spacecraft_id is the parent scope). Every argument is ForceNew and Required.
//
// Sources:
//   - terraform-provider-azurerm internal/services/orbital/orbital_contact_resource.go
//     Arguments() lines 39-83, Create() lines 97-148 (ContactsProperties:
//     ContactProfile{Id}, GroundStationName, ReservationEndTime, ReservationStartTime)
//   - Create/Read/Delete timeouts: 30m / 5m / 30m (no Update)
//   - go-azure-sdk resource-manager/orbital/2022-11-01/contact
//     model_contactsproperties.go / model_resourcereference.go
//   - id_contact.go: /providers/Microsoft.Orbital/spacecrafts/{spacecraftName}/contacts/{contactName}
type Contact struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Contact)(nil)

// NewContact returns knowledge for the Orbital contacts resource.
func NewContact() *Contact {
	return &Contact{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Orbital/spacecrafts/contacts",
			ApiVersions:  []string{"2022-11-01"},
			ForceNew: []azwise.ForceNewRule{
				// name and spacecraft_id (parent scope) are envelope-owned; the
				// remaining body fields are all ForceNew.
				{PropertyPath: "properties.contactProfile.id"},
				{PropertyPath: "properties.groundStationName"},
				{PropertyPath: "properties.reservationStartTime"},
				{PropertyPath: "properties.reservationEndTime"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.contactProfile.id",
				"properties.groundStationName",
				"properties.reservationStartTime",
				"properties.reservationEndTime",
			},
			StringRules: []azwise.StringRule{
				// name: validation.StringIsNotEmpty.
				{
					PropertyPath: "",
					MinLength:    1,
					Message:      "contact name must not be empty",
				},
				// ground_station_name: validation.StringIsNotEmpty.
				{
					PropertyPath: "properties.groundStationName",
					MinLength:    1,
					Message:      "ground_station_name must not be empty",
				},
				// reservation_start_time / reservation_end_time: validation.StringIsNotEmpty.
				{
					PropertyPath: "properties.reservationStartTime",
					MinLength:    1,
					Message:      "reservation_start_time must not be empty",
				},
				{
					PropertyPath: "properties.reservationEndTime",
					MinLength:    1,
					Message:      "reservation_end_time must not be empty",
				},
			},
			// contact_profile_id uses contactprofile.ValidateContactProfileID — a
			// semantic Microsoft.Orbital/contactProfiles resource-ID check that maps
			// onto properties.contactProfile.id. Best expressed as an AzureResourceID
			// customizer validator rather than a declarative rule.
		},
	}
}

func init() { azwise.Register(NewContact()) }
