package extendedlocation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// CustomLocation provides resource knowledge for Microsoft.ExtendedLocation/customLocations.
//
// Merges two AzureRM TF resources that map to the same ARM type:
//   - azurerm_extended_location_custom_location (current)
//   - azurerm_extended_custom_location (deprecated, superseded by the above; removed in 5.0)
//
// Both declare an identical schema, so their knowledge is unioned into this single file.
//
// Sources:
//   - terraform-provider-azurerm internal/services/extendedlocation/
//     extended_location_custom_location_resource.go (schema lines 50-125: name/namespace/
//     host_resource_id/host_type ForceNew; name+namespace StringMatch regexes; host_type
//     StringInSlice; cluster_extension_ids/host_resource_id/namespace Required), Create (135-191),
//     timeouts 30m/5m/30m/30m.
//   - extended_custom_location_resource.go (deprecated duplicate, identical schema lines 39-113).
//   - go-azure-sdk resource-manager/extendedlocation/2021-08-15/customlocations
//     CustomLocationProperties (clusterExtensionIds/displayName/hostResourceId/hostType/namespace/
//     authentication settable; provisioningState read-only), constants.go
//     PossibleValuesForHostType (Kubernetes).
type CustomLocation struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*CustomLocation)(nil)

// NewCustomLocation returns knowledge for the customLocations resource.
func NewCustomLocation() *CustomLocation {
	return &CustomLocation{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.ExtendedLocation/customLocations",
			ApiVersions:  []string{"2021-08-15"},
			SoftDelete:   false,
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "name"},
				{PropertyPath: "properties.namespace"},
				{PropertyPath: "properties.hostResourceId"},
				{PropertyPath: "properties.hostType"},
			},
			RequiredFields: []string{
				"properties.namespace",
				"properties.hostResourceId",
				"properties.clusterExtensionIds",
			},
			StringRules: []azwise.StringRule{
				{
					// Resource name (empty PropertyPath).
					Regex:   `^[A-Za-z\d.\-_]*[A-Za-z\d]$`,
					Message: "supported alphanumeric characters and periods, underscores, hyphens; must end with an alphanumeric character",
				},
				{
					PropertyPath: "properties.namespace",
					Regex:        `^[a-zA-Z0-9][a-zA-Z0-9-.]{0,252}$`,
					MinLength:    1,
					MaxLength:    253,
					Message:      "namespace must be 1-253 characters, may contain only letters, numbers, periods, hyphens, and must begin with a letter or number",
				},
				{
					// hostType enum (SDK constants.go).
					PropertyPath:  "properties.hostType",
					AllowedValues: []string{"Kubernetes"},
				},
			},
			// Server-populated, read-only ARM property.
			ComputedFields: []string{
				"properties.provisioningState",
			},
		},
	}
}

func init() { azwise.Register(NewCustomLocation()) }
