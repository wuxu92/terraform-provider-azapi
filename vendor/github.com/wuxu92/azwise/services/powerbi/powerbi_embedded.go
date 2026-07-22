// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package powerbi

import (
	"time"

	"github.com/wuxu92/azwise"
)

// Embedded provides resource knowledge for Microsoft.PowerBIDedicated/capacities.
//
// Mirrors azurerm_powerbi_embedded. sku_name maps onto sku.name; mode onto
// properties.mode; administrators (a set) onto properties.administration.members.
//
// Sources:
//   - terraform-provider-azurerm internal/services/powerbi/powerbi_embedded_resource.go
//     Schema() lines 46-94, Create() lines 104-147 (DedicatedCapacity: Sku.Name,
//     Properties.Administration.Members, Properties.Mode)
//   - Create/Read/Update/Delete timeouts: 30m / 5m / 30m / 30m
//   - go-azure-sdk resource-manager/powerbidedicated/2021-01-01/capacities
//     model_dedicatedcapacityproperties.go / model_capacitysku.go /
//     constants.go (Mode: Gen1, Gen2)
//   - internal/services/powerbi/validate/embedded_name.go (name regex)
//   - id_capacity.go: /providers/Microsoft.PowerBIDedicated/capacities/{capacityName}
type Embedded struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*Embedded)(nil)

// NewEmbedded returns knowledge for the PowerBI Dedicated capacities resource.
func NewEmbedded() *Embedded {
	return &Embedded{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.PowerBIDedicated/capacities",
			ApiVersions:  []string{"2021-01-01"},
			ForceNew: []azwise.ForceNewRule{
				// commonschema.Location() is ForceNew; name/resource_group are envelope.
				{PropertyPath: "location"},
				// mode is ForceNew.
				{PropertyPath: "properties.mode"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				// sku_name and administrators are Required in TF.
				"sku.name",
				"properties.administration.members",
			},
			DefaultValues: []azwise.DefaultValue{
				// mode Default (5.0) is GenTwo; pre-5.0 default is GenOne.
				{PropertyPath: "properties.mode", Value: "Gen2"},
			},
			StringRules: []azwise.StringRule{
				// name: validate.EmbeddedName (4-64 chars, lowercase letter first,
				// then lowercase letters or digits).
				{
					PropertyPath: "",
					Regex:        `^[a-z][a-z0-9]{3,63}$`,
					Message:      "name must be 4-64 characters, start with a lowercase letter, and contain only lowercase letters or numbers",
				},
				// sku_name enum A1-A8.
				{
					PropertyPath:  "sku.name",
					AllowedValues: []string{"A1", "A2", "A3", "A4", "A5", "A6", "A7", "A8"},
					Message:       "sku_name must be one of A1 through A8",
				},
				// mode enum. Full SDK Mode set.
				{
					PropertyPath:  "properties.mode",
					AllowedValues: []string{"Gen1", "Gen2"},
					Message:       "mode must be one of Gen1 or Gen2",
				},
			},
			// administrators[*] uses validate.EmbeddedAdministratorName (email or UUID)
			// on properties.administration.members[*] — an array-element check that is
			// not expressible as a scalar StringRule here.
		},
	}
}

func init() { azwise.Register(NewEmbedded()) }
