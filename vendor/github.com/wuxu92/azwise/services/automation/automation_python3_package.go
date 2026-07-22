package automation

import (
	"time"

	"github.com/wuxu92/azwise"
)

// AutomationPython3Package provides resource knowledge for
// Microsoft.Automation/automationAccounts/python3Packages.
//
// Sources:
//   - terraform-provider-azurerm internal/services/automation/automation_python3_package_resource.go
//     schema (35-85), Create (99-150); timeouts 30m/5m/10m/10m.
//   - go-azure-sdk resource-manager/automation/2024-10-23/python3package
//     PythonPackageCreateProperties.contentLink (contentHash{algorithm,value}/uri/version).
//   - validators: StringIsNotEmpty (name, content_version, hash_algorithm, hash_value),
//     IsURLWithHTTPorHTTPS (content_uri), RequiredWith (hash_algorithm <-> hash_value).
//
// name / automation_account_name are the envelope + parent id (ForceNew by
// construction). All content_link fields are ForceNew in AzureRM.
type AutomationPython3Package struct {
	azwise.BaseKnowledge
}

var _ azwise.ResourceKnowledge = (*AutomationPython3Package)(nil)

func NewAutomationPython3Package() *AutomationPython3Package {
	return &AutomationPython3Package{
		BaseKnowledge: azwise.BaseKnowledge{
			ResourceType: "Microsoft.Automation/automationAccounts/python3Packages",
			ApiVersions:  []string{"2024-10-23"},
			TimeoutsConfig: &azwise.Timeouts{
				Create: 30 * time.Minute,
				Read:   5 * time.Minute,
				Update: 10 * time.Minute,
				Delete: 10 * time.Minute,
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
			StringRules: []azwise.StringRule{
				// name — StringIsNotEmpty.
				{MinLength: 1, Message: "name must not be empty"},
				// content_uri → properties.contentLink.uri — IsURLWithHTTPorHTTPS.
				{
					PropertyPath: "properties.contentLink.uri",
					Regex:        `^https?://`,
					Message:      "content_uri must be a valid URL with http or https scheme",
				},
				// content_version → properties.contentLink.version — StringIsNotEmpty.
				{PropertyPath: "properties.contentLink.version", MinLength: 1, Message: "content_version must not be empty"},
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

func init() { azwise.Register(NewAutomationPython3Package()) }
