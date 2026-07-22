package cognitive

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CognitiveDeployment provides resource knowledge for
// Microsoft.CognitiveServices/accounts/deployments.
//
// Contributing TF resource:
//   - azurerm_cognitive_deployment — cognitive_deployment_resource.go
//
// Sources:
//   - AzureRM cognitive_deployment_resource.go schema + expand functions
//   - Azure SDK cognitive/2026-03-01/deployments models + constants.go
type CognitiveDeployment struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CognitiveDeployment)(nil)

// NewCognitiveDeployment returns a CognitiveDeployment knowledge instance.
func NewCognitiveDeployment() *CognitiveDeployment {
	return &CognitiveDeployment{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.CognitiveServices/accounts/deployments",
			ApiVersions:  []string{"2026-03-01"},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				// The whole `model` block is ForceNew; its name and format are ForceNew too.
				{PropertyPath: "properties.model.name"},
				{PropertyPath: "properties.model.format"},
				// sku.name/tier/size/family are ForceNew (capacity is updatable).
				{PropertyPath: "sku.name"},
				{PropertyPath: "sku.tier"},
				{PropertyPath: "sku.size"},
				{PropertyPath: "sku.family"},
			},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			RequiredFields: []string{
				"properties.model",
				"properties.model.name",
				"properties.model.format",
				"sku.name",
			},
			StringRules: []azwise.StringRule{
				// ── sku.name ──
				// validation.StringInSlice; Sku.Name is a free-form string in the SDK.
				{
					PropertyPath: "sku.name",
					AllowedValues: []string{
						"Standard", "DataZoneBatch", "DataZoneProvisionedManaged",
						"DataZoneStandard", "GlobalBatch", "GlobalProvisionedManaged",
						"GlobalStandard", "ProvisionedManaged",
					},
					Message: "must be a valid deployment SKU name",
				},
				// ── sku.tier ──
				// PossibleValuesForSkuTier().
				{
					PropertyPath:  "sku.tier",
					AllowedValues: []string{"Basic", "Enterprise", "Free", "Premium", "Standard"},
					Message:       "must be a valid SKU tier",
				},
				// ── properties.versionUpgradeOption ──
				// PossibleValuesForDeploymentModelVersionUpgradeOption().
				{
					PropertyPath: "properties.versionUpgradeOption",
					AllowedValues: []string{
						"NoAutoUpgrade", "OnceCurrentVersionExpired", "OnceNewDefaultVersionAvailable",
					},
					Message: "must be a valid version upgrade option",
				},
			},
			IntRules: []azwise.IntRule{
				// ── sku.capacity ──
				// validation.IntAtLeast(1), default 1.
				{
					PropertyPath: "sku.capacity",
					MinValue:     azwise.Ptr(int64(1)),
					Message:      "capacity must be at least 1",
				},
			},
			DefaultValues: []azwise.DefaultValue{
				{PropertyPath: "sku.capacity", Value: int64(1)},
				{PropertyPath: "properties.versionUpgradeOption", Value: "OnceNewDefaultVersionAvailable"},
				{PropertyPath: "properties.raiPolicyName"}, // O+C, server assigns default policy when omitted
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewCognitiveDeployment()) }
