package frontdoor

import (
	"time"

	"github.com/wuxu92/azwise"
)

// FrontDoorRulesEngine provides resource knowledge for
// Microsoft.Network/frontDoors/rulesEngines (classic Front Door rules engine).
//
// Contributing Terraform resource: azurerm_frontdoor_rules_engine.
//
// Sources:
//   - terraform-provider-azurerm internal/services/frontdoor/frontdoor_rules_engine_resource.go
//     (resource L26-264, create L266-295, expand L340-365)
//   - go-azure-sdk resource-manager/frontdoor/2020-05-01/frontdoors:
//     model_rulesengine.go, model_rulesengineproperties.go, constants.go
//     (FrontDoorResourceState).
//
// Notes:
//   - name, frontdoor_name (parent) & resource_group_name are envelope-owned; all three
//     are ForceNew. location is computed. None map to a body property, so ForceNew is
//     empty here.
//   - the schema's "enabled" field is not written to the ARM body (RulesEngineProperties
//     has no enabled field), so no rule is emitted for it.
//   - rule -> properties.rules[*] (name, priority, match_condition, action) is an
//     array-element sub-tree; its per-element enums (match variable/operator/transform,
//     header action type) are not expressible as scalar declarative rules here.
type FrontDoorRulesEngine struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*FrontDoorRulesEngine)(nil)

func NewFrontDoorRulesEngine() *FrontDoorRulesEngine {
	return &FrontDoorRulesEngine{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Network/frontDoors/rulesEngines",
			ApiVersions:  []string{"2020-05-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 6 * time.Hour,
				Read:   5 * time.Minute,
				Update: 6 * time.Hour,
				Delete: 6 * time.Hour,
			},
			ComputedFields: []string{
				"properties.resourceState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewFrontDoorRulesEngine()) }
