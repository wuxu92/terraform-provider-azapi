package chaosstudio

import (
	"time"

	"github.com/wuxu92/azwise"
)

// ChaosStudioCapability provides resource knowledge for
// Microsoft.Chaos/targets/capabilities.
//
// Contributing Terraform resource: azurerm_chaos_studio_capability.
//
// Parent path verification: the capability ID is
// ".../providers/Microsoft.Chaos/targets/{targetName}/capabilities/{capabilityName}"
// (go-azure-helpers commonids/chaos_studio_capability.go ID() + Segments: staticTargets
// "targets" -> targetName -> staticCapabilities "capabilities" -> capabilityName), so the
// ARM parent is "targets" and the resource type is Microsoft.Chaos/targets/capabilities.
//
// Sources:
//   - terraform-provider-azurerm internal/services/chaosstudio/chaos_studio_capability_resource.go
//     (schema L43-66, Create body L132-141, timeouts Create/Read/Delete 30m/5m/30m,
//     no Update)
//   - go-azure-helpers resourcemanager/commonids/chaos_studio_capability.go L98-113
//     (parent = targets)
//   - go-azure-sdk resource-manager/chaosstudio/2023-11-01/capabilities:
//     model_capability.go, model_capabilityproperties.go (all properties read-only)
//
// Notes:
//   - capability_type (Required, ForceNew) is the capability NAME segment of the resource
//     ID, not a body property -> ForceNew on the resource name. AzureRM validates it at
//     runtime against the target type's supported capabilities (dynamic list), which is
//     not expressible as a static enum.
//   - chaos_studio_target_id (Required, ForceNew) is the parent resource reference
//     (commonids.ValidateChaosStudioTargetID, a semantic resource-ID validator) -> not a
//     body property; belongs in an azapin customizer, not a declarative StringRule.
//   - The Create body sends an EMPTY properties object ("The API only accepts requests
//     with an empty body for Properties"); every CapabilityProperties field is
//     server-populated and read-only -> ComputedFields. There are therefore no
//     RequiredFields or settable body properties.
type ChaosStudioCapability struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*ChaosStudioCapability)(nil)

// NewChaosStudioCapability returns knowledge for Microsoft.Chaos/targets/capabilities.
func NewChaosStudioCapability() *ChaosStudioCapability {
	return &ChaosStudioCapability{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Chaos/targets/capabilities",
			ApiVersions:  []string{"2023-11-01"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// capability_type is the resource name segment and is ForceNew.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
			},
			// The create body is empty; all CapabilityProperties fields are read-only.
			ComputedFields: []string{
				"properties.description",
				"properties.parametersSchema",
				"properties.publisher",
				"properties.targetType",
				"properties.urn",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewChaosStudioCapability()) }
