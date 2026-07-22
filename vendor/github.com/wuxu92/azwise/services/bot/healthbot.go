package bot

import (
	"time"

	"github.com/wuxu92/azwise"
)

// HealthBot provides resource knowledge for Microsoft.HealthBot/healthBots.
//
// Mirrors azurerm_healthbot. name/location/resource_group_name are operational-envelope
// fields; name is ForceNew in AzureRM but is the ARM resource name (not a body property),
// so it is validated (2-64 chars, ^[a-zA-Z0-9][a-zA-Z0-9_.-]*$) but not encoded as a body
// ForceNew rule.
//
// sku_name (-> sku.name) is CONDITIONALLY ForceNew: AzureRM's CustomizeDiff forces
// replacement only when the SKU is being DOWNGRADED to "F0" (the API rejects downgrades).
// Any other SKU change is an in-place update. This cannot be expressed as a declarative
// ForceNew rule (which fires on every change), so CheckForceNew is overridden below to
// reproduce the exact condition.
//
// Sources:
//   - terraform-provider-azurerm internal/services/bot/healthbot_resource.go
//     (Schema L45-69, CustomizeDiff L71-82 F0-downgrade ForceNew, Create body L106-112,
//     Read L146-156; timeouts 30m/5m/30m/30m)
//   - terraform-provider-azurerm internal/services/bot/validate/bot_healthbot_name.go
//     (name length 2-64, regex ^[a-zA-Z0-9][a-zA-Z0-9_.-]*$)
//   - go-azure-sdk resource-manager/healthbot/2025-05-25/healthbots
//     HealthBot (model_healthbot.go), Sku (model_sku.go name), SkuName enum
//     (constants.go C0/C1/F0/PES/S1). HealthBotProperties (accessControlMethod,
//     botManagementPortalLink, keyVaultProperties, provisioningState) is entirely
//     read-only — absent from the create body.
type HealthBot struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*HealthBot)(nil)

// NewHealthBot returns knowledge for the healthBots resource.
func NewHealthBot() *HealthBot {
	return &HealthBot{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.HealthBot/healthBots",
			ApiVersions:  []string{"2025-05-25"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			StringRules: []azwise.StringRule{
				{
					PropertyPath: "",
					Regex:        `^[a-zA-Z0-9][a-zA-Z0-9_.-]*$`,
					MinLength:    2,
					MaxLength:    64,
					Message:      "name must be 2-64 characters and start with a letter or digit, containing only alphanumerics, underscores, dots and dashes",
				},
				{
					// Full ARM SDK SkuName set (AzureRM allows all via PossibleValuesForSkuName).
					PropertyPath:  "sku.name",
					AllowedValues: []string{"C0", "C1", "F0", "PES", "S1"},
					Message:       "sku_name must be one of C0, C1, F0, PES or S1",
				},
			},
			RequiredFields: []string{
				"sku.name",
			},
			// HealthBotProperties is read-only in the SDK — none of its fields are settable
			// on create. botManagementPortalLink is the only one AzureRM surfaces.
			ComputedFields: []string{
				"properties.botManagementPortalLink",
				"properties.accessControlMethod",
				"properties.provisioningState",
			},
		},
	}
}

// CheckForceNew extends the declarative check with HealthBot's conditional rule:
// downgrading the SKU to F0 forces replacement (mirrors the AzureRM CustomizeDiff).
func (h *HealthBot) CheckForceNew(oldBody, newBody map[string]interface{}) bool {
	if h.BaseKnowledge.CheckForceNew(oldBody, newBody) {
		return true
	}
	oldSku := azwise.ExtractStringValue(oldBody, "sku.name")
	newSku := azwise.ExtractStringValue(newBody, "sku.name")
	// Only a change that lands on F0 (a downgrade) requires replacement.
	if oldSku != "" && newSku == "F0" && oldSku != newSku {
		return true
	}
	return false
}

func init() { azwise.Register(NewHealthBot()) }
