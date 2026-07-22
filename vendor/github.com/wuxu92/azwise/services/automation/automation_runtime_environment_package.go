package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationRuntimeEnvironmentPackage provides resource knowledge for
// Microsoft.Automation/automationAccounts/runtimeEnvironments/packages.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_runtime_environment_package_resource.go
//     schema (38-106), Create (120-183); timeouts 30m/5m/-/30m.
//   - go-azure-sdk resource-manager/automation/2024-10-23/packageresource
//     PackageCreateOrUpdateProperties.contentLink (contentHash{algorithm,value}/uri/version).
//     PackageProperties.default/sizeInBytes/version/provisioningState are read-only.
//   - validators: StringIsNotEmpty (name, hash_algorithm, hash_value),
//     IsURLWithHTTPS (content_uri), StringMatch 2-4 numeric segments (content_version),
//     RequiredWith (hash_algorithm <-> hash_value).
//
// name / automation_runtime_environment_id are the envelope + parent id (ForceNew
// by construction). default/size_in_bytes/version are Computed attributes.
type AutomationRuntimeEnvironmentPackage struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationRuntimeEnvironmentPackage)(nil)

func NewAutomationRuntimeEnvironmentPackage() *AutomationRuntimeEnvironmentPackage {
	return &AutomationRuntimeEnvironmentPackage{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/runtimeEnvironments/packages",
			ApiVersions:  []string{"2024-10-23"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 30 * time.Minute,
				Delete: 30 * time.Minute,
			},
			// content_uri/version/hash are all ForceNew and live in the body.
			ForceNew: []azwise.ForceNewRule{
				{PropertyPath: "properties.contentLink.uri"},
				{PropertyPath: "properties.contentLink.version"},
				{PropertyPath: "properties.contentLink.contentHash.algorithm"},
				{PropertyPath: "properties.contentLink.contentHash.value"},
			},
			// content_uri is Required.
			RequiredFields: []string{
				"properties.contentLink.uri",
			},
			// default/sizeInBytes/version/provisioningState are read-only in the GET
			// model and absent from PackageCreateOrUpdateProperties.
			ComputedFields: []string{
				"properties.default",
				"properties.sizeInBytes",
				"properties.version",
				"properties.provisioningState",
			},
			StringRules: []azwise.StringRule{
				// name — StringIsNotEmpty.
				{MinLength: 1, Message: "name must not be empty"},
				// content_uri → properties.contentLink.uri — IsURLWithHTTPS.
				{
					PropertyPath: "properties.contentLink.uri",
					Regex:        `^https://`,
					Message:      "content_uri must be a valid https URL",
				},
				// content_version → properties.contentLink.version — StringMatch (2-4 segments).
				{
					PropertyPath: "properties.contentLink.version",
					Regex:        `^[0-9]+\.[0-9]+(\.[0-9]+){0,2}$`,
					Message:      "content_version must have 2 to 4 numeric segments (e.g. 1.0, 1.0.0, or 1.0.0.0)",
				},
				// hash_algorithm → properties.contentLink.contentHash.algorithm — StringIsNotEmpty.
				{PropertyPath: "properties.contentLink.contentHash.algorithm", MinLength: 1, Message: "hash_algorithm must not be empty"},
				// hash_value → properties.contentLink.contentHash.value — StringIsNotEmpty.
				{PropertyPath: "properties.contentLink.contentHash.value", MinLength: 1, Message: "hash_value must not be empty"},
			},
			// hash_algorithm and hash_value are RequiredWith each other.
			RequiredWith: []azwise.RelationalRule{
				{
					Paths:   []string{"properties.contentLink.contentHash.algorithm", "properties.contentLink.contentHash.value"},
					Message: "hash_algorithm requires hash_value",
				},
				{
					Paths:   []string{"properties.contentLink.contentHash.value", "properties.contentLink.contentHash.algorithm"},
					Message: "hash_value requires hash_algorithm",
				},
			},
		},
	}
}

func init() { azwise.Register(NewAutomationRuntimeEnvironmentPackage()) }
