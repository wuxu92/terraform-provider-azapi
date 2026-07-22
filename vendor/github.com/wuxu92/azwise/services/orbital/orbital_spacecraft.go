// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package orbital

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Spacecraft provides resource knowledge for Microsoft.Orbital/spacecrafts.
//
// Mirrors azurerm_orbital_spacecraft. two_line_elements is a 2-element TF list
// that expands one-to-many onto properties.tleLine1 / properties.tleLine2.
//
// Sources:
//   - terraform-provider-azurerm internal/services/orbital/orbital_spacecraft_resource.go
//     Arguments() lines 42-83, Create() lines 97-144 (SpacecraftsProperties: Links,
//     NoradId, TleLine1, TleLine2, TitleLine)
//   - Create/Read/Update/Delete timeouts: 30m / 5m / 30m / 30m
//   - go-azure-sdk resource-manager/orbital/2022-11-01/spacecraft
//     model_spacecraftsproperties.go (links, noradId, titleLine, tleLine1, tleLine2)
//   - id_spacecraft.go: /providers/Microsoft.Orbital/spacecrafts/{spacecraftName}
type Spacecraft struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Spacecraft)(nil)

// NewSpacecraft returns knowledge for the Orbital spacecrafts resource.
func NewSpacecraft() *Spacecraft {
	return &Spacecraft{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Orbital/spacecrafts",
			ApiVersions:  []string{"2022-11-01"},
			ForceNew: []azwise.ForceNewRule{
				// commonschema.Location() is ForceNew; name/resource_group are envelope.
				{PropertyPath: "location"},
				// links (SpacecraftLinkSchema) is ForceNew.
				{PropertyPath: "properties.links"},
				// two_line_elements is ForceNew and expands to tleLine1/tleLine2.
				{PropertyPath: "properties.tleLine1"},
				{PropertyPath: "properties.tleLine2"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				// norad_id, title_line, two_line_elements, links are Required in TF.
				"properties.noradId",
				"properties.titleLine",
				"properties.tleLine1",
				"properties.tleLine2",
				"properties.links",
			},
			StringRules: []azwise.StringRule{
				// name: validation.StringIsNotEmpty.
				{
					PropertyPath: "",
					MinLength:    1,
					Message:      "spacecraft name must not be empty",
				},
				// norad_id: validation.StringLenBetween(5, 5).
				{
					PropertyPath: "properties.noradId",
					MinLength:    5,
					MaxLength:    5,
					Message:      "norad_id must be exactly 5 characters",
				},
				// title_line: validation.StringIsNotEmpty.
				{
					PropertyPath: "properties.titleLine",
					MinLength:    1,
					Message:      "title_line must not be empty",
				},
				// two_line_elements element: validation.StringLenBetween(69, 69).
				{
					PropertyPath: "properties.tleLine1",
					MinLength:    69,
					MaxLength:    69,
					Message:      "each two-line element must be exactly 69 characters",
				},
				{
					PropertyPath: "properties.tleLine2",
					MinLength:    69,
					MaxLength:    69,
					Message:      "each two-line element must be exactly 69 characters",
				},
			},
			// links[*].direction / links[*].polarization enums live on array elements
			// (properties.links[*].*) and are not expressible as scalar StringRules; the
			// bandwidth/center_frequency floats are likewise array-element paths.
		},
	}
}

func init() { azwise.Register(NewSpacecraft()) }
