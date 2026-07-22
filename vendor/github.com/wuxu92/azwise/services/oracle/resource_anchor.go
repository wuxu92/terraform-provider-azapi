// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oracle

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ResourceAnchor provides resource knowledge for Oracle.Database/resourceAnchors.
//
// Contributing TF resource:
//   - azurerm_oracle_resource_anchor (oracle_resource_anchor_resource.go)
//
// The resource body carries no settable properties beyond envelope fields (name is
// ForceNew; location is hardcoded "global"; linked_compartment_id is computed). tags
// are the only mutable input.
//
// Sources:
//   - AzureRM oracle_resource_anchor_resource.go (schema + Create mapping)
//   - AzureRM validate/oracle_resource_anchor.go (name rule)
//   - Azure SDK oracledatabase/2025-09-01/resourceanchors
type ResourceAnchor struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ResourceAnchor)(nil)

// NewResourceAnchor returns a ResourceAnchor knowledge instance.
func NewResourceAnchor() *ResourceAnchor {
	return &ResourceAnchor{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Oracle.Database/resourceAnchors",
			ApiVersions:  []string{"2025-09-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 10 * time.Minute,
				Read:   5 * time.Minute,
				Update: 10 * time.Minute,
				Delete: 10 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				// ── Resource name (validate.ResourceAnchorName) ──
				// letters, numbers and hyphens only, 24 characters max.
				{
					Regex:     `^[a-zA-Z0-9-]*$`,
					MaxLength: 24,
					Message:   "must contain only letters, numbers and hyphens, 24 characters max",
				},
			},
		},
	}
}

func init() { azwise.Register(NewResourceAnchor()) }
