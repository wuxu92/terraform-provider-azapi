package customproviders

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CustomProvider provides resource knowledge for
// Microsoft.CustomProviders/resourceProviders.
//
// Mirrors azurerm_custom_provider (aka azurerm_custom_resource_provider).
// name/location/resource_group_name are envelope fields. tags is ForceNew
// (commonschema.TagsForceNew) and maps to the ARM envelope "tags".
//
// Notes:
//   - At least one of resource_type / action must be set: modelled as an
//     AtLeastOneOf relational rule over the ARM array paths
//     properties.resourceTypes / properties.actions.
//   - routing_type (enum Proxy | "Proxy,Cache") lives on each resourceTypes
//     array element (properties.resourceTypes[*].routingType); array-element
//     paths are not expressible as declarative rules, so it is documented here
//     but not emitted. Its AzureRM default is "Proxy".
//   - resource_type.name/action.name (NoZeroValues), *.endpoint / validation
//     specification (IsURLWithHTTPS) are all array-element string fields — not
//     emitted for the same reason.
//
// Sources:
//   - terraform-provider-azurerm internal/services/customproviders/custom_provider_resource.go
//     (schema L47-123: name ForceNew + CustomProviderName; location/RG/tags
//     ForceNew; resource_type/action AtLeastOneOf; create L127-177; timeouts
//     30m/5m/30m/30m)
//   - internal/services/customproviders/validate/custom_provider_name.go
//     (regex ^[a-zA-Z0-9_]+$, length 3-63)
//   - go-azure-sdk resource-manager/customproviders/2018-09-01-preview/customresourceprovider
//     CustomRPManifestProperties{actions,resourceTypes,validations,provisioningState};
//     ResourceTypeRouting = Proxy | "Proxy,Cache"; ProvisioningState read-only
type CustomProvider struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CustomProvider)(nil)

// NewCustomProvider returns knowledge for the resourceProviders resource.
func NewCustomProvider() *CustomProvider {
	return &CustomProvider{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.CustomProviders/resourceProviders",
			ApiVersions:  []string{"2018-09-01-preview"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				// tags uses commonschema.TagsForceNew.
				{PropertyPath: "tags"},
			},
			StringRules: []azwise.StringRule{
				// name (CustomProviderName) validates the ARM resource name.
				{Regex: `^[a-zA-Z0-9_]+$`, MinLength: 3, MaxLength: 63, Message: "name may only contain letters, digits and underscores and be 3-63 characters long"},
			},
			AtLeastOneOf: []azwise.RelationalRule{
				{Paths: []string{"properties.resourceTypes", "properties.actions"}, Message: "at least one of resource_type or action must be specified"},
			},
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

// Self-registers into the azwise registry (no central register.go).
func init() { azwise.Register(NewCustomProvider()) }
